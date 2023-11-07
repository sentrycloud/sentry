package taos

import (
	"fmt"
	"github.com/sentrycloud/sentry/pkg/newlog"
	"github.com/sentrycloud/sentry/pkg/server/config"
	"github.com/taosdata/driver-go/v3/af"
	"sync"
	"sync/atomic"
)

type ConnPool struct {
	TaosServer config.TaosConfig
	PrepareSQL string
	ConnCount  int32
	mu         sync.Mutex
	connList   []*af.Connector
}

func CreateConnPool(taosServer config.TaosConfig) *ConnPool {
	var connPool = &ConnPool{}
	connPool.TaosServer = taosServer
	connPool.ConnCount = 0
	connPool.PrepareSQL = "use " + taosServer.Database
	return connPool
}

func (p *ConnPool) GetConn() (*af.Connector, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	var conn *af.Connector
	if len(p.connList) > 0 {
		conn = p.connList[0]
		p.connList = p.connList[1:]
		return conn, nil
	} else {
		var err error
		conn, err = af.Open(p.TaosServer.Host, p.TaosServer.User, p.TaosServer.Password, p.TaosServer.Database, p.TaosServer.Port)
		if err != nil {
			newlog.Error("open taos connection failed: %v", err)
			return nil, fmt.Errorf("open taos connection failed: %v", err)
		}

		atomic.AddInt32(&p.ConnCount, 1)
		newlog.Info("open taos connection, totalCount=%d", atomic.LoadInt32(&p.ConnCount))

		_, err = conn.Exec(p.PrepareSQL)
		if err != nil {
			newlog.Error("exec prepare SQL failed: %v", err)
		}
		return conn, nil
	}
}

func (p *ConnPool) PutConn(conn *af.Connector) {
	p.mu.Lock()
	p.connList = append(p.connList, conn)
	p.mu.Unlock()
}

func (p *ConnPool) SchemalessWrite(payload string) error {
	conn, err := p.GetConn()
	if err != nil {
		return err
	}

	err = conn.OpenTSDBInsertJsonPayload(payload)
	// if insert success put the connection back into the connection poll, otherwise do not push back and close the connection
	if err == nil {
		p.PutConn(conn)
	} else {
		_ = conn.Close()
		atomic.AddInt32(&p.ConnCount, -1)
		newlog.Info("close broken taos connection, totalCount=%d", atomic.LoadInt32(&p.ConnCount))
	}

	return err
}
