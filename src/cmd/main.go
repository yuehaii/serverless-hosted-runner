package main

import (
	"os"
	app "serverless-hosted-runner/cmd/app"

	"github.com/ingka-group-digital/app-monitor-agent/logrus"
)

func main() {
	cmd := app.NewRootCmd()
	if err := cmd.Execute(); err != nil {
		logrus.Errorf("fail to execute root cmd, %v", err)
		os.Exit(-1)
	}
}
