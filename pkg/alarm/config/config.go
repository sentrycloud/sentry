package config

import (
	"encoding/json"
	"fmt"
	"github.com/sentrycloud/sentry/pkg/dbmodel"
	"github.com/sentrycloud/sentry/pkg/newlog"
	"os"
	"strings"
)

const HttpProtocol = "http://"

type AlarmConfig struct {
	HttpPort      int                 `json:"http_port"`
	ProfilePort   int                 `json:"profile_port"`
	ServerAddress string              `json:"server_address"`
	Log           newlog.LogConfig    `json:"log"`
	MySQLServer   dbmodel.MySQLConfig `json:"mysql_server"`
	Mail          MailConfig          `json:"mail"`
}

type MailConfig struct {
	SmtpHost string `json:"smtp_host"`
	SmtpPort int    `json:"smtp_port"`
	From     string `json:"from"`
	Password string `json:"password"`
}

// TODO: missing config for sending phone message and IM message

func (ac *AlarmConfig) ParseConfig(configPath string) {
	content, err := os.ReadFile(configPath)
	if err != nil {
		fmt.Printf("read config file failed: %s\n", err)
		os.Exit(1)
	}

	err = json.Unmarshal(content, ac)
	if err != nil {
		fmt.Printf("unmarshal json for config file failed: %s\n", err)
		os.Exit(1)
	}

	if !strings.HasPrefix(ac.ServerAddress, HttpProtocol) {
		ac.ServerAddress = HttpProtocol + ac.ServerAddress
	}
}
