package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
)

func Sha256(valueMap map[string]string) string {
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
	src := strings.Join(elements, "&")

	m := sha256.New()
	m.Write([]byte(src))
	res := hex.EncodeToString(m.Sum(nil))
	return res
}
