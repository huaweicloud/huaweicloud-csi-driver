package main

import (
	"encoding/json"
	"fmt"
	"github.com/golang/glog"
	"github.com/huaweicloud/huaweicloud-csi-driver/pkg/obs"
	"github.com/huaweicloud/huaweicloud-csi-driver/pkg/utils"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/exec"
)

const (
	socketPath  = "/dev/csi-tool/connector.sock"
	bucketName  = "bucketName"
	targetPath  = "targetPath"
	region      = "region"
	cloud       = "cloud"
	credential  = "credential"
	defaultOpts = "-o big_writes -o max_write=131072 -o use_ino"
)

type ResponseBody struct {
	Data string
}

type MyHandler struct{}

func (myHandler *MyHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		generateResponseBody(writer, http.StatusBadRequest, err.Error())
		return
	}
	if len(body) == 0 {
		generateResponseBody(writer, http.StatusBadRequest, "Request body is empty")
		return
	}
	glog.Infof("ServerHTTP Request body: %s", string(body))
	var commandRPC obs.CommandRPC
	if err := json.Unmarshal(body, &commandRPC); err != nil {
		generateResponseBody(writer, http.StatusBadRequest, err.Error())
		return
	}
	if err := checkRequestToken(commandRPC.Token, commandRPC.Parameters); err != nil {
		generateResponseBody(writer, http.StatusBadRequest,
			fmt.Sprintf("Failed to check token, err: %v", err.Error()))
		return
	}

	if commandRPC.Action == obs.ActionMount {
		if err := checkMountParameters(commandRPC.Parameters); err != nil {
			generateResponseBody(writer, http.StatusBadRequest,
				fmt.Sprintf("Failed to check mount action parameters, err: %v", err.Error()))
			return
		}
		if err := mountHandler(commandRPC.Parameters); err != nil {
			generateResponseBody(writer, http.StatusInternalServerError, err.Error())
			return
		}
		generateResponseBody(writer, http.StatusOK, "success")
		return
	}
	generateResponseBody(writer, http.StatusBadRequest, fmt.Sprintf("Invalid action %s", commandRPC.Action))
}

func checkRequestToken(token string, parameters map[string]string) error {
	hashMd5, err := utils.Md5SortMap(parameters)
	if err != nil {
		return err
	}
	original, err := utils.DecryptAESCBC(obs.Secret, token)
	if err != nil {
		return err
	}
	if hashMd5 != original {
		return fmt.Errorf("invalid token")
	}
	return nil
}

func checkMountParameters(parameters map[string]string) error {
	keys := [5]string{bucketName, targetPath, region, cloud, credential}
	for _, k := range keys {
		if len(parameters[k]) == 0 {
			return fmt.Errorf("param %s cannot be empty", k)
		}
	}
	credentialValue := parameters[credential]
	if !checkFileExists(credentialValue) {
		return fmt.Errorf("credential file %s not exist", credential)
	}
	return nil
}

func generateResponseBody(writer http.ResponseWriter, statusCode int, data string) {
	responseBody := ResponseBody{
		Data: data,
	}
	response, _ := json.Marshal(responseBody)
	writer.WriteHeader(statusCode)
	writer.Header().Set("Content-Type", "application/json")
	size, err := writer.Write(response)
	if err != nil {
		glog.Errorf("Failed to write response, err: %v", err)
		return
	}
	glog.Infof("Write response size: %d", size)
}

func main() {
	if checkFileExists(socketPath) {
		if err := os.Remove(socketPath); err != nil {
			glog.Fatalf("Failed to remove path: %s, err: %v", socketPath, err)
		}
	}

	listen, err := net.Listen("unix", socketPath)
	if err != nil {
		glog.Fatalf("net.Listen failed. path: %s, err: %v", socketPath, err)
		return
	}
	glog.Infof("Success start listener")
	server := http.Server{Handler: &MyHandler{}}
	if err := server.Serve(listen); err != nil {
		glog.Fatalf("Socket Listener failed to close, err: %v", err)
	}
}

func checkFileExists(filename string) bool {
	_, err := os.Stat(filename)
	if err == nil {
		return true
	}
	return !os.IsNotExist(err)
}

func mountHandler(parameters map[string]string) error {
	mntCmd := fmt.Sprintf("obsfs %s %s -o url=obs.%s.%s -o passwd_file=%s %s",
		parameters[bucketName], parameters[targetPath],
		parameters[region], parameters[cloud], parameters[credential], defaultOpts)

	if _, err := run(mntCmd); err != nil {
		glog.Errorf("Failed to mount cmd: %s, err: %v", mntCmd, err)
		return err
	}
	glog.Infof("Success to mount cmd: %s", mntCmd)
	if err := os.RemoveAll(parameters[credential]); err != nil {
		glog.Errorf("Failed to remove passwd file, %v", err)
	}
	return nil
}

func run(cmd string) (string, error) {
	out, err := exec.Command("sh", "-c", cmd).CombinedOutput()
	if err != nil {
		return "", err
	}
	return string(out), nil
}
