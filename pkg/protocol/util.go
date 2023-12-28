package protocol

import (
	"errors"
	"github.com/sentrycloud/sentry/pkg/newlog"
	"strconv"
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

func TransferDownSample(downSample string) (int64, error) {
	downSample = strings.TrimSpace(downSample)
	downSample = strings.ToLower(downSample)
	size := len(downSample)
	if size == 0 {
		return 0, errors.New("empty downSample")
	}

	unit := downSample[size-1]
	if unit >= '0' && unit <= '9' {
		// no unit, default to second
		return strconv.ParseInt(downSample, 10, 64)
	}

	multiply := 1
	switch unit {
	case 's':
		multiply = 1
	case 'm':
		multiply = 60
	case 'h':
		multiply = 3600
	case 'd':
		multiply = 24 * 3600
	default:
		return 0, errors.New("invalid downSample unit")
	}

	base, err := strconv.ParseInt(downSample[0:size-1], 10, 64)
	if err != nil {
		return 0, err
	}

	return base * int64(multiply), nil
}
