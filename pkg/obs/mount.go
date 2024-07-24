package obs

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	log "k8s.io/klog/v2"
)

const (
	ActionMount string = "mount"
)

type CommandRPC struct {
	Action     string
	Token      string
	Parameters map[string]string
}

func sendCommand(cmd CommandRPC, mountClient http.Client) error {
	marshal, err := json.Marshal(cmd)
	if err != nil {
		return err
	}
	response, err := mountClient.Post("http://unix", "application/json", bytes.NewReader(marshal))
	if err != nil {
		return status.Errorf(codes.Internal, "Failed to post command, err: %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		respBody, err := ioutil.ReadAll(response.Body)
		log.Infof("start to mount bucket, cmd: %v", string(respBody))
		if err != nil {
			return status.Errorf(codes.Internal, "Failed to read responseBody, err: %v", err)
		}
		return status.Errorf(codes.Internal, "Failed to execute the command, body: %v", string(respBody))
	}
	return nil
}
