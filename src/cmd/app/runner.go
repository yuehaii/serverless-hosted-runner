package app

import (
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"

	"serverless-hosted-runner/agent"
	runner "serverless-hosted-runner/runner"

	"github.com/ingka-group-digital/app-monitor-agent/logrus"
	"github.com/spf13/cobra"
)

type RunnerArgs struct {
	ContainerType   string
	RunnerID        string
	RunnerToken     string
	RunnerRepoURL   string
	RunnerOrg       string
	RunnerRepoName  string
	RunnerAction    string
	RunnerRepoOwner string
	RunnerLabels    string
	RunnerGroup     string
	ImageVer        string
	CtxLogLevel     string
	EventPush       bool
	CloudPr         string
	DisIP           string
	RepoRegToken    string
}

var (
	runnerArgs RunnerArgs
	runnerCmd  = &cobra.Command{
		Use:   "runner",
		Short: "serverless runner",
		FParseErrWhitelist: cobra.FParseErrWhitelist{
			UnknownFlags: true,
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			logrus.Printf("runner args. ContainerType:%v, RunnerID:%v, RunnerToken:%v, CtxLogLevel:%v, RunnerRepoURL:%v, RunnerOrg:%v, RunnerRepoName:%v, RunnerLabels:%v, RunnerGroup:%v,ImageVer:%v, CtxLogLevel:%v, EventPush:%v, CloudPr:%v, DisIP:%v, RepoRegToken:%v",
				runnerArgs.ContainerType, runnerArgs.RunnerID, runnerArgs.RunnerToken, runnerArgs.CtxLogLevel, runnerArgs.RunnerRepoURL,
				runnerArgs.RunnerOrg, runnerArgs.RunnerRepoName, runnerArgs.RunnerLabels, runnerArgs.RunnerGroup, runnerArgs.ImageVer,
				runnerArgs.CtxLogLevel, runnerArgs.EventPush, runnerArgs.CloudPr, runnerArgs.DisIP, runnerArgs.RepoRegToken)
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			runner := runner.EciRunnerCreator(runnerArgs.ContainerType, runnerArgs.RunnerID, runnerArgs.RunnerToken, runnerArgs.RunnerRepoURL,
				runnerArgs.RunnerOrg, runnerArgs.RunnerRepoName, runnerArgs.RunnerAction, runnerArgs.RunnerRepoOwner, runnerArgs.ImageVer,
				runnerArgs.RunnerLabels, runnerArgs.RunnerGroup, runnerArgs.CloudPr, runnerArgs.DisIP, runnerArgs.CtxLogLevel, runnerArgs.RepoRegToken)
			runner.Init()
			runner.Configure()
			if err := runner.Start(); err != nil {
				logrus.Errorf("fail to start runner, %v", err)
			}

			if runnerArgs.EventPush {
				logrus.Infof("Monitoring state...")
				qAgent := agent.CreateAliMNSAgent(os.Getenv("TF_VAR_MNS_URL"), os.Getenv("ALICLOUD_ACCESS_KEY"),
					os.Getenv("ALICLOUD_SECRET_KEY"), agent.NotificationQueue, runner.Monitor, runner.Info())
				qAgent.MonitorOnAgent()
			}

			sigCh := make(chan os.Signal, 1)
			signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
			s := <-sigCh
			logrus.Infof("received a signal %s, complete shutdown the self-hosted runner", s)
		},
	}
)

func init() {
	runnerArgs = RunnerArgs{}
	runnerCmd.PersistentFlags().StringVarP(&runnerArgs.ContainerType, "container_type", "t", "", "container type")
	runnerCmd.PersistentFlags().StringVarP(&runnerArgs.RunnerID, "runner_id", "i", "", "runner id")
	runnerCmd.PersistentFlags().StringVarP(&runnerArgs.RunnerToken, "runner_token", "k", "", "runner token")
	runnerCmd.PersistentFlags().StringVarP(&runnerArgs.RunnerRepoURL, "repo_url", "l", "", "repo url")
	runnerCmd.PersistentFlags().StringVarP(&runnerArgs.RunnerOrg, "repo_org", "o", "", "repo orgnization")
	runnerCmd.PersistentFlags().StringVarP(&runnerArgs.RunnerRepoName, "repo_name", "n", "", "runner repo name")
	runnerCmd.PersistentFlags().StringVarP(&runnerArgs.RunnerAction, "runner_action", "a", "", "runner action name")
	runnerCmd.PersistentFlags().StringVarP(&runnerArgs.RunnerRepoOwner, "repo_owner", "p", "", "runner repo owner")
	runnerCmd.PersistentFlags().StringVarP(&runnerArgs.RunnerLabels, "runner_labels", "b", "", "runner labels")
	runnerCmd.PersistentFlags().StringVarP(&runnerArgs.RunnerGroup, "runner_group", "g", "", "runner group")
	runnerCmd.PersistentFlags().StringVarP(&runnerArgs.ImageVer, "image_ver", "v", "", "runner image version")
	runnerCmd.PersistentFlags().StringVarP(&runnerArgs.CtxLogLevel, "ctx_log_level", "m", "", "context log level")
	runnerCmd.PersistentFlags().BoolVarP(&runnerArgs.EventPush, "push_enable", "w", false, "git webhook push enable")
	runnerCmd.PersistentFlags().StringVarP(&runnerArgs.CloudPr, "cloud_pr", "c", "", "cloud provider")
	runnerCmd.PersistentFlags().StringVarP(&runnerArgs.DisIP, "dis_ip", "d", "", "dispacher ip")
	runnerCmd.PersistentFlags().StringVarP(&runnerArgs.RepoRegToken, "repo_reg_tk", "r", "", "repo registration token")
}
