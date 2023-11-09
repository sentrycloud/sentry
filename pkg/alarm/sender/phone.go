package sender

import "github.com/sentrycloud/sentry/pkg/newlog"

func PhoneMessage(to string, msg string) {
	newlog.Info("send phone alarm message: %s to %s", msg, to)
}
