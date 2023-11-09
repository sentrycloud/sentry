package sender

import (
	"fmt"
	"github.com/sentrycloud/sentry/pkg/alarm/config"
	"os"
	"strconv"
	"testing"
)

func TestMail(t *testing.T) {
	config := &config.MailConfig{}
	config.SmtpHost = os.Getenv("SMTP_HOST")
	config.SmtpPort, _ = strconv.Atoi(os.Getenv("SMTP_PORT"))
	config.From = os.Getenv("SMTP_USER")
	config.Password = os.Getenv("SMTP_PASSWORD")
	to := os.Getenv("SMTP_TO")

	fmt.Printf("host=%s, port=%d, from=%s, passwd=%s, to=%s\n", config.SmtpHost, config.SmtpPort, config.From, config.Password, to)
	InitMailConfig(config)

	MailMessage(to, "test msg")
}
