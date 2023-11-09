package sender

import "github.com/sentrycloud/sentry/pkg/newlog"

func WeChatMessage(to string, msg string) {
	newlog.Warn("send WeChat alarm message: %s to %s", msg, to)
}
