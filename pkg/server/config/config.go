package config

import (
	"encoding/json"
	"fmt"
	"github.com/sentrycloud/sentry/pkg/dbmodel"
	"github.com/sentrycloud/sentry/pkg/newlog"
	"os"
)

type TaosConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Database string `json:"database"`
}

type ServerConfig struct {
	TcpPort      int                 `json:"tcp_port"`
	HttpPort     int                 `json:"http_port"`
	ProfilePort  int                 `json:"profile_port"`
	MaxConnCount int                 `json:"max_conn_count"`
	ScanTable    bool                `json:"scan_table"`
	Log          newlog.LogConfig    `json:"log"`
	TaosServer   TaosConfig          `json:"taos_server"`
	MySQLServer  dbmodel.MySQLConfig `json:"mysql_server"`
	Merge        MergeConfig         `json:"merge"`
}

type MergeConfig struct {
	ChanSize         int `json:"chan_size"`
	PayloadMaxSize   int `json:"payload_max_size"`
	PayloadBatchSize int `json:"payload_batch_size"`
	TickInterval     int `json:"tick_interval"`
}

func (c *ServerConfig) setDefault() {
	c.TcpPort = 51000
	c.HttpPort = 51001
	c.ProfilePort = 51002
	c.MaxConnCount = 1000
	c.ScanTable = false

	c.Log.Path = "logs/sentryServer.log"
	c.Log.Level = "info"
	c.Log.MaxSize = 200  // MB
	c.Log.MaxBackup = 10 // 10 files

	c.TaosServer.Host = "127.0.0.1"
	c.TaosServer.Port = 6030
	c.TaosServer.User = "sentry"
	c.TaosServer.Password = "123456"
	c.TaosServer.Database = "sentry"

	c.Merge.ChanSize = 20000
	c.Merge.PayloadMaxSize = 600 * 1024 * 1024 // 600 MB
	c.Merge.PayloadBatchSize = 600 * 1024      // 600 KB
	c.Merge.TickInterval = 5
}

func (c *ServerConfig) Parse(configPath string) {
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
