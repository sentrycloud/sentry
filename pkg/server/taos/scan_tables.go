package taos

import (
	"database/sql/driver"
	"fmt"
	"github.com/sentrycloud/sentry/pkg/newlog"
	"io"
	"time"
)

const (
	ScanTableInterval    = 24 * 3600 // scan all tables once everyday
	BestScanStartHour    = 6         // start scan at 6am every day, I guess that is low peak time of most business
	DaysToDeleteOldTable = 10        // delete tables that are not updated for these days
	ShowStablesSql       = "SHOW stables;"
	OldTableSqlFormat    = "select * from (select last_row(_ts) as last_ts,tbname from `%s` group by tbname) where last_ts < now - %dd;"
	DropTableSqlFormat   = "DROP TABLE IF EXISTS %s;"
)

var scanConnPool *ConnPool

// StartScanTables used to scan all tables and delete too old tables that not updated for days
func StartScanTables(pool *ConnPool) {
	newlog.Info("StartScanTables")
	scanConnPool = pool

	sleepDuration := 0 * time.Second
	now := time.Now()
	if now.Hour() > BestScanStartHour {
		tomorrowBestHour := time.Date(now.Year(), now.Month(), now.Day()+1, BestScanStartHour, 0, 0, 0, now.Location())
		sleepDuration = tomorrowBestHour.Sub(now)
	} else if now.Hour() < BestScanStartHour {
		todayBestHour := time.Date(now.Year(), now.Month(), now.Day(), BestScanStartHour, 0, 0, 0, now.Location())
		sleepDuration = todayBestHour.Sub(now)
	}

	if sleepDuration > 0 {
		newlog.Warn("not start in the best hour, sleep until next %dam", BestScanStartHour)
		time.Sleep(sleepDuration)
	}

	// time.NewTicker is not very exactly, use sleep
	for {
		scanStartTick := time.Now().Unix()
		scanAllTables()
		scanTotalTime := time.Now().Unix() - scanStartTick
		time.Sleep(time.Duration(ScanTableInterval-scanTotalTime) * time.Second)
	}
}

func scanAllTables() {
	newlog.Info("start scanAllTables")
	startTime := time.Now().Unix()

	metrics, err := queryTables(ShowStablesSql, 1)
	if err != nil {
		return
	}

	totalDropTableCount := 0
	for index, metric := range metrics {
		totalDropTableCount += deleteOldTable(metric)

		// log the process
		if index%100 == 0 {
			newlog.Info("scan to index=%d, currentDropTableCount=%d", index, totalDropTableCount)
		}
	}

	newlog.Info("scanAllTable complete in %d second, totalDropTableCount=%d", time.Now().Unix()-startTime, totalDropTableCount)
}

func queryTables(sql string, columnCount int) ([]string, error) {
	conn, err := scanConnPool.GetConn()
	if err != nil {
		newlog.Error("get taos conn failed: %v", err)
		return nil, err
	}

	defer scanConnPool.PutConn(conn)
	rows, err := conn.Query(sql)
	if err != nil {
		newlog.Error("%s failed: %v", sql, err)
		return nil, err
	}

	defer rows.Close()
	var tables []string
	for {
		values := make([]driver.Value, columnCount) // return stable_name for metric or last_ts and tbname for old tables
		if rows.Next(values) == io.EOF {
			break
		}

		table := values[columnCount-1].(string)
		tables = append(tables, table)
	}
	return tables, nil
}

func deleteOldTable(metric string) int {
	oldTableSql := fmt.Sprintf(OldTableSqlFormat, metric, DaysToDeleteOldTable)
	oldTables, err := queryTables(oldTableSql, 2)
	if err != nil {
		return 0
	}

	if len(oldTables) == 0 {
		return 0
	}

	conn, err := scanConnPool.GetConn()
	if err != nil {
		newlog.Error("get taos conn failed: %v", err)
		return 0
	}

	defer scanConnPool.PutConn(conn)

	dropTableCount := 0
	for _, tableName := range oldTables {
		dropTableSql := fmt.Sprintf(DropTableSqlFormat, tableName)
		_, err = conn.Exec(dropTableSql)
		if err != nil {
			newlog.Error("drop metric=%s failed: %v", metric, err)
		} else {
			dropTableCount++
		}
	}

	return dropTableCount
}
