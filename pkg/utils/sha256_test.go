package utils

import (
	"testing"
)

func TestSha256(t *testing.T) {
	tests := []struct {
		name        string
		value       map[string]string
		expected    string
		description string
	}{
		{
			name:        "test1",
			value:       map[string]string{"a": "aaa"},
			expected:    "db4d92158cf0e6a0704b25876dd16617e0f5d8442113111e92a92a59e3bb5c0b",
			description: "normal",
		},
		{
			name:        "test2",
			value:       map[string]string{"a": "aaa", "b": "bbb", "c": "ccc"},
			expected:    "3240c32e312e948a28409764a5b012a74d4bc7f09ddb4276537d8690b47b0113",
			description: "normal",
		},
		{
			name:        "test3",
			value:       map[string]string{"b": "bbb", "a": "aaa", "c": "ccc"},
			expected:    "3240c32e312e948a28409764a5b012a74d4bc7f09ddb4276537d8690b47b0113",
			description: "encrypting the same data will get the same result, even the order is different",
		},
		{
			name:        "test4",
			value:       map[string]string{"c": "ccc", "a": "aaa", "b": "bbb"},
			expected:    "3240c32e312e948a28409764a5b012a74d4bc7f09ddb4276537d8690b47b0113",
			description: "encrypting the same data will get the same result, even the order is different",
		},
		{
			name:     "test5",
			value:    map[string]string{},
			expected: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		},
		{
			name:     "test6",
			value:    map[string]string{"a": ""},
			expected: "96a36cea6c5e54242c319d0b353f76decb03ba9358b6ccdf91dea37d828357a7",
		},
		{
			name:     "test7",
			value:    map[string]string{"": ""},
			expected: "380918b946a526640a40df5dced6516794f3d97bbd9e6bb553d037c4439f31c3",
		},
		{
			name:     "test8",
			value:    map[string]string{"": "aaa"},
			expected: "e02c35825dc8971d439d972fa795a5e3f1c3bd33ff355e1915b24c474d8be754",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			res := Sha256(testCase.value)
			if testCase.expected != res {
				t.Errorf("expected: %v, got %v", testCase.expected, res)
			}
		})
	}
}
