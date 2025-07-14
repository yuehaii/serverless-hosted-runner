package grpc

import (
	context "context"
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"net"
	"path/filepath"
	"runtime"
	common "serverless-hosted-runner/common"
	"strings"
	"time"

	"github.com/ingka-group-digital/app-monitor-agent/logrus"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

type INotifier interface {
	Notify(RunnerState)
}

type IListener interface {
	RunnerListenerServer
	Start()
}

type GrpcServer struct {
	addr      string
	port      string
	tls_agent common.Cryptography
}

type Listener struct {
	GrpcServer
}

type Notifier struct {
	GrpcServer
}

type GrpcListener struct {
	Listener
	Finished common.DestroyRunner
}

type GrpcNotifier struct {
	Notifier
}

func CreateListener(cb common.DestroyRunner, listener_addr string) IListener {
	return createGrpcListener(cb, listener_addr)
}

func CreateNotifier(addr string) INotifier {
	return createGrpcNotifier(addr)
}

func createGrpcListener(cb common.DestroyRunner, listener_addr string) IListener {
	return &GrpcListener{Listener{GrpcServer{"localhost", "9090", common.RSACryptography(listener_addr)}}, cb}
}

func createGrpcNotifier(addr string) INotifier {
	return &GrpcNotifier{Notifier{GrpcServer{addr, "9090", common.RSACryptography("")}}}
}

func (ntf *GrpcServer) absolutePath(path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	_, cur_f, _, _ := runtime.Caller(0)
	base_path := filepath.Dir(cur_f)
	return filepath.Join(base_path, path)
}

func (ntf *GrpcNotifier) Notify(state RunnerState) {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTimeout(20*time.Second))
	ca_cert := ntf.tls_agent.GetCertificate(false, true)
	logrus.Infof("ca_cert is %s", ca_cert)
	if len(ca_cert) > 0 {
		creds, err := credentials.NewClientTLSFromFile(ca_cert, "")
		// creds, err := credentials.NewClientTLSFromFile(ntf.absolutePath(ca_cert), "")
		if err != nil {
			logrus.Errorf("Failed to create TLS credentials: %v", err)
		}
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}
	conn, err := grpc.NewClient(fmt.Sprintf("%s:%s", ntf.addr, ntf.port), opts...)
	if err != nil {
		logrus.Warnf("GrpcNotifier Notify, fail to dial %s:%s: %v", ntf.addr, ntf.port, err)
	}
	defer conn.Close()
	client := NewRunnerListenerClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	process_state, err := client.NotifyRunnerState(ctx, &state)
	if err == nil {
		logrus.Warnf("NotifyRunnerState process_state, <%v:%s>, <%s> ", process_state.GetState(),
			process_state.GetStateMsg(), process_state.String())
	} else {
		logrus.Errorf("NotifyRunnerState err %v", err)
	}
}

func (lis *GrpcListener) NotifyRunnerState(ctx context.Context, state *RunnerState) (*ProcessState, error) {
	logrus.Infof("NotifyRunnerState grpc func start")
	p_state := false
	p_msg := "processing"
	if *state.State == "Finished" {
		logrus.Warnf("NotifyRunnerState paras: act %s, runer_name %s, repo_name %s, org_name %s, run_wf %s, labels %v, url %s, owner %s",
			*state.Act, *state.RunerName, *state.RepoName, *state.OrgName, *state.RunWf,
			strings.Split(*state.Labels, ","), *state.Url, *state.Owner)
		lis.Finished(*state.Act, *state.RunerName, *state.RepoName, *state.OrgName, *state.RunWf,
			strings.Split(*state.Labels, ","), *state.Url, *state.Owner)
		p_state = true
		p_msg = "done"
	}
	return &ProcessState{State: &p_state, StateMsg: &p_msg}, nil
}

func (lis *GrpcListener) mustEmbedUnimplementedRunnerListenerServer() {
	return
}

func (lis *GrpcListener) startGrpcServer() {
	lisen, err := net.Listen("tcp", fmt.Sprintf(":%s", lis.port))
	if err != nil {
		logrus.Warnf("GrpcListener Start, failed to listen: %v", err)
	}
	var opts []grpc.ServerOption
	server_cert := lis.tls_agent.GetCertificate(false, false)
	server_key := lis.tls_agent.GetCertificate(true, false)
	logrus.Infof("server_cert is %s, server_key is %s", server_cert, server_key)
	if len(server_cert) > 0 && len(server_key) > 0 {
		creds, err := credentials.NewServerTLSFromFile(server_cert, server_key)
		// creds, err := credentials.NewServerTLSFromFile(lis.absolutePath(server_cert), lis.absolutePath(server_key))
		if err != nil {
			logrus.Errorf("Failed to generate credentials: %v", err)
		}
		opts = []grpc.ServerOption{grpc.Creds(creds)}
	}
	grpcServer := grpc.NewServer(opts...)
	RegisterRunnerListenerServer(grpcServer, lis)
	logrus.Infof("Grpc server start. Listening on %s:%s", lis.addr, lis.port)
	grpcServer.Serve(lisen)
}

func (lis *GrpcListener) configTls() error {
	err, ca_cert := lis.tls_agent.LoadCertificate(lis.tls_agent.GetCertificate(false, true), false)
	if err != nil {
		logrus.Errorf("Fail to load ca certificate: %v", err)
		return err
	}
	err, ca_key := lis.tls_agent.LoadCertificate(lis.tls_agent.GetCertificate(true, true), true)
	if err != nil {
		logrus.Errorf("Fail to load ca key: %v", err)
		return err
	}
	err, _, _ = lis.tls_agent.GenCertificate(ca_cert.(*x509.Certificate), ca_key.(*rsa.PrivateKey))
	if err != nil {
		logrus.Errorf("Fail to generate server certificates: %v", err)
		return err
	}
	return nil
}

func (lis *GrpcListener) Start() {
	lis.configTls()
	lis.startGrpcServer()
}
