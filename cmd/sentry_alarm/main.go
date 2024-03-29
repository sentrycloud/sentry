package main

import (
	"github.com/sentrycloud/sentry/pkg/alarm/config"
	"github.com/sentrycloud/sentry/pkg/alarm/contact"
	"github.com/sentrycloud/sentry/pkg/alarm/query"
	"github.com/sentrycloud/sentry/pkg/alarm/rule"
	"github.com/sentrycloud/sentry/pkg/alarm/schedule"
	"github.com/sentrycloud/sentry/pkg/alarm/sender"
	"github.com/sentrycloud/sentry/pkg/cmdflags"
	"github.com/sentrycloud/sentry/pkg/dbmodel"
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

	sender.InitMailConfig(&alarmConfig.Mail)

	schedule.Start()
	defer schedule.Stop()

	err := dbmodel.NewMySQL(&alarmConfig.MySQLServer)
	if err != nil {
		return
	}

	contact.Init()

	ruleManager := rule.NewManager()
	ruleManager.Start()

	newlog.Info("sentry alarm server start complete in %d ms", time.Now().UnixMilli()-startTime)
	profile.StartProfileInBlockMode(alarmConfig.ProfilePort)
}
