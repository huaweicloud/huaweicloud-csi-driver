package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"

	log "k8s.io/klog/v2"

	"github.com/huaweicloud/huaweicloud-csi-driver/pkg/obs"
	"github.com/huaweicloud/huaweicloud-csi-driver/pkg/utils"
)

const (
	credentialDir = "/var/lib/csi"
	bucketName    = "bucketName"
	mountFlags    = "mountFlags"
	targetPath    = "targetPath"
	region        = "region"
	cloud         = "cloud"
	credential    = "credential"
	defaultOpts   = "-o big_writes -o max_write=131072 -o use_ino"
)

type ResponseBody struct {
	Data string
}

type ActionHandler struct{}

func (*ActionHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	body, err := io.ReadAll(request.Body)
	if err != nil {
		genResponse(writer, http.StatusBadRequest, err.Error())
		return
	}
	if len(body) == 0 {
		genResponse(writer, http.StatusBadRequest, "Request body is empty")
		return
	}
	log.Infof("ServerHTTP Request body: %s", string(body))
	var commandRPC obs.CommandRPC
	if err := json.Unmarshal(body, &commandRPC); err != nil {
		genResponse(writer, http.StatusBadRequest, err.Error())
		return
	}
	if err := checkRequestToken(commandRPC.Token, commandRPC.Parameters); err != nil {
		genResponse(writer, http.StatusBadRequest, fmt.Sprintf("Failed to check token, err: %v", err.Error()))
		return
	}

	if commandRPC.Action == obs.ActionMount {
		if err := checkMountParameters(commandRPC.Parameters); err != nil {
			genResponse(writer, http.StatusBadRequest, fmt.Sprintf("Failed to check mount action parameters, err: %v", err.Error()))
			return
		}
		if err := mountHandler(commandRPC.Parameters); err != nil {
			genResponse(writer, http.StatusInternalServerError, err.Error())
			return
		}
		genResponse(writer, http.StatusOK, "success")
		return
	}
	genResponse(writer, http.StatusBadRequest, fmt.Sprintf("Invalid action %s", commandRPC.Action))
}

func checkRequestToken(token string, parameters map[string]string) error {
	ciphertext := utils.Sha256(parameters)
	original, err := utils.DecryptAESCBC(obs.Secret, token)
	if err != nil {
		return err
	}
	if ciphertext != original {
		log.Errorf("Invalid token: %s %s", ciphertext, original)
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
	credentialFile := parameters[credential]
	if !strings.HasPrefix(credentialFile, credentialDir) {
		return fmt.Errorf("credential file can only use DIR: %s， current: %s", credentialDir, credentialFile)
	}
	if !checkFileExists(credentialFile) {
		return fmt.Errorf("credential file %s not exist", credential)
	}
	return nil
}

func genResponse(writer http.ResponseWriter, statusCode int, data string) {
	responseBody := ResponseBody{
		Data: data,
	}
	response, _ := json.Marshal(responseBody)
	writer.WriteHeader(statusCode)
	writer.Header().Set("Content-Type", "application/json")
	size, err := writer.Write(response)
	if err != nil {
		log.Errorf("Failed to write response: %s, err: %v", data, err)
		return
	}
	log.Infof("Response: %s size: %d", data, size)
}

func main() {
	//nolint:errcheck
	flag.CommandLine.Parse([]string{})

	initObsfsUtil()
	if checkFileExists(obs.SocketPath) {
		if err := os.Remove(obs.SocketPath); err != nil {
			log.Fatalf("Failed to remove path: %s, err: %v", obs.SocketPath, err)
		}
	}

	listen, err := net.Listen("unix", obs.SocketPath)
	if err != nil {
		log.Fatalf("net.Listen failed. path: %s, err: %v", obs.SocketPath, err)
		return
	}
	log.Infof("Success start listener")
	server := http.Server{Handler: &ActionHandler{}}
	if err := server.Serve(listen); err != nil {
		log.Fatalf("Socket Listener failed to close, err: %v", err)
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
	credentialFile := parameters[credential]
	defer deleteCredential(credentialFile)
	obsName := "obs"
	if parameters[cloud] == "prod-cloud-ocb.orange-business.com" {
		obsName = "oss"
	}

	mountOpts := parameters[mountFlags]
	if mountOpts == "" {
		mountOpts = defaultOpts
	}
	options := []string{
		"obsfs",
		parameters[bucketName],
		parameters[targetPath],
		fmt.Sprintf("-o url=%s.%s.%s", obsName, parameters[region], parameters[cloud]),
		fmt.Sprintf("-o passwd_file=%s", credentialFile),
		mountOpts,
	}

	cmd := exec.Command("sh", "-c")
	cmd.Args = append(cmd.Args, strings.Join(options, " "))
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Errorf("failed to mount CMD: obsfs %s, output: %s, error: %v", strings.Join(options, " "), string(out), err)
		return fmt.Errorf("failed to mount CMD: obsfs %s, output: %s, error: %v", strings.Join(options, " "), string(out), err)
	}
	log.Infof("success to mount CMD: %s", strings.Join(options, " "))
	return nil
}

func deleteCredential(credential string) {
	if strings.HasPrefix(credential, credentialDir) {
		if err := os.RemoveAll(credential); err != nil {
			log.Warningf("Failed to remove credential file: %s, error: %s", credential, err)
		}
	}
}

func initObsfsUtil() {
	cmd := fmt.Sprintf("sh %s/install_obsfs.sh >> %s/connector.log 2>&1 &", credentialDir, credentialDir)
	out, err := exec.Command("sh", "-c", cmd).CombinedOutput()
	log.Infof("install obsfs %s", string(out))
	if err != nil {
		log.Errorf("error install obsfs: %s", err)
	}
}
