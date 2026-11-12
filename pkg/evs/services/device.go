package services

import (
	"fmt"
	"strings"
)

// nextFreeDevice returns the first unused name from /dev/vdb to /dev/vdz. /dev/sdX is treated as /dev/vdX.
func nextFreeDevice(used []string) (string, error) {
	taken := make(map[string]bool, len(used))
	for _, d := range used {
		taken[strings.Replace(d, "/dev/sd", "/dev/vd", 1)] = true
	}
	for ch := 'b'; ch <= 'z'; ch++ {
		if dev := "/dev/vd" + string(ch); !taken[dev] {
			return dev, nil
		}
	}
	return "", fmt.Errorf("no free /dev/vdX device name among %d attachments", len(used))
}
