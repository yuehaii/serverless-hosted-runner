package grpc

import (
	context "context"

	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

const _ = grpc.SupportPackageIsVersion9

const (
	RunnerListener_NotifyRunnerState_FullMethodName = "/grpc.RunnerListener/NotifyRunnerState"
)

type RunnerListenerClient interface {
	NotifyRunnerState(ctx context.Context, in *RunnerState, opts ...grpc.CallOption) (*ProcessState, error)
}

type runnerListenerClient struct {
	cc grpc.ClientConnInterface
}

func NewRunnerListenerClient(cc grpc.ClientConnInterface) RunnerListenerClient {
	return &runnerListenerClient{cc}
}

func (c *runnerListenerClient) NotifyRunnerState(ctx context.Context, in *RunnerState, opts ...grpc.CallOption) (*ProcessState, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(ProcessState)
	err := c.cc.Invoke(ctx, RunnerListener_NotifyRunnerState_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

type RunnerListenerServer interface {
	NotifyRunnerState(context.Context, *RunnerState) (*ProcessState, error)
	mustEmbedUnimplementedRunnerListenerServer()
}

type UnimplementedRunnerListenerServer struct{}

func (UnimplementedRunnerListenerServer) NotifyRunnerState(context.Context, *RunnerState) (*ProcessState, error) {
	return nil, status.Errorf(codes.Unimplemented, "method NotifyRunnerState not implemented")
}
func (UnimplementedRunnerListenerServer) mustEmbedUnimplementedRunnerListenerServer() {}
func (UnimplementedRunnerListenerServer) testEmbeddedByValue()                        {}

type UnsafeRunnerListenerServer interface {
	mustEmbedUnimplementedRunnerListenerServer()
}

func RegisterRunnerListenerServer(s grpc.ServiceRegistrar, srv RunnerListenerServer) {
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&RunnerListener_ServiceDesc, srv)
}

func _RunnerListener_NotifyRunnerState_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RunnerState)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RunnerListenerServer).NotifyRunnerState(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: RunnerListener_NotifyRunnerState_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RunnerListenerServer).NotifyRunnerState(ctx, req.(*RunnerState))
	}
	return interceptor(ctx, in, info, handler)
}

var RunnerListener_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "grpc.RunnerListener",
	HandlerType: (*RunnerListenerServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "NotifyRunnerState",
			Handler:    _RunnerListener_NotifyRunnerState_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "grpc/listener.proto",
}
