package config

import (
	"encoding/json"
	"fmt"
	"github.com/sentrycloud/sentry/pkg/newlog"
	"os"
	"strings"
)

const HttpProtocol = "http://"

type AlarmConfig struct {
	HttpPort      int              `json:"http_port"`
	ProfilePort   int              `json:"profile_port"`
	ServerAddress string           `json:"server_address"`
	Log           newlog.LogConfig `json:"log"`
	Db            MySQLConfig      `json:"db"`
}

type MySQLConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	DBName   string `json:"db_name"`
}

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
