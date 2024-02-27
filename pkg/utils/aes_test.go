package utils

import (
	"fmt"
	"testing"
)

const (
	keyLength16    = "hnbhacdsfsdfsadn"
	keyLength24    = "hnbhacsfsd187hds287fsadn"
	keyLength32    = "hnbhacdsfsd1296873hads283g7fsadn"
	keySymbol      = "hnbhacdsfsd-_=n@"
	keyNone        = ""
	keyOtherLength = "cdsfsdfs"
)

func TestEncryptAESCBC(t *testing.T) {
	tests := []struct {
		name        string
		key         string
		original    string
		expected    bool
		description string
	}{
		{
			name:        "test1",
			key:         keyLength16,
			original:    "xxx",
			expected:    true,
			description: "normal, 16 bits key",
		},
		{
			name:        "test2",
			key:         keyLength16,
			original:    "xxx",
			expected:    true,
			description: "key and original are repeat with test1 but result does not",
		},
		{
			name:        "test3",
			key:         keyLength16,
			original:    "xxx123!-_=?",
			expected:    true,
			description: "original contains letters, numbers and symbols",
		},
		{
			name:        "test4",
			key:         keySymbol,
			original:    "xxx",
			expected:    true,
			description: "key contains letters, numbers and symbols",
		},
		{
			name:        "test5",
			key:         keyLength16,
			original:    "",
			expected:    true,
			description: "original is empty",
		},
		{
			name:        "test6",
			key:         keyLength24,
			original:    "xxx",
			expected:    true,
			description: "24 bits key",
		},
		{
			name:        "test7",
			key:         keyLength32,
			original:    "xxx",
			expected:    true,
			description: "32 bits key",
		},
		{
			name:        "test8",
			key:         keyOtherLength,
			original:    "xxx",
			expected:    false,
			description: "other bits key",
		},
		{
			name:        "test9",
			key:         keyNone,
			original:    "xxx",
			expected:    false,
			description: "the key is empty",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			result, err := EncryptAESCBC(testCase.key, testCase.original)
			fmt.Println(result)
			if testCase.expected && err != nil {
				t.Errorf("expected: %v, but got error: %v", testCase.expected, err)
			}
		})
	}
}

func TestDecryptAESCBC(t *testing.T) {
	tests := []struct {
		name        string
		key         string
		ciphertext  string
		expected    bool
		description string
	}{
		{
			name:        "test1",
			key:         keyLength16,
			ciphertext:  "ThrTeoOfCEsVRvKFEgW4ZG6BRNygr2bta1O1lFWm4Ho=",
			expected:    true,
			description: "normal, 16 bits key",
		},
		{
			name:        "test2",
			key:         keyLength16,
			ciphertext:  "ThrTeoOfCEsVRvKFEgW4ZG6BRNygr2bta1O1lFWm4Ho=",
			expected:    true,
			description: "key and ciphertext are repeat with test1, and result does",
		},
		{
			name:        "test3",
			key:         keyLength16,
			ciphertext:  "TswZGBanBtk5qQz+qT5SygGQdPZGf7FuimBs3MEM/e8=",
			expected:    true,
			description: "original contains letters, numbers and symbols",
		},
		{
			name:        "test4",
			key:         keySymbol,
			ciphertext:  "NJoBGH0Ywk69OPkCT3ljcguWOtgunCZxRfVaaHAms4M=",
			expected:    true,
			description: "key contains letters, numbers and symbols",
		},
		{
			name:        "test5",
			key:         keyLength24,
			ciphertext:  "taWef5OtOTvjmZog59j5NWme/zwqWwnJwb5a9ui9W4w=",
			expected:    true,
			description: "24 bits key",
		},
		{
			name:        "test6",
			key:         keyLength32,
			ciphertext:  "wHbjK3ftTqggd8wcLhtIDO9D+YcQSQHJ41LJuAGgSvA=",
			expected:    true,
			description: "32 bits key",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := DecryptAESCBC(testCase.key, testCase.ciphertext)
			if testCase.expected && err != nil {
				t.Errorf("expected: %v, but got error: %v", testCase.expected, err)
			}
		})
	}
}
