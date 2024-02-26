package common

import (
	"errors"
	"testing"

	"github.com/chnsz/golangsdk"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/util/wait"
)

func TestIsNotFound(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "test1",
			err:      status.Error(codes.OK, "OK test"),
			expected: false,
		},
		{
			name:     "test2",
			err:      status.Error(codes.NotFound, "NotFound test"),
			expected: true,
		},
		{
			name:     "test3",
			err:      status.Error(codes.Internal, "Internal test"),
			expected: false,
		},
		{
			name:     "test4",
			err:      golangsdk.ErrDefault400{},
			expected: false,
		},
		{
			name:     "test5",
			err:      golangsdk.ErrDefault404{},
			expected: true,
		},
		{
			name:     "test6",
			err:      golangsdk.ErrDefault503{},
			expected: false,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			isNotFound := IsNotFound(testCase.err)
			if testCase.expected != isNotFound {
				t.Fatalf("expected: %v, got: %v", testCase.expected, isNotFound)
			}
		})
	}
}

func TestWaitForCompleted(t *testing.T) {
	tests := []struct {
		name        string
		condition   wait.ConditionFunc
		expected    bool
		expectedTxt string
	}{
		{
			name: "test1",
			condition: func() (done bool, err error) {
				return true, nil
			},
			expected: false,
		},
		{
			name: "test2",
			condition: func() (done bool, err error) {
				return true, errors.New("true message test")
			},
			expected:    true,
			expectedTxt: "true message test",
		},
		{
			name: "test3",
			condition: func() (done bool, err error) {
				return false, nil
			},
			expected:    true,
			expectedTxt: "timed out waiting for the condition",
		},
		{
			name: "test4",
			condition: func() (done bool, err error) {
				return false, errors.New("false message test")
			},
			expected:    true,
			expectedTxt: "false message test",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			err := WaitForCompleted(testCase.condition)
			if testCase.expected && err == nil {
				t.Fatalf("expected error but get nil")
			}
			if !testCase.expected && err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			if testCase.expected && err != nil && err.Error() != testCase.expectedTxt {
				t.Fatalf("expected: %v, got: %v", testCase.expected, err.Error())
			}
		})
	}
}

func TestGetAZFromTopology(t *testing.T) {
	topologyKey := "zone"

	testCases := []struct {
		name     string
		req      *csi.TopologyRequirement
		expected string
	}{
		{
			name: "test1",
			req: &csi.TopologyRequirement{
				Preferred: []*csi.Topology{{
					Segments: map[string]string{
						topologyKey: "zone1",
					}},
				},
			},
			expected: "zone1",
		},
		{
			name: "test2",
			req: &csi.TopologyRequirement{
				Requisite: []*csi.Topology{{
					Segments: map[string]string{
						topologyKey: "zone2",
					}},
				},
			},
			expected: "zone2",
		},
		{
			name: "test3",
			req: &csi.TopologyRequirement{
				Preferred: []*csi.Topology{{
					Segments: map[string]string{
						topologyKey: "zone1",
					}},
				},
				Requisite: []*csi.Topology{{
					Segments: map[string]string{
						topologyKey: "zone2",
					}},
				},
			},
			expected: "zone1",
		},
		{
			name: "test4",
			req: &csi.TopologyRequirement{
				Preferred: []*csi.Topology{{
					Segments: map[string]string{}},
				},
				Requisite: []*csi.Topology{{
					Segments: map[string]string{}},
				},
			},
			expected: "",
		},
		{
			name: "test5",
			req: &csi.TopologyRequirement{
				Preferred: []*csi.Topology{{
					Segments: map[string]string{
						topologyKey: "zone2",
					}},
				},
				Requisite: []*csi.Topology{{
					Segments: map[string]string{
						topologyKey: "zone2",
					}},
				},
			},
			expected: "zone2",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result := GetAZFromTopology(testCase.req, topologyKey)
			if testCase.expected != result {
				t.Fatalf("expected: %v, got: %v", testCase.expected, result)
			}
		})
	}
}
