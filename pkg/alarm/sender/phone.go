package sender

import "github.com/sentrycloud/sentry/pkg/newlog"

func PhoneMessage(msg string) {
	newlog.Warn("send phone alarm message: %s", msg)
}
