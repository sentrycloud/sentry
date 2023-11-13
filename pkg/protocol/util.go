package protocol

import (
	"github.com/sentrycloud/sentry/pkg/newlog"
	"strings"
)

// GetIPFromConnAddr get ip from the following format:
// ipv4 format: "192.0.2.1:25", ipv6 format: "[2001:db8::1]:80"
func GetIPFromConnAddr(addr string) string {
	var ip string
	index := strings.LastIndex(addr, ":")
	if index != -1 {
		if addr[0] != '[' {
			ip = addr[0:index]
		} else {
			ip = addr[1 : index-1]
		}
	} else {
		newlog.Error("address=%s is not valid", addr) // error return empty ip
	}
	return ip
}
