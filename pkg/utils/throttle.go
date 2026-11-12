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
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"golang.org/x/time/rate"
	"k8s.io/klog/v2"
)

const (
	throttleInitDelay = 2 * time.Second
	throttleMaxDelay  = 60 * time.Second
)

// ThrottleRoundTripper keeps API calls under a rate limit and retries requests the
// gateway rejected with 429 Too Many Requests.
type ThrottleRoundTripper struct {
	Rt      http.RoundTripper
	Limiter *rate.Limiter // nil: no client-side limit
	Retries int           // retries after a 429, 0: none

	sleep func(time.Duration)
}

// NewThrottleRoundTripper returns rt wrapped with a limit of perSecond calls (burst calls at once)
// and up to retries retries after a 429. perSecond <= 0 means no client-side limit.
func NewThrottleRoundTripper(rt http.RoundTripper, perSecond float64, burst, retries int) *ThrottleRoundTripper {
	t := &ThrottleRoundTripper{Rt: rt, Retries: retries}
	if perSecond > 0 {
		if burst < 1 {
			burst = 1
		}
		t.Limiter = rate.NewLimiter(rate.Limit(perSecond), burst)
	}
	return t
}

// RoundTrip waits for the limiter, sends the request and retries it after a 429.
func (t *ThrottleRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	sleep := t.sleep
	if sleep == nil {
		sleep = time.Sleep
	}
	for attempt := 0; ; attempt++ {
		if t.Limiter != nil {
			if err := t.Limiter.Wait(req.Context()); err != nil {
				return nil, err
			}
		}

		r := req
		if attempt > 0 {
			r = req.Clone(req.Context())
			if req.GetBody != nil {
				body, err := req.GetBody()
				if err != nil {
					return nil, err
				}
				r.Body = body
			}
		}

		resp, err := t.Rt.RoundTrip(r)
		if err != nil || resp.StatusCode != http.StatusTooManyRequests || attempt >= t.Retries {
			return resp, err
		}
		// a request with a body can only be sent again if it can be rebuilt
		if req.Body != nil && req.GetBody == nil {
			return resp, nil
		}

		delay := retryDelay(resp, attempt)
		resp.Body.Close()
		klog.Warningf("%s %s: 429 Too Many Requests, retry %d/%d in %s",
			req.Method, req.URL.Path, attempt+1, t.Retries, delay.Round(time.Millisecond))
		sleep(delay)
	}
}

// retryDelay honors Retry-After in seconds, otherwise doubles from 2s up to 60s, with jitter.
func retryDelay(resp *http.Response, attempt int) time.Duration {
	if s, err := strconv.Atoi(resp.Header.Get("Retry-After")); err == nil && s > 0 {
		return time.Duration(s) * time.Second
	}
	d := throttleInitDelay << uint(attempt)
	if d > throttleMaxDelay || d <= 0 {
		d = throttleMaxDelay
	}
	return d/2 + time.Duration(rand.Int63n(int64(d/2)+1)) //nolint:gosec // jitter only, not security sensitive
}
