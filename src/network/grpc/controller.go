// Package grpc service between dispacher and its runners
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
	Notify(*RunnerState)
}

type IListener interface {
	RunnerListenerServer
	Start()
}

type GrpcServer struct {
	addr     string
	port     string
	tlsAgent common.Cryptography
	absPath  bool
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

func CreateListener(cb common.DestroyRunner, listenerAddr string) IListener {
	return createGrpcListener(cb, listenerAddr)
}

func CreateNotifier(addr string) INotifier {
	return createGrpcNotifier(addr)
}

func createGrpcListener(cb common.DestroyRunner, listenerAddr string) IListener {
	return &GrpcListener{Listener{GrpcServer{"localhost", "9090", common.RSACryptography(listenerAddr), false}}, cb}
}

func createGrpcNotifier(addr string) INotifier {
	return &GrpcNotifier{Notifier{GrpcServer{addr, "9090", common.RSACryptography(""), false}}}
}

func (ntf *GrpcServer) absolutePath(path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	_, curFile, _, _ := runtime.Caller(0)
	basePath := filepath.Dir(curFile)
	return filepath.Join(basePath, path)
}

func (ntf *GrpcNotifier) Notify(state *RunnerState) {
	var opts []grpc.DialOption
	transCred := insecure.NewCredentials()
	opts = append(opts, grpc.WithTimeout(20*time.Second))
	caCert := ntf.tlsAgent.GetCertificate(false, true)
	logrus.Infof("caCert is %s", caCert)
	if len(caCert) > 0 {
		if ntf.absPath {
			caCert = ntf.absolutePath(caCert)
		}
		creds, err := credentials.NewClientTLSFromFile(caCert, "")
		if err != nil {
			logrus.Errorf("failed to create TLS credentials: %v, use default", err)
		} else {
			transCred = creds
		}
	}
	opts = append(opts, grpc.WithTransportCredentials(transCred))

	conn, err := grpc.NewClient(fmt.Sprintf("%s:%s", ntf.addr, ntf.port), opts...)
	if err != nil {
		logrus.Warnf("grpcNotifier Notify, fail to dial %s:%s: %v", ntf.addr, ntf.port, err)
	}
	connClose := func() {
		if err := conn.Close(); err != nil {
			logrus.Errorf("notify client, fail to close connection, err %v", err)
		}
	}
	defer connClose()
	client := NewRunnerListenerClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	processState, err := client.NotifyRunnerState(ctx, state)
	if err == nil {
		logrus.Warnf("notifyRunnerState processState, <%v:%s>, <%s> ", processState.GetState(),
			processState.GetStateMsg(), processState.String())
	} else {
		logrus.Errorf("notifyRunnerState err %v", err)
	}
}

func (lis *GrpcListener) NotifyRunnerState(ctx context.Context, state *RunnerState) (*ProcessState, error) {
	logrus.Infof("NotifyRunnerState grpc func start")
	pState := false
	pMsg := "processing"
	if *state.State == "Finished" {
		logrus.Warnf("NotifyRunnerState paras: act %s, runer_name %s, repo_name %s, org_name %s, run_wf %s, labels %v, url %s, owner %s",
			*state.Act, *state.RunerName, *state.RepoName, *state.OrgName, *state.RunWf,
			strings.Split(*state.Labels, ","), *state.URL, *state.Owner)
		lis.Finished(*state.Act, *state.RunerName, *state.RepoName, *state.OrgName, *state.RunWf,
			strings.Split(*state.Labels, ","), *state.URL, *state.Owner)
		pState = true
		pMsg = "done"
	}
	return &ProcessState{State: &pState, StateMsg: &pMsg}, nil
}

func (lis *GrpcListener) mustEmbedUnimplementedRunnerListenerServer() {

}

func (lis *GrpcListener) startGrpcServer() {
	lisen, err := net.Listen("tcp", fmt.Sprintf(":%s", lis.port))
	if err != nil {
		logrus.Warnf("grpcListener Start, failed to listen: %v", err)
	}
	var opts []grpc.ServerOption
	serverCert := lis.tlsAgent.GetCertificate(false, false)
	serverKey := lis.tlsAgent.GetCertificate(true, false)
	logrus.Infof("serverCert is %s, serverKey is %s", serverCert, serverKey)
	if len(serverCert) > 0 && len(serverKey) > 0 {
		if lis.absPath {
			serverCert = lis.absolutePath(serverCert)
			serverKey = lis.absolutePath(serverKey)
		}
		creds, err := credentials.NewServerTLSFromFile(serverCert, serverKey)
		if err != nil {
			logrus.Errorf("failed to generate credentials: %v", err)
		}
		opts = []grpc.ServerOption{grpc.Creds(creds)}
	}
	grpcServer := grpc.NewServer(opts...)
	RegisterRunnerListenerServer(grpcServer, lis)
	logrus.Infof("grpc server start. Listening on :%s", lis.port)
	if err := grpcServer.Serve(lisen); err != nil {
		logrus.Errorf("failed to start gRPC server: %v", err)
	}
}

func (lis *GrpcListener) configTLS() error {
	caCert, err := lis.tlsAgent.LoadCertificate(lis.tlsAgent.GetCertificate(false, true), false)
	if err != nil {
		logrus.Errorf("fail to load ca certificate: %v", err)
		return err
	} else {
		switch caCert.(type) {
		default:
			logrus.Errorf("load ca cert with incorrect type")
			return fmt.Errorf("load cert with incorrect type")
		case *x509.Certificate:
			logrus.Infof("load x509 ca cert")
		}
	}
	caKey, err := lis.tlsAgent.LoadCertificate(lis.tlsAgent.GetCertificate(true, true), true)
	if err != nil {
		logrus.Errorf("fail to load ca key: %v", err)
		return err
	} else {
		switch caKey.(type) {
		default:
			logrus.Errorf("load ca key with incorrect type")
			return fmt.Errorf("load ca key with incorrect type")
		case *rsa.PrivateKey:
			logrus.Infof("Load rsa ca key")
		}
	}
	_, _, err = lis.tlsAgent.GenCertificate(caCert.(*x509.Certificate), caKey.(*rsa.PrivateKey))
	if err != nil {
		logrus.Errorf("fail to generate server certificates: %v", err)
		return err
	}
	return nil
}

func (lis *GrpcListener) Start() {
	if err := lis.configTLS(); err != nil {
		logrus.Errorf("fail to config grpc server tls: %v", err)
	}
	lis.startGrpcServer()
}
