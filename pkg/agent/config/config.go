package config

import (
	"encoding/json"
	"fmt"
	"github.com/sentrycloud/sentry/pkg/newlog"
	"os"
)

type AgentConfig struct {
	TcpPort       int              `json:"tcp_port"`
	HttpPort      int              `json:"http_port"`
	ProfilePort   int              `json:"profile_port"`
	ServerAddress string           `json:"server_address"`
	Log           newlog.LogConfig `json:"log"`
	Scripts       []ScriptConfig   `json:"scripts"`
}

type ScriptConfig struct {
	ScriptPath string `json:"path"`
	ScriptType string `json:"type"`
}

func (c *AgentConfig) setDefault() {
	c.TcpPort = 50000
	c.HttpPort = 50001
	c.ProfilePort = 50002
	c.ServerAddress = "127.0.0.1:51000"

	c.Log.Path = "logs/sentryAgent.log"
	c.Log.Level = "info"
	c.Log.MaxSize = 100  // 100 MB
	c.Log.MaxBackup = 10 // 10 files
}

func (c *AgentConfig) Parse(configPath string) {
	c.setDefault()

	content, err := os.ReadFile(configPath)
	if err != nil {
		fmt.Printf("read config file failed: %s\n", err)
		os.Exit(1)
	}

	err = json.Unmarshal(content, c)
	if err != nil {
		fmt.Printf("unmarshal json for config file failed: %s\n", err)
		os.Exit(1)
	}
}
