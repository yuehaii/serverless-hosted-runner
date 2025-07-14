package main

import (
	"os"
	"os/signal"
	agent "serverless-hosted-runner/agent"
	"syscall"

	kingpin "github.com/alecthomas/kingpin/v2"
	"github.com/ingka-group-digital/app-monitor-agent/logrus"
)

var (
	container_type    = kingpin.Flag("container_type", "container type.").Short('t').String()
	runner_id         = kingpin.Flag("runner_id", "runner id.").Short('i').String()
	runner_token      = kingpin.Flag("runner_token", "runner token.").Short('k').String()
	runner_repo_url   = kingpin.Flag("repo_url", "repo url.").Short('l').String()
	runner_org        = kingpin.Flag("repo_org", "repo orgnization.").Short('o').String()
	runner_repo_name  = kingpin.Flag("repo_name", "runner repo name.").Short('n').String()
	runner_action     = kingpin.Flag("runner_action", "runner action name.").Short('a').String()
	runner_repo_owner = kingpin.Flag("repo_owner", "runner repo owner.").Short('p').String()
	runner_labels     = kingpin.Flag("runner_labels", "runner labels.").Short('b').String()
	runner_group      = kingpin.Flag("runner_group", "runner group.").Short('g').String()
	image_ver         = kingpin.Flag("image_ver", "runner image version.").Short('v').String()
	ctx_log_level     = kingpin.Flag("ctx_log_level", "context log level.").Short('m').String()
	event_push        = kingpin.Flag("push_enable", "git webhook push enable.").Default("false").Short('w').Bool()
	cloud_pr          = kingpin.Flag("cloud_pr", "cloud provider.").Short('c').String()
	dis_ip            = kingpin.Flag("dis_ip", "dispacher ip.").Short('d').String()
)

func main() {
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	runner := EciRunnerCreator(*container_type, *runner_id, *runner_token, *runner_repo_url,
		*runner_org, *runner_repo_name, *runner_action, *runner_repo_owner, *image_ver,
		*runner_labels, *runner_group, *cloud_pr, *dis_ip)
	runner.Init()
	runner.Configure()
	if err := runner.Start(); err != nil {
		logrus.Errorf("fail to start runner, %v", err)
	}

	if *event_push {
		logrus.Infof("Monitoring state...")
		qAgent := agent.CreateAliMNSAgent(os.Getenv("TF_VAR_MNS_URL"), os.Getenv("ALICLOUD_ACCESS_KEY"),
			os.Getenv("ALICLOUD_SECRET_KEY"), agent.NOTIFICATION_Q, runner.Monitor, runner.Info())
		qAgent.MonitorOnAgent()
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	s := <-sigCh
	logrus.Infof("received a signal %s, complete shutdown the self-hosted runner", s)
}
