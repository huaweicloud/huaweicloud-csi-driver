package utils

import (
	"crypto/md5"
	"fmt"
	"io"
	"sort"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func Md5SortMap(valueMap map[string]string) (string, error) {
	var keys []string
	for k := range valueMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var elements []string
	for _, k := range keys {
		element := fmt.Sprintf("%s=%s", k, valueMap[k])
		elements = append(elements, element)
	}
	original := strings.Join(elements, "&")
	hash := md5.New()
	if _, err := io.WriteString(hash, original); err != nil {
		return "", status.Errorf(codes.Internal, "Failed to write string, err: %v", err)
	}
	sum := hash.Sum(nil)
	return fmt.Sprintf("%x", sum), nil
}
