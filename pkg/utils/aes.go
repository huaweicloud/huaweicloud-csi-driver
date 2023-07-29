package utils

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"io"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func EncryptAESCBC(key, original string) (string, error) {
	originalBytes := []byte(original)
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", status.Errorf(codes.Internal, "failed to init cipher: %v", err)
	}

	padding := aes.BlockSize - len(originalBytes)%aes.BlockSize
	paddingText := append(originalBytes, bytes.Repeat([]byte{byte(padding)}, padding)...)

	ciphertext := make([]byte, aes.BlockSize+len(paddingText))
	iv := ciphertext[:aes.BlockSize]

	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext[aes.BlockSize:], paddingText)

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func DecryptAESCBC(key, ciphertext string) (string, error) {
	cipherBytes, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", status.Errorf(codes.Internal, "failed to decode string: %v", err)
	}

	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", status.Errorf(codes.Internal, "failed to init cipher: %v", err)
	}

	iv := cipherBytes[:aes.BlockSize]
	cipherBytes = cipherBytes[aes.BlockSize:]

	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(cipherBytes, cipherBytes)

	padding := int(cipherBytes[len(cipherBytes)-1])
	decryptedText := cipherBytes[:len(cipherBytes)-padding]

	return string(decryptedText), nil
}
