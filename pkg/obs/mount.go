package obs

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io/ioutil"
	log "k8s.io/klog/v2"
	"net/http"
	"strings"
)

func sendCommand(cmd, url string, mountClient http.Client) error {
	log.Infof("Start sending command: %s to socket listener: %s", cmd, url)
	response, err := mountClient.Post(url, "application/json", strings.NewReader(cmd))
	if err != nil {
		return status.Errorf(codes.Internal, "Failed to post command, err: %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		respBody, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return status.Errorf(codes.Internal, "Failed to read responseBody, err: %v", err)
		}
		return status.Errorf(codes.Internal, "Failed to execute the command, body: %v", string(respBody))
	}
	return nil
}
