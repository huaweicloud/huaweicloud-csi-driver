package utils

import (
	"io"
	"net/http"
	"testing"
)

func TestRoundTrip(t *testing.T) {
	tests := []struct {
		name        string
		request     *http.Request
		response    *http.Response
		expected    bool
		description string
	}{
		{
			name:    "test1",
			request: httpNewRequest("GET", "https://hub.docker.com/", nil, t),
			response: &http.Response{
				Status:     "200 OK",
				StatusCode: http.StatusOK,
				Proto:      "HTTP/1.1",
				ProtoMajor: 1,
				ProtoMinor: 1,
			},
			expected:    true,
			description: "request and response are real entities",
		},
		{
			name:    "test2",
			request: httpNewRequest("GET", "https://hub.docker.com/", nil, t),
			response: &http.Response{
				Status:     "404 Not Found",
				StatusCode: http.StatusNotFound,
				Proto:      "HTTP/1.1",
			},
			expected:    false,
			description: "request is real entity, but response is fake",
		},
		{
			name:        "test3",
			request:     httpNewRequest("GET", "https://hub.docker.com/", nil, t),
			response:    &http.Response{},
			expected:    false,
			description: "request is real entity, but response is empty",
		},
		{
			name:        "test4",
			request:     httpNewRequest("GET", "https://abishdjlnsadndh790h123dadsffdsfsDFbhdvhdsadsa.com", nil, t),
			response:    &http.Response{},
			expected:    false,
			description: "url is fake",
		},
		{
			name:    "test5",
			request: httpNewRequest("", "", nil, t),
			response: &http.Response{
				Status:     "200 OK",
				StatusCode: http.StatusOK,
				Proto:      "HTTP/1.1",
			},
			expected:    false,
			description: "request is empty",
		},
		{
			name:    "test6",
			request: httpNewRequest("POST", "https://hub.docker.com/", nil, t),
			response: &http.Response{
				Status:     "405 Method Not Allowed",
				StatusCode: http.StatusMethodNotAllowed,
				Proto:      "HTTP/1.1",
				ProtoMajor: 1,
				ProtoMinor: 1,
			},
			expected:    true,
			description: "the method of request is POST and url is real entity",
		},
		{
			name:    "test7",
			request: httpNewRequest("PUT", "https://hub.docker.com/", nil, t),
			response: &http.Response{
				Status:     "405 Method Not Allowed",
				StatusCode: http.StatusMethodNotAllowed,
				Proto:      "HTTP/1.1",
				ProtoMajor: 1,
				ProtoMinor: 1,
			},
			expected:    true,
			description: "the method of request is PUT and url is real entity",
		},
		{
			name:    "test8",
			request: httpNewRequest("DELETE", "https://hub.docker.com/", nil, t),
			response: &http.Response{
				Status:     "405 Method Not Allowed",
				StatusCode: http.StatusMethodNotAllowed,
				Proto:      "HTTP/1.1",
				ProtoMajor: 1,
				ProtoMinor: 1,
			},
			expected:    true,
			description: "the method of request is DELETE and url is real entity",
		},
		{
			name:    "test9",
			request: httpNewRequest("xxx", "https://hub.docker.com/", nil, t),
			response: &http.Response{
				Status:     "404 Not Found",
				StatusCode: http.StatusNotFound,
				Proto:      "HTTP/1.1",
			},
			expected:    false,
			description: "the method of request is not exists",
		},
		{
			name:    "test10",
			request: httpNewRequest("GET", "https://hub.docker.com/", nil, t),
			response: &http.Response{
				Status:     "200 OK",
				StatusCode: http.StatusOK,
				Proto:      "HTTP/2.0",
			},
			expected:    false,
			description: "the proto of customized response does not match proto of actual response",
		},
		{
			name:    "test11",
			request: httpNewRequest("GET", "https://hub.docker.com/", nil, t),
			response: &http.Response{
				Status:     "200 OK",
				StatusCode: http.StatusOK,
				Proto:      "HTTP/1.1",
				ProtoMajor: 1,
				ProtoMinor: 1,
			},
			expected:    true,
			description: "the ProtoMajor of customized response is 1",
		},
		{
			name:    "test12",
			request: httpNewRequest("GET", "https://hub.docker.com/", nil, t),
			response: &http.Response{
				Status:     "200 OK",
				StatusCode: http.StatusOK,
				Proto:      "HTTP/1.1",
				ProtoMajor: 1,
				ProtoMinor: 1,
			},
			expected:    true,
			description: "the ProtoMinor of customized response is 1",
		},
		{
			name:    "test13",
			request: httpNewRequest("GET", "https://hub.docker.com/", nil, t),
			response: &http.Response{
				Status:     "200 OK",
				StatusCode: http.StatusOK,
				Proto:      "HTTP/1.0",
				ProtoMajor: 2,
			},
			expected:    false,
			description: "the ProtoMajor of customized response is 2",
		},
		{
			name:    "test14",
			request: httpNewRequest("GET", "https://hub.docker.com/", nil, t),
			response: &http.Response{
				Status:     "200 OK",
				StatusCode: http.StatusOK,
				Proto:      "HTTP/1.0",
				ProtoMinor: 2,
			},
			expected:    false,
			description: "the ProtoMinor of customized response is 2",
		},
		{
			name:    "test15",
			request: httpNewRequest("GET", "https://github.com/", nil, t),
			response: &http.Response{
				Status:     "200 OK",
				StatusCode: http.StatusOK,
				Proto:      "HTTP/2.0",
				ProtoMajor: 2,
				ProtoMinor: 0,
			},
			expected:    true,
			description: "request and response are real entities, the url is different",
		},
		{
			name:    "test16",
			request: httpNewRequest("GET", "https://www.huaweicloud.com/", nil, t),
			response: &http.Response{
				Status:     "200 OK",
				StatusCode: http.StatusOK,
				Proto:      "HTTP/2.0",
				ProtoMajor: 2,
				ProtoMinor: 0,
			},
			expected:    true,
			description: "request and response are real entities, the url is different",
		},
		{
			name:    "test17",
			request: httpNewRequest("POST", "https://www.huaweicloud.com/", nil, t),
			response: &http.Response{
				Status:     "405 Method Not Allowed",
				StatusCode: http.StatusMethodNotAllowed,
				Proto:      "HTTP/2.0",
				ProtoMajor: 2,
				ProtoMinor: 0,
			},
			expected:    false,
			description: "request and response are real entities, the url and method are different",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			transport := &http.Transport{Proxy: http.ProxyFromEnvironment}
			rt := &LogRoundTripper{
				Rt: transport,
			}
			response, err := rt.RoundTrip(testCase.request)

			if testCase.expected && err != nil {
				t.Errorf("expected: %v, but got err: %v", testCase.expected, err)
			}
			if testCase.expected && testCase.response.Status != response.Status {
				t.Errorf("expected response status: %v, got: %v", testCase.response.Status, response.Status)
			}
			if testCase.expected && testCase.response.StatusCode != response.StatusCode {
				t.Errorf("expected response status code: %v, got: %v", testCase.response.StatusCode, response.StatusCode)
			}
			if testCase.expected && testCase.response.Proto != response.Proto {
				t.Errorf("expected response proto: %v, got: %v", testCase.response.Proto, response.Proto)
			}
			if testCase.expected && testCase.response.ProtoMajor != response.ProtoMajor {
				t.Errorf("expected response proto: %v, got: %v", testCase.response.ProtoMajor, response.ProtoMajor)
			}
			if testCase.expected && testCase.response.ProtoMinor != response.ProtoMinor {
				t.Errorf("expected response proto: %v, got: %v", testCase.response.ProtoMinor, response.ProtoMinor)
			}
		})
	}
}

func httpNewRequest(method, url string, body io.Reader, t *testing.T) *http.Request {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		t.Errorf("failed to create new request: %v", err)
		return nil
	}
	return req
}
