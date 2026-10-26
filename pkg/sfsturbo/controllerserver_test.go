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

package sfsturbo

import (
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestResolveExpandParams(t *testing.T) {
	testCases := []struct {
		name           string
		parameters     map[string]string
		wantExpandType string
		wantHpcBw      string
		wantIsHpc      bool
	}{
		{
			name:           "neither expandType nor hpcBw -> legacy share",
			parameters:     map[string]string{"shareType": "STANDARD"},
			wantExpandType: "",
			wantHpcBw:      "",
			wantIsHpc:      false,
		},
		{
			name:           "only shareType set -> legacy share",
			parameters:     map[string]string{"shareType": "PERFORMANCE"},
			wantExpandType: "",
			wantHpcBw:      "",
			wantIsHpc:      false,
		},
		{
			name:           "hpcBw alone infers expandType=hpc",
			parameters:     map[string]string{"shareType": "STANDARD", "hpcBw": "20M"},
			wantExpandType: "hpc",
			wantHpcBw:      "20M",
			wantIsHpc:      true,
		},
		{
			name:           "both expandType=hpc and hpcBw set",
			parameters:     map[string]string{"expandType": "hpc", "hpcBw": "125M"},
			wantExpandType: "hpc",
			wantHpcBw:      "125M",
			wantIsHpc:      true,
		},
		{
			name:           "expandType=hpc without hpcBw (value validated by API)",
			parameters:     map[string]string{"expandType": "hpc"},
			wantExpandType: "hpc",
			wantHpcBw:      "",
			wantIsHpc:      true,
		},
		{
			name:           "non-hpc expandType is passed through untouched",
			parameters:     map[string]string{"expandType": "bandwidth", "hpcBw": "20M"},
			wantExpandType: "bandwidth",
			wantHpcBw:      "20M",
			wantIsHpc:      false,
		},
		{
			name:           "hpcBw with non-hpc expandType is not inferred",
			parameters:     map[string]string{"expandType": "hpc_cache", "hpcBw": "2G"},
			wantExpandType: "hpc_cache",
			wantHpcBw:      "2G",
			wantIsHpc:      false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			expandType, hpcBw := resolveExpandParams(tc.parameters)
			isHpc := expandType == expandTypeHpc
			if expandType != tc.wantExpandType {
				t.Errorf("expandType = %q, want %q", expandType, tc.wantExpandType)
			}
			if hpcBw != tc.wantHpcBw {
				t.Errorf("hpcBw = %q, want %q", hpcBw, tc.wantHpcBw)
			}
			if isHpc != tc.wantIsHpc {
				t.Errorf("isHpc = %v, want %v", isHpc, tc.wantIsHpc)
			}
		})
	}
}

func TestValidateLegacyCreateSize(t *testing.T) {
	testCases := []struct {
		name      string
		sizeInGiB int
		wantSize  int
		wantErr   bool
		wantCode  codes.Code
	}{
		{name: "below minimum is raised to 500", sizeInGiB: 100, wantSize: 500, wantErr: false},
		{name: "exactly minimum", sizeInGiB: 500, wantSize: 500, wantErr: false},
		{name: "mid range", sizeInGiB: 1000, wantSize: 1000, wantErr: false},
		{name: "exactly maximum", sizeInGiB: 32768, wantSize: 32768, wantErr: false},
		{name: "above maximum is rejected", sizeInGiB: 32769, wantSize: 0, wantErr: true, wantCode: codes.InvalidArgument},
		{name: "far above maximum is rejected", sizeInGiB: 1048576, wantSize: 0, wantErr: true, wantCode: codes.InvalidArgument},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := validateLegacyCreateSize(tc.sizeInGiB)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if tc.wantCode != codes.OK {
					if s, ok := status.FromError(err); ok {
						if s.Code() != tc.wantCode {
							t.Errorf("error code = %v, want %v", s.Code(), tc.wantCode)
						}
					} else {
						t.Errorf("error is not a grpc status error: %v", err)
					}
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.wantSize {
				t.Errorf("size = %d, want %d", got, tc.wantSize)
			}
		})
	}
}
