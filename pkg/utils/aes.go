package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func EncryptAESCBC(key, original string) (string, error) {
	originalBytes := []byte(original)
	if len(originalBytes)%aes.BlockSize != 0 {
		return "", status.Errorf(codes.InvalidArgument, "Length of original Invalid, %v", len(originalBytes))
	}
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", status.Errorf(codes.Internal, "Failed to init cipher, err: %v", err)
	}
	cipherBytes := make([]byte, aes.BlockSize+len(originalBytes))
	iv := cipherBytes[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(cipherBytes[aes.BlockSize:], originalBytes)
	return fmt.Sprintf("%x", cipherBytes), nil
}

func DecryptAESCBC(key, ciphertext string) (string, error) {
	cipherBytes, _ := hex.DecodeString(ciphertext)
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", status.Errorf(codes.Internal, "Failed to init cipher, err: %v", err)
	}
	if len(cipherBytes) < aes.BlockSize {
		return "", status.Errorf(codes.InvalidArgument, "Length of ciphertext too short, %v", len(cipherBytes))
	}
	iv := cipherBytes[:aes.BlockSize]
	cipherBytes = cipherBytes[aes.BlockSize:]
	if len(cipherBytes)%aes.BlockSize != 0 {
		return "", status.Errorf(codes.InvalidArgument, "Length of ciphertext Invalid, %v", len(cipherBytes))
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(cipherBytes, cipherBytes)
	return string(cipherBytes[:]), nil
}
