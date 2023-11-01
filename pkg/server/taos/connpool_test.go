package taos

import (
	"fmt"
	"sentry/internal/server/config"
	"testing"
	"time"
)

func TestSchemalessWrite(t *testing.T) {
	var taosServer = config.TaosConfig{
		Host:     "127.0.0.1",
		Port:     6030,
		User:     "sentry",
		Password: "123456",
		Database: "sentry",
	}

	ts := time.Now().Unix()
	var payload = fmt.Sprintf("[{\"metric\":\"sentry_test_metric\", \"tags\":{\"machine\":\"pod\"}, \"timestamp\":%d,\"value\":1}]", ts)
	var pool = CreateConnPool(taosServer)
	err := pool.SchemalessWrite(payload)
	if err != nil {
		fmt.Printf("schemaless write failed: %v\n", err)
	} else {
		fmt.Println("schemaless write success")
	}
}
