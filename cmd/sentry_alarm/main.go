package main

import (
	"github.com/sentrycloud/sentry/pkg/alarm/config"
	"github.com/sentrycloud/sentry/pkg/alarm/mysql"
	"github.com/sentrycloud/sentry/pkg/alarm/query"
	"github.com/sentrycloud/sentry/pkg/alarm/rule"
	"github.com/sentrycloud/sentry/pkg/alarm/schedule"
	"github.com/sentrycloud/sentry/pkg/cmdflags"
	"github.com/sentrycloud/sentry/pkg/newlog"
	"github.com/sentrycloud/sentry/pkg/profile"
	"time"
)

func main() {
	startTime := time.Now().UnixMilli()

	// parse command parameters
	var cmdParams = cmdflags.CmdParams{}
	cmdParams.Parse("SentryAlarm.conf")

	// parse config file
	var alarmConfig = &config.AlarmConfig{}
	alarmConfig.ParseConfig(cmdParams.ConfigPath)

	// set log level, path, max file size and max file backups
	newlog.SetConfig(&alarmConfig.Log)

	query.InitServerAddr(alarmConfig.ServerAddress)

	schedule.Start()
	defer schedule.Stop()

	db, err := mysql.NewMySQL(&alarmConfig.Db)
	if err != nil {
		return
	}

	ruleManager := rule.NewManager(db)
	ruleManager.Start()

	// parse config file
	newlog.Info("sentry alarm server start complete in %d ms", time.Now().UnixMilli()-startTime)
	profile.StartProfileInBlockMode(alarmConfig.ProfilePort)
}
