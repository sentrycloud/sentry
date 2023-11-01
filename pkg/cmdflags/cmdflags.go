package cmdflags

import (
	"flag"
	"fmt"
	"github.com/sentrycloud/sentry/pkg"
	"os"
)

type CmdParams struct {
	Version    bool
	Help       bool
	ConfigPath string
}

func (c *CmdParams) Parse(defaultConfigPath string) {
	version := flag.Bool("v", false, "show version")
	help := flag.Bool("h", false, "show usage")
	configPath := flag.String("c", defaultConfigPath, "config file path")
	flag.Parse()

	c.Version = *version
	c.Help = *help
	c.ConfigPath = *configPath

	if c.Version {
		fmt.Printf("version: %s\n", pkg.Version)
		os.Exit(0)
	}

	if c.Help {
		fmt.Println("usage: ")
		fmt.Println("  -v :show version")
		fmt.Println("  -h :show help")
		fmt.Println("  -c configPath :set the configuration file path")
		os.Exit(0)
	}
}
