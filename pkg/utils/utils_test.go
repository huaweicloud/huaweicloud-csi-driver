package utils

import (
	"testing"
)

func TestParseEndpoint(t *testing.T) {
	tests := []struct {
		name         string
		endpoint     string
		expected     bool
		expectedTxt1 string
		expectedTxt2 string
		description  string
	}{
		{
			name:         "test1",
			endpoint:     "unix://xxx.123.yyy",
			expected:     true,
			expectedTxt1: "unix",
			expectedTxt2: "xxx.123.yyy",
			description:  "unix normal",
		},
		{
			name:         "test2",
			endpoint:     "tcp://xxx.123.yyy",
			expected:     true,
			expectedTxt1: "tcp",
			expectedTxt2: "xxx.123.yyy",
			description:  "tcp normal",
		},
		{
			name:        "test3",
			endpoint:    "unix://",
			expected:    false,
			description: "address of request is empty",
		},
		{
			name:        "test4",
			endpoint:    "unix:://xxx.123.yyy",
			expected:    false,
			description: "url format is error",
		},
		{
			name:         "test5",
			endpoint:     "unix:///xxx.123.yyy",
			expected:     true,
			expectedTxt1: "unix",
			expectedTxt2: "/xxx.123.yyy",
			description:  "url format have three /",
		},
		{
			name:         "test6",
			endpoint:     "unix://xxx.123.yyy.tcp://",
			expected:     true,
			expectedTxt1: "unix",
			expectedTxt2: "xxx.123.yyy.tcp://",
			description:  "at the end of the url has prefix format string",
		},
		{
			name:        "test7",
			endpoint:    "unix.xxx.yyy",
			expected:    false,
			description: "at the start of the url has no prefix format string",
		},
		{
			name:        "test8",
			endpoint:    "",
			expected:    false,
			description: "url is empty",
		},
		{
			name:         "test9",
			endpoint:     "unix://!@#$%^&*?|-_=\\",
			expected:     true,
			expectedTxt1: "unix",
			expectedTxt2: "!@#$%^&*?|-_=\\",
			description:  "url has symbol",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			res1, res2, err := ParseEndpoint(testCase.endpoint)
			if testCase.expected == true && err != nil {
				t.Errorf("expected: %v;%v;%v, got: %v;%v;%v",
					testCase.expected, testCase.expectedTxt1, testCase.expectedTxt2, err, res1, res2)
			}
			if testCase.expected == false && err == nil {
				t.Errorf("expected error but get nil")
			}
		})
	}
}

func TestRoundUpSize(t *testing.T) {
	tests := []struct {
		name                string
		volumeSizeBytes     int64
		allocationUnitBytes int64
		expected            int64
		description         string
	}{
		{
			name:                "test1",
			volumeSizeBytes:     100,
			allocationUnitBytes: 1024,
			expected:            1,
			description:         "less than Ki",
		},
		{
			name:                "test2",
			volumeSizeBytes:     1024,
			allocationUnitBytes: 1024,
			expected:            1,
			description:         "equal to Ki",
		},
		{
			name:                "test3",
			volumeSizeBytes:     100 * 1024,
			allocationUnitBytes: 1024,
			expected:            100,
			description:         "larger than Ki",
		},
		{
			name:                "test4",
			volumeSizeBytes:     100 * 1024,
			allocationUnitBytes: 1024 * 1024,
			expected:            1,
			description:         "Ki round up to Mi",
		},
		{
			name:                "test5",
			volumeSizeBytes:     100 * 1024 * 1024,
			allocationUnitBytes: 1024 * 1024,
			expected:            100,
			description:         "Mi round up to Mi",
		},
		{
			name:                "test6",
			volumeSizeBytes:     10 * 1024 * 1024 * 1024,
			allocationUnitBytes: 1024 * 1024,
			expected:            10240,
			description:         "Gi round up to Mi",
		},
		{
			name:                "test7",
			volumeSizeBytes:     100 * 1024,
			allocationUnitBytes: 1024 * 1024 * 1024,
			expected:            1,
			description:         "Ki round up to Gi",
		},
		{
			name:                "test8",
			volumeSizeBytes:     100 * 1024 * 1024,
			allocationUnitBytes: 1024 * 1024 * 1024,
			expected:            1,
			description:         "Mi round up to Gi",
		},
		{
			name:                "test9",
			volumeSizeBytes:     0.5 * 1024 * 1024 * 1024,
			allocationUnitBytes: 1024 * 1024 * 1024,
			expected:            1,
			description:         "decimal Gi round up to Gi",
		},
		{
			name:                "test10",
			volumeSizeBytes:     100 * 1024 * 1024 * 1024,
			allocationUnitBytes: 1024 * 1024 * 1024,
			expected:            100,
			description:         "Gi round up to Gi",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			roundedUp := RoundUpSize(testCase.volumeSizeBytes, testCase.allocationUnitBytes)
			if testCase.expected != roundedUp {
				t.Errorf("expected: %v, got %v", testCase.expected, roundedUp)
			}
		})
	}
}

func TestBytesToGB(t *testing.T) {
	tests := []struct {
		name        string
		size        interface{}
		expected    int64
		description string
	}{
		{
			name:        "test1",
			size:        1,
			expected:    1073741824,
			description: "normal",
		},
		{
			name:        "test2",
			size:        2,
			expected:    2147483648,
			description: "normal",
		},
		{
			name:        "test3",
			size:        0.5,
			expected:    -1,
			description: "decimal Byte to GB",
		},
		{
			name:        "test4",
			size:        0,
			expected:    0,
			description: "zero Byte to GB",
		},
		{
			name:        "test5",
			size:        -1,
			expected:    -1073741824,
			description: "negative Byte to GB",
		},
		{
			name:        "test6",
			size:        -0.5,
			expected:    -1,
			description: "Negative decimal Byte to GB",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result := BytesToGB(testCase.size)
			if testCase.expected != result {
				t.Errorf("expected: %v, got %v", testCase.expected, result)
			}
		})
	}
}
