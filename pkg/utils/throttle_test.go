/*
Copyright 2026 The Kubernetes Authors.

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
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func newTestServer(t *testing.T, fail int, bodies *[]string) (*httptest.Server, *int) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		b, _ := io.ReadAll(r.Body)
		*bodies = append(*bodies, string(b))
		if calls <= fail {
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"error_code":"APIGW.0308"}`))
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)
	return srv, &calls
}

func TestThrottleRetriesAfter429WithSameBody(t *testing.T) {
	var bodies []string
	srv, calls := newTestServer(t, 2, &bodies)
	var slept []time.Duration
	rt := NewThrottleRoundTripper(http.DefaultTransport, 0, 0, 3)
	rt.sleep = func(d time.Duration) { slept = append(slept, d) }

	req, _ := http.NewRequest(http.MethodPost, srv.URL, bytes.NewReader([]byte(`{"volume":{}}`)))
	resp, err := rt.RoundTrip(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK || *calls != 3 || len(slept) != 2 {
		t.Fatalf("status %d after %d calls and %d sleeps", resp.StatusCode, *calls, len(slept))
	}
	for i, b := range bodies {
		if b != `{"volume":{}}` {
			t.Fatalf("call %d sent body %q", i+1, b)
		}
	}
}

func TestThrottleGivesUpAfterRetries(t *testing.T) {
	var bodies []string
	srv, calls := newTestServer(t, 10, &bodies)
	rt := NewThrottleRoundTripper(http.DefaultTransport, 0, 0, 2)
	rt.sleep = func(time.Duration) {}

	req, _ := http.NewRequest(http.MethodGet, srv.URL, nil)
	resp, err := rt.RoundTrip(req)
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != http.StatusTooManyRequests || *calls != 3 || !strings.Contains(string(body), "APIGW.0308") {
		t.Fatalf("status %d after %d calls, body %q", resp.StatusCode, *calls, body)
	}
}

func TestThrottleNoRetriesByDefault(t *testing.T) {
	var bodies []string
	srv, calls := newTestServer(t, 1, &bodies)
	rt := NewThrottleRoundTripper(http.DefaultTransport, 0, 0, 0)

	req, _ := http.NewRequest(http.MethodGet, srv.URL, nil)
	resp, err := rt.RoundTrip(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusTooManyRequests || *calls != 1 {
		t.Fatalf("status %d after %d calls", resp.StatusCode, *calls)
	}
}

func TestThrottleLimitsRate(t *testing.T) {
	var bodies []string
	srv, _ := newTestServer(t, 0, &bodies)
	rt := NewThrottleRoundTripper(http.DefaultTransport, 20, 1, 0)

	start := time.Now()
	for i := 0; i < 5; i++ {
		req, _ := http.NewRequest(http.MethodGet, srv.URL, nil)
		resp, err := rt.RoundTrip(req)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
	}
	// burst 1 at 20/s: the 4 calls after the first wait 50ms each
	if elapsed := time.Since(start); elapsed < 190*time.Millisecond {
		t.Fatalf("5 calls took %s, want at least 200ms", elapsed)
	}
}

func TestRetryDelay(t *testing.T) {
	resp := &http.Response{Header: http.Header{"Retry-After": []string{"7"}}}
	if d := retryDelay(resp, 0); d != 7*time.Second {
		t.Fatalf("Retry-After 7: got %s", d)
	}
	resp.Header = http.Header{}
	for attempt, max := range []time.Duration{2, 4, 8, 16, 32, 60, 60, 60} {
		d := retryDelay(resp, attempt)
		if d < max*time.Second/2 || d > max*time.Second {
			t.Fatalf("attempt %d: %s outside [%s, %s]", attempt, d, max*time.Second/2, max*time.Second)
		}
	}
}
