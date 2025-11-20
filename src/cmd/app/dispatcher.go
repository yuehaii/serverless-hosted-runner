package app

import (
	"context"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	dispacher "serverless-hosted-runner/dispatcher"

	"github.com/ingka-group-digital/app-monitor-agent/logrus"
	"github.com/spf13/cobra"
)

type DispacherArgs struct {
	ListenPort         string
	RunnerPath         string
	ImageVersion       string
	CtxLogLevel        string
	LazyRegistrations  string
	EventPush          bool
	AllenRegistrations string
	PoolMode           bool
	CloudProvider      string
	TfController       string
}

var (
	disArgs      DispacherArgs
	dispacherCmd = &cobra.Command{
		Use:   "dispatcher",
		Short: "serverless dispatcher",
		FParseErrWhitelist: cobra.FParseErrWhitelist{
			UnknownFlags: true,
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			logrus.Printf("dispacher args. ListenPort:%v, RunnerPath:%v, ImageVersion:%v, CtxLogLevel:%v, LazyRegistrations:%v, EventPush:%v, AllenRegistrations:%v, PoolMode:%v, CloudProvider:%v, TfController:%v",
				disArgs.ListenPort, disArgs.RunnerPath, disArgs.ImageVersion, disArgs.CtxLogLevel, disArgs.LazyRegistrations, disArgs.EventPush, disArgs.AllenRegistrations, disArgs.PoolMode, disArgs.CloudProvider, disArgs.TfController)
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			dispatch := dispacher.EciDispatcherConstruct(disArgs.ImageVersion, disArgs.LazyRegistrations, disArgs.AllenRegistrations,
				disArgs.CtxLogLevel, disArgs.PoolMode, disArgs.CloudProvider, disArgs.EventPush, disArgs.TfController)
			dispatch.Init()
			http.HandleFunc(disArgs.RunnerPath, dispatch.HandleEvents)
			logrus.Infof("Listening on: %s", disArgs.ListenPort)
			server := &http.Server{
				Addr:         "0.0.0.0:" + disArgs.ListenPort,
				ReadTimeout:  600 * time.Second,
				WriteTimeout: 600 * time.Second,
			}
			go dispatch.Refresh()
			go func() {
				logrus.Infof("ListenAndServe: %v", server.ListenAndServe())
			}()
			sigCh := make(chan os.Signal, 1)
			signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
			s := <-sigCh
			logrus.Infof("received a signal %s, shutdown the server", s)
			stCtx, shCancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer shCancel()
			if err := server.Shutdown(stCtx); err != nil {
				logrus.Fatalf("server force shutdown. err: %s", err)
			}
			logrus.Infof("complete to shutdown the dispatcher server.")
		},
	}
)

func init() {
	disArgs = DispacherArgs{}
	dispacherCmd.PersistentFlags().StringVar(&disArgs.ListenPort, "addr", "61201", "listen on port")
	dispacherCmd.PersistentFlags().StringVar(&disArgs.RunnerPath, "path", "/runner", "runner path")
	dispacherCmd.PersistentFlags().StringVarP(&disArgs.ImageVersion, "image_ver", "v", "", "image version")
	dispacherCmd.PersistentFlags().StringVarP(&disArgs.CtxLogLevel, "ctx_log_level", "m", "", "context log level")
	dispacherCmd.PersistentFlags().StringVarP(&disArgs.LazyRegistrations, "lazy_regs", "r", "", "lazy install registrations")
	dispacherCmd.PersistentFlags().BoolVarP(&disArgs.EventPush, "push_enable", "w", false, "git webhook push enable")
	dispacherCmd.PersistentFlags().StringVarP(&disArgs.AllenRegistrations, "allen_regs", "a", "none", "allen registration enable")
	dispacherCmd.PersistentFlags().BoolVarP(&disArgs.PoolMode, "pool_enable", "p", false, "pool mode enabled")
	dispacherCmd.PersistentFlags().StringVarP(&disArgs.CloudProvider, "cloud", "c", "ali", "cloud provider")
	dispacherCmd.PersistentFlags().StringVarP(&disArgs.TfController, "tfctl", "t", "go", "terraform controller")
}
