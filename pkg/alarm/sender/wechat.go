package sender

import "github.com/sentrycloud/sentry/pkg/newlog"

func WeChatMessage(msg string) {
	newlog.Warn("send WeChat alarm message: %s", msg)
}
