package tsdb

import (
	"database/sql/driver"
	"errors"
	"github.com/sentrycloud/sentry/pkg/newlog"
	"github.com/sentrycloud/sentry/pkg/server/config"
	"github.com/sentrycloud/sentry/pkg/server/taos"
	"io"
)

// separate connection pool for query
var connPool *taos.ConnPool

func Init(taosServer config.TaosConfig) {
	connPool = taos.CreateConnPool(taosServer)
}

// QueryTSDB no reflection version, user need to parse value, but no need to open conn, query, parse each row
func QueryTSDB(sql string, totalColumn int) ([][]driver.Value, error) {
	conn, err := connPool.GetConn()
	if err != nil {
		errMsg := "get TSDB conn from pool failed: " + err.Error()
		newlog.Error(errMsg)
		return nil, errors.New(errMsg)
	}

	defer connPool.PutConn(conn)

	rows, err := conn.Query(sql)
	if err != nil {
		errMsg := "query TSDB failed: " + err.Error()
		newlog.Error(errMsg)
		return nil, errors.New(errMsg)
	}

	defer rows.Close()

	var result [][]driver.Value
	for {
		values := make([]driver.Value, totalColumn)
		if rows.Next(values) == io.EOF {
			break
		}

		result = append(result, values)
	}
	return result, nil
}
