package main

import (
	"github.com/sentrycloud/sentry/pkg/cmdflags"
	"github.com/sentrycloud/sentry/pkg/newlog"
	"github.com/sentrycloud/sentry/pkg/profile"
	"github.com/sentrycloud/sentry/pkg/server/collector"
	"github.com/sentrycloud/sentry/pkg/server/config"
	"github.com/sentrycloud/sentry/pkg/server/http"
	"github.com/sentrycloud/sentry/pkg/server/merge"
	"github.com/sentrycloud/sentry/pkg/server/taos"
	"time"
)

func main() {
	startTime := time.Now().UnixMilli()

	// parse command parameters
	var cmdParams = cmdflags.CmdParams{}
	cmdParams.Parse("SentryServer.conf")

	// parse config file
	var serverConfig = config.ServerConfig{}
	serverConfig.Parse(cmdParams.ConfigPath)

	// set log level, path, max file size and max file backups
	newlog.SetConfig(&serverConfig.Log)

	// create time series database connection pool
	var connPool = taos.CreateConnPool(serverConfig.TaosServer)
	// crate merger to send all payload in batch mode
	var merger = merge.CreateMerge(serverConfig.Merge, connPool)
	merger.Start()

	// start the server
	var server = collector.Collector{}
	server.Start(serverConfig, merger)

	http.Start(&serverConfig, &server)

	newlog.Info("sentry server start complete in %d ms", time.Now().UnixMilli()-startTime)
	profile.StartProfileInBlockMode(serverConfig.ProfilePort)
}
