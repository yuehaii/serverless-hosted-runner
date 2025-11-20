package grpc

import (
	reflect "reflect"
	sync "sync"
	unsafe "unsafe"

	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
)

const (
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type RunnerState struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	RunnerID      *string                `protobuf:"bytes,1,opt,name=runnerId" json:"runnerId,omitempty"`
	State         *string                `protobuf:"bytes,2,opt,name=state" json:"state,omitempty"`
	StateMsg      *string                `protobuf:"bytes,3,opt,name=stateMsg" json:"stateMsg,omitempty"`
	Act           *string                `protobuf:"bytes,4,opt,name=act" json:"act,omitempty"`
	RunerName     *string                `protobuf:"bytes,5,opt,name=runer_name,json=runerName" json:"runer_name,omitempty"`
	RepoName      *string                `protobuf:"bytes,6,opt,name=repo_name,json=repoName" json:"repo_name,omitempty"`
	OrgName       *string                `protobuf:"bytes,7,opt,name=org_name,json=orgName" json:"org_name,omitempty"`
	RunWf         *string                `protobuf:"bytes,8,opt,name=run_wf,json=runWf" json:"run_wf,omitempty"`
	Labels        *string                `protobuf:"bytes,9,opt,name=labels" json:"labels,omitempty"`
	URL           *string                `protobuf:"bytes,10,opt,name=url" json:"url,omitempty"`
	Owner         *string                `protobuf:"bytes,11,opt,name=owner" json:"owner,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *RunnerState) Reset() {
	*x = RunnerState{}
	mi := &fileGrpcListenerProtoMsgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *RunnerState) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RunnerState) ProtoMessage() {}

func (x *RunnerState) ProtoReflect() protoreflect.Message {
	mi := &fileGrpcListenerProtoMsgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*RunnerState) Descriptor() ([]byte, []int) {
	return fileGrpcListenerProtoRawDescGZIP(), []int{0}
}

func (x *RunnerState) GetRunnerID() string {
	if x != nil && x.RunnerID != nil {
		return *x.RunnerID
	}
	return ""
}

func (x *RunnerState) GetState() string {
	if x != nil && x.State != nil {
		return *x.State
	}
	return ""
}

func (x *RunnerState) GetStateMsg() string {
	if x != nil && x.StateMsg != nil {
		return *x.StateMsg
	}
	return ""
}

func (x *RunnerState) GetAct() string {
	if x != nil && x.Act != nil {
		return *x.Act
	}
	return ""
}

func (x *RunnerState) GetRunerName() string {
	if x != nil && x.RunerName != nil {
		return *x.RunerName
	}
	return ""
}

func (x *RunnerState) GetRepoName() string {
	if x != nil && x.RepoName != nil {
		return *x.RepoName
	}
	return ""
}

func (x *RunnerState) GetOrgName() string {
	if x != nil && x.OrgName != nil {
		return *x.OrgName
	}
	return ""
}

func (x *RunnerState) GetRunWf() string {
	if x != nil && x.RunWf != nil {
		return *x.RunWf
	}
	return ""
}

func (x *RunnerState) GetLabels() string {
	if x != nil && x.Labels != nil {
		return *x.Labels
	}
	return ""
}

func (x *RunnerState) GetURL() string {
	if x != nil && x.URL != nil {
		return *x.URL
	}
	return ""
}

func (x *RunnerState) GetOwner() string {
	if x != nil && x.Owner != nil {
		return *x.Owner
	}
	return ""
}

type ProcessState struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	State         *bool                  `protobuf:"varint,1,opt,name=state" json:"state,omitempty"`
	StateMsg      *string                `protobuf:"bytes,2,opt,name=stateMsg" json:"stateMsg,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ProcessState) Reset() {
	*x = ProcessState{}
	mi := &fileGrpcListenerProtoMsgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ProcessState) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ProcessState) ProtoMessage() {}

func (x *ProcessState) ProtoReflect() protoreflect.Message {
	mi := &fileGrpcListenerProtoMsgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*ProcessState) Descriptor() ([]byte, []int) {
	return fileGrpcListenerProtoRawDescGZIP(), []int{1}
}

func (x *ProcessState) GetState() bool {
	if x != nil && x.State != nil {
		return *x.State
	}
	return false
}

func (x *ProcessState) GetStateMsg() string {
	if x != nil && x.StateMsg != nil {
		return *x.StateMsg
	}
	return ""
}

var FileGrpcListenerProto protoreflect.FileDescriptor

const fileGrpcListenerProtoRawDesc = "" +
	"\n" +
	"\x13grpc/listener.proto\x12\x04grpc\"\x9b\x02\n" +
	"\vRunnerState\x12\x1a\n" +
	"\brunnerId\x18\x01 \x01(\tR\brunnerId\x12\x14\n" +
	"\x05state\x18\x02 \x01(\tR\x05state\x12\x1a\n" +
	"\bstateMsg\x18\x03 \x01(\tR\bstateMsg\x12\x10\n" +
	"\x03act\x18\x04 \x01(\tR\x03act\x12\x1d\n" +
	"\n" +
	"runer_name\x18\x05 \x01(\tR\trunerName\x12\x1b\n" +
	"\trepo_name\x18\x06 \x01(\tR\brepoName\x12\x19\n" +
	"\borg_name\x18\a \x01(\tR\aorgName\x12\x15\n" +
	"\x06run_wf\x18\b \x01(\tR\x05runWf\x12\x16\n" +
	"\x06labels\x18\t \x01(\tR\x06labels\x12\x10\n" +
	"\x03url\x18\n" +
	" \x01(\tR\x03url\x12\x14\n" +
	"\x05owner\x18\v \x01(\tR\x05owner\"@\n" +
	"\fProcessState\x12\x14\n" +
	"\x05state\x18\x01 \x01(\bR\x05state\x12\x1a\n" +
	"\bstateMsg\x18\x02 \x01(\tR\bstateMsg2N\n" +
	"\x0eRunnerListener\x12<\n" +
	"\x11NotifyRunnerState\x12\x11.grpc.RunnerState\x1a\x12.grpc.ProcessState\"\x00B,Z*serverless-hosted-runner/src/network/grpc/b\beditionsp\xe8\a"

var (
	fileGrpcListenerProtoRawDescOnce sync.Once
	fileGrpcListenerProtoRawDescData []byte
)

func fileGrpcListenerProtoRawDescGZIP() []byte {
	fileGrpcListenerProtoRawDescOnce.Do(func() {
		fileGrpcListenerProtoRawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(fileGrpcListenerProtoRawDesc), len(fileGrpcListenerProtoRawDesc)))
	})
	return fileGrpcListenerProtoRawDescData
}

var fileGrpcListenerProtoMsgTypes = make([]protoimpl.MessageInfo, 2)
var fileGrpcListenerProtoGoTypes = []any{
	(*RunnerState)(nil),
	(*ProcessState)(nil),
}
var fileGrpcListenerProtoDepIdxs = []int32{
	0,
	1,
	1,
	0,
	0,
	0,
	0,
}

func init() { fileGrpcListenerProtoInit() }
func fileGrpcListenerProtoInit() {
	if FileGrpcListenerProto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(fileGrpcListenerProtoRawDesc), len(fileGrpcListenerProtoRawDesc)),
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           fileGrpcListenerProtoGoTypes,
		DependencyIndexes: fileGrpcListenerProtoDepIdxs,
		MessageInfos:      fileGrpcListenerProtoMsgTypes,
	}.Build()
	FileGrpcListenerProto = out.File
	fileGrpcListenerProtoGoTypes = nil
	fileGrpcListenerProtoDepIdxs = nil
}
