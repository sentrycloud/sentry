package tsdb

import (
	"fmt"
	"github.com/sentrycloud/sentry/pkg/server/config"
	"testing"
)

func TestWrapper(t *testing.T) {
	var taosServer = config.TaosConfig{
		Host:     "127.0.0.1",
		Port:     6030,
		User:     "sentry",
		Password: "123456",
		Database: "sentry",
	}

	Init(taosServer)

	// query metric
	result, err := QueryTSDB("show stables", 1)
	if err != nil {
		fmt.Println("Query failed: " + err.Error())
		return
	}

	fmt.Println("metrics: ")
	for _, row := range result {
		fmt.Println(row[0].(string))
	}
}
