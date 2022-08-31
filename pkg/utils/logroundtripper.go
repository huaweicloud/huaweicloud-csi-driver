/*
Copyright 2020 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"
	"sync/atomic"

	"github.com/kubernetes-csi/csi-lib-utils/protosanitizer"
	"github.com/unknwon/com"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"k8s.io/klog/v2"
)

var (
	serverGRPCEndpointCallCounter uint64
)

// LogRoundTripper satisfies the http.RoundTripper interface and is used to
// customize the default http client RoundTripper to allow for logging.
type LogRoundTripper struct {
	Rt http.RoundTripper
}

// RoundTrip performs a round-trip HTTP request and logs relevant information about it.
func (lrt *LogRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	defer func() {
		if request.Body != nil {
			request.Body.Close()
		}
	}()

	var err error

	klog.V(6).Infof("Request URL: %s %s", request.Method, request.URL)
	klog.V(6).Infof("Request Headers:\n%s", FormatHeaders(request.Header, "\n"))

	if request.Body != nil {
		request.Body, err = lrt.logRequest(request.Body, request.Header.Get("Content-Type"))
		if err != nil {
			return nil, err
		}
	}

	response, err := lrt.Rt.RoundTrip(request)
	if response == nil {
		return nil, err
	}

	klog.V(6).Infof("Response Code: %d", response.StatusCode)
	klog.V(6).Infof("Response Headers:\n%s", FormatHeaders(response.Header, "\n"))

	response.Body, err = lrt.logResponse(response.Body, response.Header.Get("Content-Type"))

	return response, err
}

// logRequest will log the HTTP Request details.
// If the body is JSON, it will attempt to be pretty-formatted.
func (lrt *LogRoundTripper) logRequest(original io.ReadCloser, contentType string) (io.ReadCloser, error) {
	defer original.Close()

	var bs bytes.Buffer
	_, err := io.Copy(&bs, original)
	if err != nil {
		return nil, err
	}

	// Handle request contentType
	if strings.HasPrefix(contentType, "application/json") {
		debugInfo := lrt.formatJSON(bs.Bytes())
		klog.V(6).Infof("Request Body: %s", debugInfo)
	} else {
		klog.V(6).Infof("Request Body: %s", bs.String())
	}

	return ioutil.NopCloser(strings.NewReader(bs.String())), nil
}

// logResponse will log the HTTP Response details.
// If the body is JSON, it will attempt to be pretty-formatted.
func (lrt *LogRoundTripper) logResponse(original io.ReadCloser, contentType string) (io.ReadCloser, error) {
	if strings.HasPrefix(contentType, "application/json") {
		var bs bytes.Buffer
		defer original.Close()
		_, err := io.Copy(&bs, original)
		if err != nil {
			return nil, err
		}
		debugInfo := lrt.formatJSON(bs.Bytes())
		if debugInfo != "" {
			klog.V(6).Infof("Response Body: %s", debugInfo)
		}
		return ioutil.NopCloser(strings.NewReader(bs.String())), nil
	}

	klog.V(6).Infof("Not logging because response body isn't JSON")
	return original, nil
}

// formatJSON will try to pretty-format a JSON body.
// It will also mask known fields which contain sensitive information.
func (lrt *LogRoundTripper) formatJSON(raw []byte) string {
	var data map[string]interface{}

	err := json.Unmarshal(raw, &data)
	if err != nil {
		klog.V(6).Infof("Unable to parse JSON: %s", err)
		return string(raw)
	}

	// Mask known password fields
	if v, ok := data["auth"].(map[string]interface{}); ok {
		if v, ok := v["identity"].(map[string]interface{}); ok {
			if v, ok := v["password"].(map[string]interface{}); ok {
				if v, ok := v["user"].(map[string]interface{}); ok {
					v["password"] = "***"
				}
			}
		}
	}

	// Ignore the catalog
	if v, ok := data["token"].(map[string]interface{}); ok {
		if _, ok := v["catalog"]; ok {
			return ""
		}
	}

	pretty, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		klog.V(6).Infof("Unable to re-marshal JSON: %s", err)
		return string(raw)
	}

	return string(pretty)
}

// RedactHeaders processes a headers object, returning a redacted list
func RedactHeaders(headers http.Header) (processedHeaders []string) {
	// redactheaders Lists of headers that need to be redacted
	var redactheaders = []string{"x-auth-token", "x-auth-key", "x-service-token",
		"x-storage-token", "x-account-meta-temp-url-key", "x-account-meta-temp-url-key-2",
		"x-container-meta-temp-url-key", "x-container-meta-temp-url-key-2", "set-cookie",
		"x-subject-token", "authorization"}

	for name, header := range headers {
		for _, v := range header {
			if com.IsSliceContainsStr(redactheaders, name) {
				processedHeaders = append(processedHeaders, fmt.Sprintf("%v: %v", name, "***"))
			} else {
				processedHeaders = append(processedHeaders, fmt.Sprintf("%v: %v", name, v))
			}
		}
	}
	return
}

// FormatHeaders processes a headers object plus a deliminator, returning a string
func FormatHeaders(headers http.Header, seperator string) string {
	redactedHeaders := RedactHeaders(headers)
	sort.Strings(redactedHeaders)

	return strings.Join(redactedHeaders, seperator)
}

func LogGRPC(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{},
	error) {

	callID := atomic.AddUint64(&serverGRPCEndpointCallCounter, 1)

	klog.V(5).Infof("[ID:%d] GRPC call: %s", callID, info.FullMethod)
	klog.V(5).Infof("[ID:%d] GRPC request: %s", callID, protosanitizer.StripSecrets(req))
	resp, err := handler(ctx, req)
	if err != nil {
		klog.Errorf("[ID:%d] GRPC call %s, error: %v", callID, info.FullMethod, err)
	} else {
		klog.V(5).Infof("[ID:%d] GRPC response: %s", callID, protosanitizer.StripSecrets(resp))
	}

	return resp, err
}
