// Code generated by protoc-gen-go. DO NOT EDIT.
// source: ssh.proto

package gitalypb // import "gitlab.com/gitlab-org/gitaly-proto/go/gitalypb"

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type SSHUploadPackRequest struct {
	// 'repository' must be present in the first message.
	Repository *Repository `protobuf:"bytes,1,opt,name=repository,proto3" json:"repository,omitempty"`
	// A chunk of raw data to be copied to 'git upload-pack' standard input
	Stdin []byte `protobuf:"bytes,2,opt,name=stdin,proto3" json:"stdin,omitempty"`
	// Parameters to use with git -c (key=value pairs)
	GitConfigOptions []string `protobuf:"bytes,4,rep,name=git_config_options,json=gitConfigOptions,proto3" json:"git_config_options,omitempty"`
	// Git protocol version
	GitProtocol          string   `protobuf:"bytes,5,opt,name=git_protocol,json=gitProtocol,proto3" json:"git_protocol,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *SSHUploadPackRequest) Reset()         { *m = SSHUploadPackRequest{} }
func (m *SSHUploadPackRequest) String() string { return proto.CompactTextString(m) }
func (*SSHUploadPackRequest) ProtoMessage()    {}
func (*SSHUploadPackRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_ssh_7bc71c1984deb95b, []int{0}
}
func (m *SSHUploadPackRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SSHUploadPackRequest.Unmarshal(m, b)
}
func (m *SSHUploadPackRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SSHUploadPackRequest.Marshal(b, m, deterministic)
}
func (dst *SSHUploadPackRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SSHUploadPackRequest.Merge(dst, src)
}
func (m *SSHUploadPackRequest) XXX_Size() int {
	return xxx_messageInfo_SSHUploadPackRequest.Size(m)
}
func (m *SSHUploadPackRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_SSHUploadPackRequest.DiscardUnknown(m)
}

var xxx_messageInfo_SSHUploadPackRequest proto.InternalMessageInfo

func (m *SSHUploadPackRequest) GetRepository() *Repository {
	if m != nil {
		return m.Repository
	}
	return nil
}

func (m *SSHUploadPackRequest) GetStdin() []byte {
	if m != nil {
		return m.Stdin
	}
	return nil
}

func (m *SSHUploadPackRequest) GetGitConfigOptions() []string {
	if m != nil {
		return m.GitConfigOptions
	}
	return nil
}

func (m *SSHUploadPackRequest) GetGitProtocol() string {
	if m != nil {
		return m.GitProtocol
	}
	return ""
}

type SSHUploadPackResponse struct {
	// A chunk of raw data from 'git upload-pack' standard output
	Stdout []byte `protobuf:"bytes,1,opt,name=stdout,proto3" json:"stdout,omitempty"`
	// A chunk of raw data from 'git upload-pack' standard error
	Stderr []byte `protobuf:"bytes,2,opt,name=stderr,proto3" json:"stderr,omitempty"`
	// This field may be nil. This is intentional: only when the remote
	// command has finished can we return its exit status.
	ExitStatus           *ExitStatus `protobuf:"bytes,3,opt,name=exit_status,json=exitStatus,proto3" json:"exit_status,omitempty"`
	XXX_NoUnkeyedLiteral struct{}    `json:"-"`
	XXX_unrecognized     []byte      `json:"-"`
	XXX_sizecache        int32       `json:"-"`
}

func (m *SSHUploadPackResponse) Reset()         { *m = SSHUploadPackResponse{} }
func (m *SSHUploadPackResponse) String() string { return proto.CompactTextString(m) }
func (*SSHUploadPackResponse) ProtoMessage()    {}
func (*SSHUploadPackResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_ssh_7bc71c1984deb95b, []int{1}
}
func (m *SSHUploadPackResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SSHUploadPackResponse.Unmarshal(m, b)
}
func (m *SSHUploadPackResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SSHUploadPackResponse.Marshal(b, m, deterministic)
}
func (dst *SSHUploadPackResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SSHUploadPackResponse.Merge(dst, src)
}
func (m *SSHUploadPackResponse) XXX_Size() int {
	return xxx_messageInfo_SSHUploadPackResponse.Size(m)
}
func (m *SSHUploadPackResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_SSHUploadPackResponse.DiscardUnknown(m)
}

var xxx_messageInfo_SSHUploadPackResponse proto.InternalMessageInfo

func (m *SSHUploadPackResponse) GetStdout() []byte {
	if m != nil {
		return m.Stdout
	}
	return nil
}

func (m *SSHUploadPackResponse) GetStderr() []byte {
	if m != nil {
		return m.Stderr
	}
	return nil
}

func (m *SSHUploadPackResponse) GetExitStatus() *ExitStatus {
	if m != nil {
		return m.ExitStatus
	}
	return nil
}

type SSHReceivePackRequest struct {
	// 'repository' must be present in the first message.
	Repository *Repository `protobuf:"bytes,1,opt,name=repository,proto3" json:"repository,omitempty"`
	// A chunk of raw data to be copied to 'git upload-pack' standard input
	Stdin []byte `protobuf:"bytes,2,opt,name=stdin,proto3" json:"stdin,omitempty"`
	// Contents of GL_ID, GL_REPOSITORY, and GL_USERNAME environment variables
	// for 'git receive-pack'
	GlId         string `protobuf:"bytes,3,opt,name=gl_id,json=glId,proto3" json:"gl_id,omitempty"`
	GlRepository string `protobuf:"bytes,4,opt,name=gl_repository,json=glRepository,proto3" json:"gl_repository,omitempty"`
	GlUsername   string `protobuf:"bytes,5,opt,name=gl_username,json=glUsername,proto3" json:"gl_username,omitempty"`
	// Git protocol version
	GitProtocol string `protobuf:"bytes,6,opt,name=git_protocol,json=gitProtocol,proto3" json:"git_protocol,omitempty"`
	// Parameters to use with git -c (key=value pairs)
	GitConfigOptions     []string `protobuf:"bytes,7,rep,name=git_config_options,json=gitConfigOptions,proto3" json:"git_config_options,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *SSHReceivePackRequest) Reset()         { *m = SSHReceivePackRequest{} }
func (m *SSHReceivePackRequest) String() string { return proto.CompactTextString(m) }
func (*SSHReceivePackRequest) ProtoMessage()    {}
func (*SSHReceivePackRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_ssh_7bc71c1984deb95b, []int{2}
}
func (m *SSHReceivePackRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SSHReceivePackRequest.Unmarshal(m, b)
}
func (m *SSHReceivePackRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SSHReceivePackRequest.Marshal(b, m, deterministic)
}
func (dst *SSHReceivePackRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SSHReceivePackRequest.Merge(dst, src)
}
func (m *SSHReceivePackRequest) XXX_Size() int {
	return xxx_messageInfo_SSHReceivePackRequest.Size(m)
}
func (m *SSHReceivePackRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_SSHReceivePackRequest.DiscardUnknown(m)
}

var xxx_messageInfo_SSHReceivePackRequest proto.InternalMessageInfo

func (m *SSHReceivePackRequest) GetRepository() *Repository {
	if m != nil {
		return m.Repository
	}
	return nil
}

func (m *SSHReceivePackRequest) GetStdin() []byte {
	if m != nil {
		return m.Stdin
	}
	return nil
}

func (m *SSHReceivePackRequest) GetGlId() string {
	if m != nil {
		return m.GlId
	}
	return ""
}

func (m *SSHReceivePackRequest) GetGlRepository() string {
	if m != nil {
		return m.GlRepository
	}
	return ""
}

func (m *SSHReceivePackRequest) GetGlUsername() string {
	if m != nil {
		return m.GlUsername
	}
	return ""
}

func (m *SSHReceivePackRequest) GetGitProtocol() string {
	if m != nil {
		return m.GitProtocol
	}
	return ""
}

func (m *SSHReceivePackRequest) GetGitConfigOptions() []string {
	if m != nil {
		return m.GitConfigOptions
	}
	return nil
}

type SSHReceivePackResponse struct {
	// A chunk of raw data from 'git receive-pack' standard output
	Stdout []byte `protobuf:"bytes,1,opt,name=stdout,proto3" json:"stdout,omitempty"`
	// A chunk of raw data from 'git receive-pack' standard error
	Stderr []byte `protobuf:"bytes,2,opt,name=stderr,proto3" json:"stderr,omitempty"`
	// This field may be nil. This is intentional: only when the remote
	// command has finished can we return its exit status.
	ExitStatus           *ExitStatus `protobuf:"bytes,3,opt,name=exit_status,json=exitStatus,proto3" json:"exit_status,omitempty"`
	XXX_NoUnkeyedLiteral struct{}    `json:"-"`
	XXX_unrecognized     []byte      `json:"-"`
	XXX_sizecache        int32       `json:"-"`
}

func (m *SSHReceivePackResponse) Reset()         { *m = SSHReceivePackResponse{} }
func (m *SSHReceivePackResponse) String() string { return proto.CompactTextString(m) }
func (*SSHReceivePackResponse) ProtoMessage()    {}
func (*SSHReceivePackResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_ssh_7bc71c1984deb95b, []int{3}
}
func (m *SSHReceivePackResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SSHReceivePackResponse.Unmarshal(m, b)
}
func (m *SSHReceivePackResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SSHReceivePackResponse.Marshal(b, m, deterministic)
}
func (dst *SSHReceivePackResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SSHReceivePackResponse.Merge(dst, src)
}
func (m *SSHReceivePackResponse) XXX_Size() int {
	return xxx_messageInfo_SSHReceivePackResponse.Size(m)
}
func (m *SSHReceivePackResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_SSHReceivePackResponse.DiscardUnknown(m)
}

var xxx_messageInfo_SSHReceivePackResponse proto.InternalMessageInfo

func (m *SSHReceivePackResponse) GetStdout() []byte {
	if m != nil {
		return m.Stdout
	}
	return nil
}

func (m *SSHReceivePackResponse) GetStderr() []byte {
	if m != nil {
		return m.Stderr
	}
	return nil
}

func (m *SSHReceivePackResponse) GetExitStatus() *ExitStatus {
	if m != nil {
		return m.ExitStatus
	}
	return nil
}

type SSHUploadArchiveRequest struct {
	// 'repository' must be present in the first message.
	Repository *Repository `protobuf:"bytes,1,opt,name=repository,proto3" json:"repository,omitempty"`
	// A chunk of raw data to be copied to 'git upload-archive' standard input
	Stdin                []byte   `protobuf:"bytes,2,opt,name=stdin,proto3" json:"stdin,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *SSHUploadArchiveRequest) Reset()         { *m = SSHUploadArchiveRequest{} }
func (m *SSHUploadArchiveRequest) String() string { return proto.CompactTextString(m) }
func (*SSHUploadArchiveRequest) ProtoMessage()    {}
func (*SSHUploadArchiveRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_ssh_7bc71c1984deb95b, []int{4}
}
func (m *SSHUploadArchiveRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SSHUploadArchiveRequest.Unmarshal(m, b)
}
func (m *SSHUploadArchiveRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SSHUploadArchiveRequest.Marshal(b, m, deterministic)
}
func (dst *SSHUploadArchiveRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SSHUploadArchiveRequest.Merge(dst, src)
}
func (m *SSHUploadArchiveRequest) XXX_Size() int {
	return xxx_messageInfo_SSHUploadArchiveRequest.Size(m)
}
func (m *SSHUploadArchiveRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_SSHUploadArchiveRequest.DiscardUnknown(m)
}

var xxx_messageInfo_SSHUploadArchiveRequest proto.InternalMessageInfo

func (m *SSHUploadArchiveRequest) GetRepository() *Repository {
	if m != nil {
		return m.Repository
	}
	return nil
}

func (m *SSHUploadArchiveRequest) GetStdin() []byte {
	if m != nil {
		return m.Stdin
	}
	return nil
}

type SSHUploadArchiveResponse struct {
	// A chunk of raw data from 'git upload-archive' standard output
	Stdout []byte `protobuf:"bytes,1,opt,name=stdout,proto3" json:"stdout,omitempty"`
	// A chunk of raw data from 'git upload-archive' standard error
	Stderr []byte `protobuf:"bytes,2,opt,name=stderr,proto3" json:"stderr,omitempty"`
	// This value will only be set on the last message
	ExitStatus           *ExitStatus `protobuf:"bytes,3,opt,name=exit_status,json=exitStatus,proto3" json:"exit_status,omitempty"`
	XXX_NoUnkeyedLiteral struct{}    `json:"-"`
	XXX_unrecognized     []byte      `json:"-"`
	XXX_sizecache        int32       `json:"-"`
}

func (m *SSHUploadArchiveResponse) Reset()         { *m = SSHUploadArchiveResponse{} }
func (m *SSHUploadArchiveResponse) String() string { return proto.CompactTextString(m) }
func (*SSHUploadArchiveResponse) ProtoMessage()    {}
func (*SSHUploadArchiveResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_ssh_7bc71c1984deb95b, []int{5}
}
func (m *SSHUploadArchiveResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SSHUploadArchiveResponse.Unmarshal(m, b)
}
func (m *SSHUploadArchiveResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SSHUploadArchiveResponse.Marshal(b, m, deterministic)
}
func (dst *SSHUploadArchiveResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SSHUploadArchiveResponse.Merge(dst, src)
}
func (m *SSHUploadArchiveResponse) XXX_Size() int {
	return xxx_messageInfo_SSHUploadArchiveResponse.Size(m)
}
func (m *SSHUploadArchiveResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_SSHUploadArchiveResponse.DiscardUnknown(m)
}

var xxx_messageInfo_SSHUploadArchiveResponse proto.InternalMessageInfo

func (m *SSHUploadArchiveResponse) GetStdout() []byte {
	if m != nil {
		return m.Stdout
	}
	return nil
}

func (m *SSHUploadArchiveResponse) GetStderr() []byte {
	if m != nil {
		return m.Stderr
	}
	return nil
}

func (m *SSHUploadArchiveResponse) GetExitStatus() *ExitStatus {
	if m != nil {
		return m.ExitStatus
	}
	return nil
}

func init() {
	proto.RegisterType((*SSHUploadPackRequest)(nil), "gitaly.SSHUploadPackRequest")
	proto.RegisterType((*SSHUploadPackResponse)(nil), "gitaly.SSHUploadPackResponse")
	proto.RegisterType((*SSHReceivePackRequest)(nil), "gitaly.SSHReceivePackRequest")
	proto.RegisterType((*SSHReceivePackResponse)(nil), "gitaly.SSHReceivePackResponse")
	proto.RegisterType((*SSHUploadArchiveRequest)(nil), "gitaly.SSHUploadArchiveRequest")
	proto.RegisterType((*SSHUploadArchiveResponse)(nil), "gitaly.SSHUploadArchiveResponse")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// SSHServiceClient is the client API for SSHService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type SSHServiceClient interface {
	// To forward 'git upload-pack' to Gitaly for SSH sessions
	SSHUploadPack(ctx context.Context, opts ...grpc.CallOption) (SSHService_SSHUploadPackClient, error)
	// To forward 'git receive-pack' to Gitaly for SSH sessions
	SSHReceivePack(ctx context.Context, opts ...grpc.CallOption) (SSHService_SSHReceivePackClient, error)
	// To forward 'git upload-archive' to Gitaly for SSH sessions
	SSHUploadArchive(ctx context.Context, opts ...grpc.CallOption) (SSHService_SSHUploadArchiveClient, error)
}

type sSHServiceClient struct {
	cc *grpc.ClientConn
}

func NewSSHServiceClient(cc *grpc.ClientConn) SSHServiceClient {
	return &sSHServiceClient{cc}
}

func (c *sSHServiceClient) SSHUploadPack(ctx context.Context, opts ...grpc.CallOption) (SSHService_SSHUploadPackClient, error) {
	stream, err := c.cc.NewStream(ctx, &_SSHService_serviceDesc.Streams[0], "/gitaly.SSHService/SSHUploadPack", opts...)
	if err != nil {
		return nil, err
	}
	x := &sSHServiceSSHUploadPackClient{stream}
	return x, nil
}

type SSHService_SSHUploadPackClient interface {
	Send(*SSHUploadPackRequest) error
	Recv() (*SSHUploadPackResponse, error)
	grpc.ClientStream
}

type sSHServiceSSHUploadPackClient struct {
	grpc.ClientStream
}

func (x *sSHServiceSSHUploadPackClient) Send(m *SSHUploadPackRequest) error {
	return x.ClientStream.SendMsg(m)
}

func (x *sSHServiceSSHUploadPackClient) Recv() (*SSHUploadPackResponse, error) {
	m := new(SSHUploadPackResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *sSHServiceClient) SSHReceivePack(ctx context.Context, opts ...grpc.CallOption) (SSHService_SSHReceivePackClient, error) {
	stream, err := c.cc.NewStream(ctx, &_SSHService_serviceDesc.Streams[1], "/gitaly.SSHService/SSHReceivePack", opts...)
	if err != nil {
		return nil, err
	}
	x := &sSHServiceSSHReceivePackClient{stream}
	return x, nil
}

type SSHService_SSHReceivePackClient interface {
	Send(*SSHReceivePackRequest) error
	Recv() (*SSHReceivePackResponse, error)
	grpc.ClientStream
}

type sSHServiceSSHReceivePackClient struct {
	grpc.ClientStream
}

func (x *sSHServiceSSHReceivePackClient) Send(m *SSHReceivePackRequest) error {
	return x.ClientStream.SendMsg(m)
}

func (x *sSHServiceSSHReceivePackClient) Recv() (*SSHReceivePackResponse, error) {
	m := new(SSHReceivePackResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *sSHServiceClient) SSHUploadArchive(ctx context.Context, opts ...grpc.CallOption) (SSHService_SSHUploadArchiveClient, error) {
	stream, err := c.cc.NewStream(ctx, &_SSHService_serviceDesc.Streams[2], "/gitaly.SSHService/SSHUploadArchive", opts...)
	if err != nil {
		return nil, err
	}
	x := &sSHServiceSSHUploadArchiveClient{stream}
	return x, nil
}

type SSHService_SSHUploadArchiveClient interface {
	Send(*SSHUploadArchiveRequest) error
	Recv() (*SSHUploadArchiveResponse, error)
	grpc.ClientStream
}

type sSHServiceSSHUploadArchiveClient struct {
	grpc.ClientStream
}

func (x *sSHServiceSSHUploadArchiveClient) Send(m *SSHUploadArchiveRequest) error {
	return x.ClientStream.SendMsg(m)
}

func (x *sSHServiceSSHUploadArchiveClient) Recv() (*SSHUploadArchiveResponse, error) {
	m := new(SSHUploadArchiveResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// SSHServiceServer is the server API for SSHService service.
type SSHServiceServer interface {
	// To forward 'git upload-pack' to Gitaly for SSH sessions
	SSHUploadPack(SSHService_SSHUploadPackServer) error
	// To forward 'git receive-pack' to Gitaly for SSH sessions
	SSHReceivePack(SSHService_SSHReceivePackServer) error
	// To forward 'git upload-archive' to Gitaly for SSH sessions
	SSHUploadArchive(SSHService_SSHUploadArchiveServer) error
}

func RegisterSSHServiceServer(s *grpc.Server, srv SSHServiceServer) {
	s.RegisterService(&_SSHService_serviceDesc, srv)
}

func _SSHService_SSHUploadPack_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(SSHServiceServer).SSHUploadPack(&sSHServiceSSHUploadPackServer{stream})
}

type SSHService_SSHUploadPackServer interface {
	Send(*SSHUploadPackResponse) error
	Recv() (*SSHUploadPackRequest, error)
	grpc.ServerStream
}

type sSHServiceSSHUploadPackServer struct {
	grpc.ServerStream
}

func (x *sSHServiceSSHUploadPackServer) Send(m *SSHUploadPackResponse) error {
	return x.ServerStream.SendMsg(m)
}

func (x *sSHServiceSSHUploadPackServer) Recv() (*SSHUploadPackRequest, error) {
	m := new(SSHUploadPackRequest)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _SSHService_SSHReceivePack_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(SSHServiceServer).SSHReceivePack(&sSHServiceSSHReceivePackServer{stream})
}

type SSHService_SSHReceivePackServer interface {
	Send(*SSHReceivePackResponse) error
	Recv() (*SSHReceivePackRequest, error)
	grpc.ServerStream
}

type sSHServiceSSHReceivePackServer struct {
	grpc.ServerStream
}

func (x *sSHServiceSSHReceivePackServer) Send(m *SSHReceivePackResponse) error {
	return x.ServerStream.SendMsg(m)
}

func (x *sSHServiceSSHReceivePackServer) Recv() (*SSHReceivePackRequest, error) {
	m := new(SSHReceivePackRequest)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _SSHService_SSHUploadArchive_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(SSHServiceServer).SSHUploadArchive(&sSHServiceSSHUploadArchiveServer{stream})
}

type SSHService_SSHUploadArchiveServer interface {
	Send(*SSHUploadArchiveResponse) error
	Recv() (*SSHUploadArchiveRequest, error)
	grpc.ServerStream
}

type sSHServiceSSHUploadArchiveServer struct {
	grpc.ServerStream
}

func (x *sSHServiceSSHUploadArchiveServer) Send(m *SSHUploadArchiveResponse) error {
	return x.ServerStream.SendMsg(m)
}

func (x *sSHServiceSSHUploadArchiveServer) Recv() (*SSHUploadArchiveRequest, error) {
	m := new(SSHUploadArchiveRequest)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

var _SSHService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "gitaly.SSHService",
	HandlerType: (*SSHServiceServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "SSHUploadPack",
			Handler:       _SSHService_SSHUploadPack_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
		{
			StreamName:    "SSHReceivePack",
			Handler:       _SSHService_SSHReceivePack_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
		{
			StreamName:    "SSHUploadArchive",
			Handler:       _SSHService_SSHUploadArchive_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "ssh.proto",
}

func init() { proto.RegisterFile("ssh.proto", fileDescriptor_ssh_7bc71c1984deb95b) }

var fileDescriptor_ssh_7bc71c1984deb95b = []byte{
	// 492 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xbc, 0x53, 0xd1, 0x6e, 0xd3, 0x30,
	0x14, 0x55, 0xba, 0x34, 0xac, 0xb7, 0x19, 0xaa, 0xcc, 0x36, 0xa2, 0x0a, 0x58, 0x08, 0x2f, 0x79,
	0x60, 0xed, 0xd4, 0x7d, 0x01, 0x20, 0xa4, 0xc1, 0x0b, 0x93, 0xa3, 0x49, 0x08, 0x1e, 0x22, 0x37,
	0x31, 0xae, 0x85, 0x1b, 0x07, 0xdb, 0xad, 0x36, 0x09, 0xc4, 0x17, 0xf0, 0xcc, 0x1f, 0xf0, 0x39,
	0x7c, 0x10, 0x4f, 0x68, 0x4e, 0x28, 0x4d, 0xb3, 0xbe, 0xc1, 0xde, 0x7c, 0xef, 0xb9, 0x3e, 0xf7,
	0xde, 0x73, 0x6c, 0xe8, 0x69, 0x3d, 0x1b, 0x95, 0x4a, 0x1a, 0x89, 0x3c, 0xc6, 0x0d, 0x11, 0x57,
	0x43, 0x5f, 0xcf, 0x88, 0xa2, 0x79, 0x95, 0x8d, 0x7e, 0x3a, 0xb0, 0x9f, 0x24, 0x67, 0x17, 0xa5,
	0x90, 0x24, 0x3f, 0x27, 0xd9, 0x47, 0x4c, 0x3f, 0x2d, 0xa8, 0x36, 0x68, 0x02, 0xa0, 0x68, 0x29,
	0x35, 0x37, 0x52, 0x5d, 0x05, 0x4e, 0xe8, 0xc4, 0xfd, 0x09, 0x1a, 0x55, 0x1c, 0x23, 0xbc, 0x42,
	0xf0, 0x5a, 0x15, 0xda, 0x87, 0xae, 0x36, 0x39, 0x2f, 0x82, 0x4e, 0xe8, 0xc4, 0x3e, 0xae, 0x02,
	0xf4, 0x14, 0x10, 0xe3, 0x26, 0xcd, 0x64, 0xf1, 0x81, 0xb3, 0x54, 0x96, 0x86, 0xcb, 0x42, 0x07,
	0x6e, 0xb8, 0x13, 0xf7, 0xf0, 0x80, 0x71, 0xf3, 0xc2, 0x02, 0x6f, 0xaa, 0x3c, 0x7a, 0x0c, 0xfe,
	0x75, 0xb5, 0x9d, 0x2e, 0x93, 0x22, 0xe8, 0x86, 0x4e, 0xdc, 0xc3, 0x7d, 0xc6, 0xcd, 0x79, 0x9d,
	0x7a, 0xed, 0xee, 0xee, 0x0c, 0x5c, 0x7c, 0xb0, 0x46, 0x5a, 0x12, 0x45, 0xe6, 0xd4, 0x50, 0xa5,
	0xa3, 0xcf, 0x70, 0xb0, 0xb1, 0x8f, 0x2e, 0x65, 0xa1, 0x29, 0x3a, 0x04, 0x4f, 0x9b, 0x5c, 0x2e,
	0x8c, 0x5d, 0xc6, 0xc7, 0x75, 0x54, 0xe7, 0xa9, 0x52, 0xf5, 0xd4, 0x75, 0x84, 0x4e, 0xa1, 0x4f,
	0x2f, 0xb9, 0x49, 0xb5, 0x21, 0x66, 0xa1, 0x83, 0x9d, 0xa6, 0x02, 0x2f, 0x2f, 0xb9, 0x49, 0x2c,
	0x82, 0x81, 0xae, 0xce, 0xd1, 0xb7, 0x8e, 0x6d, 0x8f, 0x69, 0x46, 0xf9, 0x92, 0xfe, 0x1f, 0x3d,
	0xef, 0x41, 0x97, 0x89, 0x94, 0xe7, 0x76, 0xa4, 0x1e, 0x76, 0x99, 0x78, 0x95, 0xa3, 0x27, 0xb0,
	0xc7, 0x44, 0xba, 0xd6, 0xc1, 0xb5, 0xa0, 0xcf, 0xc4, 0x5f, 0x6e, 0x74, 0x04, 0x7d, 0x26, 0xd2,
	0x85, 0xa6, 0xaa, 0x20, 0x73, 0x5a, 0x4b, 0x0b, 0x4c, 0x5c, 0xd4, 0x99, 0x96, 0xf8, 0x5e, 0x4b,
	0xfc, 0x2d, 0x6e, 0xde, 0xb9, 0xd9, 0xcd, 0xe8, 0x0b, 0x1c, 0x6e, 0xca, 0x71, 0x9b, 0x76, 0x64,
	0x70, 0x7f, 0xf5, 0x18, 0x9e, 0xa9, 0x6c, 0xc6, 0x97, 0xf4, 0x9f, 0xfb, 0x11, 0x7d, 0x85, 0xa0,
	0xdd, 0xe4, 0x16, 0xb7, 0x9c, 0xfc, 0xe8, 0x00, 0x24, 0xc9, 0x59, 0x42, 0xd5, 0x92, 0x67, 0x14,
	0xbd, 0x85, 0xbd, 0xc6, 0x0f, 0x40, 0x0f, 0xfe, 0xdc, 0xbf, 0xe9, 0xa3, 0x0f, 0x1f, 0x6e, 0x41,
	0xab, 0x0d, 0x22, 0xef, 0xd7, 0xf7, 0xb8, 0xb3, 0xdb, 0x89, 0x9d, 0x13, 0x07, 0xbd, 0x87, 0xbb,
	0x4d, 0x37, 0xd1, 0xfa, 0xe5, 0xf6, 0xa3, 0x1f, 0x3e, 0xda, 0x06, 0x37, 0xc8, 0x1d, 0x4b, 0x4e,
	0x60, 0xb0, 0x29, 0x23, 0x3a, 0x6a, 0xcd, 0xd6, 0x74, 0x71, 0x18, 0x6e, 0x2f, 0x68, 0xb7, 0x78,
	0x7e, 0xf2, 0xee, 0xba, 0x5c, 0x90, 0xe9, 0x28, 0x93, 0xf3, 0x71, 0x75, 0x3c, 0x96, 0x8a, 0x8d,
	0x2b, 0x92, 0x63, 0xfb, 0xee, 0xc7, 0x4c, 0xd6, 0x71, 0x39, 0x9d, 0x7a, 0x36, 0x75, 0xfa, 0x3b,
	0x00, 0x00, 0xff, 0xff, 0x39, 0xfb, 0x3b, 0x88, 0x48, 0x05, 0x00, 0x00,
}
