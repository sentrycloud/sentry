package sender

import "github.com/sentrycloud/sentry/pkg/newlog"

func MailMessage(msg string) {
	newlog.Warn("send mail alarm message: %s", msg)
}
