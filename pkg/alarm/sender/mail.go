package sender

import (
	"fmt"
	"github.com/sentrycloud/sentry/pkg/alarm/config"
	"github.com/sentrycloud/sentry/pkg/newlog"
	"net/smtp"
)

var (
	mailConfig *config.MailConfig
	auth       smtp.Auth
	msgFormat  string
	smtpAddr   string
)

func InitMailConfig(config *config.MailConfig) {
	mailConfig = config
	auth = smtp.PlainAuth("", mailConfig.From, mailConfig.Password, mailConfig.SmtpHost)
	// The msg parameter should be an RFC 822-style email
	msgFormat = "To: %s\r\nFrom: sentry <" + mailConfig.From + ">\r\nSubject: sentry alarm message\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s"
	smtpAddr = fmt.Sprintf("%s:%d", mailConfig.SmtpHost, mailConfig.SmtpPort)
}

func MailMessage(to string, msg string) {
	newlog.Info("send mail alarm message: %s to %s", msg, to)

	msgBody := fmt.Sprintf(msgFormat, to, msg)
	err := smtp.SendMail(smtpAddr, auth, mailConfig.From, []string{to}, []byte(msgBody))
	if err != nil {
		newlog.Error("send mail failed: %v", err)
	}
}
