package main

import (
	"github.com/sentrycloud/sentry/pkg/agent/config"
	"github.com/sentrycloud/sentry/pkg/agent/httpcollector"
	"github.com/sentrycloud/sentry/pkg/agent/reporter"
	"github.com/sentrycloud/sentry/pkg/agent/script"
	"github.com/sentrycloud/sentry/pkg/agent/system"
	"github.com/sentrycloud/sentry/pkg/cmdflags"
	"github.com/sentrycloud/sentry/pkg/newlog"
	"github.com/sentrycloud/sentry/pkg/profile"
	"time"
)

func main() {
	startTime := time.Now().UnixMilli()

	// parse command parameters
	var cmdParams = cmdflags.CmdParams{}
	cmdParams.Parse("SentryAgent.conf")

	// parse config file
	var agentConfig = config.AgentConfig{}
	agentConfig.Parse(cmdParams.ConfigPath)

	// set log level, path, max file size and max file backups
	newlog.SetConfig(&agentConfig.Log)

	var agentReporter = &reporter.Reporter{}
	agentReporter.Start(agentConfig.ServerAddress)

	httpcollector.Start(agentReporter, agentConfig.HttpPort)

	for _, s := range agentConfig.Scripts {
		script.StartScriptScheduler(s.ScriptPath, s.ScriptType, agentReporter)
	}

	go system.CollectSystemMetric(agentReporter)

	newlog.Info("start complete in %d ms", time.Now().UnixMilli()-startTime)
	profile.StartProfileInBlockMode(agentConfig.ProfilePort)
}
