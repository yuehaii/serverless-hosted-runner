package main

import (
	"context"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/ingka-group-digital/app-monitor-agent/logrus"
)

var (
	listenAddress = kingpin.Flag("addr", "listen on port").Default("61201").String()
	runnerPath    = kingpin.Flag("path", "runner path").Default("/runner").String()
	image_ver     = kingpin.Flag("image_ver", "image version.").Short('v').String()
	ctx_log_level = kingpin.Flag("ctx_log_level", "context log level.").Short('m').String()
	lazy_regs     = kingpin.Flag("lazy_regs", "lazy install registrations.").Short('r').String()
	event_push    = kingpin.Flag("push_enable", "git webhook push enable.").Default("false").Short('w').Bool()
	allen_regs    = kingpin.Flag("allen_regs", "allen registration enable.").Default("none").Short('a').String()
	pool_mode     = kingpin.Flag("pool_enable", "pool mode enabled.").Default("false").Short('p').Bool()
	cloud_pr      = kingpin.Flag("cloud", "cloud provider").Default("ali").Short('c').String()
)

func main() {
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()
	dispatch := EciDispatcherConstruct(*image_ver, *lazy_regs, *allen_regs)
	http.HandleFunc(*runnerPath, dispatch.HandleEvents)
	logrus.Infof("Listening on: %s", *listenAddress)
	server := &http.Server{
		Addr:         "0.0.0.0:" + *listenAddress,
		ReadTimeout:  600 * time.Second,
		WriteTimeout: 600 * time.Second,
		// MaxHeaderBytes: 1 << 20,
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
}
