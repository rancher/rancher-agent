package main

import (
	"context"
	"os"
	"strings"

	"github.com/rancher/system-agent/pkg/version"

	"github.com/rancher/system-agent/pkg/applyinator"
	"github.com/rancher/system-agent/pkg/config"
	"github.com/rancher/system-agent/pkg/localplan"
	"github.com/rancher/system-agent/pkg/remoteplan"
	"github.com/rancher/wrangler/pkg/signals"

	"github.com/mattn/go-colorable"
	"github.com/sirupsen/logrus"
)

func main() {
	logrus.SetOutput(colorable.NewColorableStdout())

	rawLevel := os.Getenv("CATTLE_LOGLEVEL")

	if rawLevel != "" {
		if lvl, err := logrus.ParseLevel(os.Getenv("CATTLE_LOGLEVEL")); err != nil {
			logrus.Fatal(err)
		} else {
			logrus.SetLevel(lvl)
		}
	}

	err := run()

	if err != nil {
		logrus.Fatal(err)
	}

}

func isTrue(input string) bool {
	if strings.ToLower(input) == "true" || input == "1" {
		return true
	}
	return false
}

func run() error {
	topContext := signals.SetupSignalHandler(context.Background())

	logrus.Infof("Rancher System Agent version %s is starting", version.FriendlyVersion())

	configFile := os.Getenv("CATTLE_AGENT_CONFIG")

	if configFile == "" {
		configFile = "/etc/rancher/agent/config.yaml"
	}

	var cf config.AgentConfig

	err := config.Parse(configFile, &cf)
	if err != nil {
		logrus.Fatalf("Unable to parse config file %v", err)
	}

	logrus.Infof("Using directory %s for work", cf.WorkDir)

	applyinator := applyinator.NewApplyinator(cf.WorkDir, "")

	if cf.RemoteEnabled {
		logrus.Infof("Starting remote watch of plans")

		var connInfo config.ConnectionInfo

		err = config.Parse(cf.ConnectionInfoFile, &connInfo)

		if err != nil {
			logrus.Fatalf("Unable to parse connection info file %v", err)
		}

		remoteplan.Watch(topContext, *applyinator, connInfo)
	}

	logrus.Infof("Starting local watch of plans in %s", cf.LocalPlanDir)
	localplan.WatchFiles(topContext, *applyinator, cf.LocalPlanDir)

	<-topContext.Done()
	return nil
}
