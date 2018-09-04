// Code generated by protoc-gen-go. DO NOT EDIT.
// source: smarthttp.proto

package gitaly

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

type InfoRefsRequest struct {
	Repository *Repository `protobuf:"bytes,1,opt,name=repository" json:"repository,omitempty"`
	// Parameters to use with git -c (key=value pairs)
	GitConfigOptions []string `protobuf:"bytes,2,rep,name=git_config_options,json=gitConfigOptions" json:"git_config_options,omitempty"`
	// Git protocol version
	GitProtocol string `protobuf:"bytes,3,opt,name=git_protocol,json=gitProtocol" json:"git_protocol,omitempty"`
}

func (m *InfoRefsRequest) Reset()                    { *m = InfoRefsRequest{} }
func (m *InfoRefsRequest) String() string            { return proto.CompactTextString(m) }
func (*InfoRefsRequest) ProtoMessage()               {}
func (*InfoRefsRequest) Descriptor() ([]byte, []int) { return fileDescriptor12, []int{0} }

func (m *InfoRefsRequest) GetRepository() *Repository {
	if m != nil {
		return m.Repository
	}
	return nil
}

func (m *InfoRefsRequest) GetGitConfigOptions() []string {
	if m != nil {
		return m.GitConfigOptions
	}
	return nil
}

func (m *InfoRefsRequest) GetGitProtocol() string {
	if m != nil {
		return m.GitProtocol
	}
	return ""
}

type InfoRefsResponse struct {
	Data []byte `protobuf:"bytes,1,opt,name=data,proto3" json:"data,omitempty"`
}

func (m *InfoRefsResponse) Reset()                    { *m = InfoRefsResponse{} }
func (m *InfoRefsResponse) String() string            { return proto.CompactTextString(m) }
func (*InfoRefsResponse) ProtoMessage()               {}
func (*InfoRefsResponse) Descriptor() ([]byte, []int) { return fileDescriptor12, []int{1} }

func (m *InfoRefsResponse) GetData() []byte {
	if m != nil {
		return m.Data
	}
	return nil
}

type PostUploadPackRequest struct {
	// repository should only be present in the first message of the stream
	Repository *Repository `protobuf:"bytes,1,opt,name=repository" json:"repository,omitempty"`
	// Raw data to be copied to stdin of 'git upload-pack'
	Data []byte `protobuf:"bytes,2,opt,name=data,proto3" json:"data,omitempty"`
	// Parameters to use with git -c (key=value pairs)
	GitConfigOptions []string `protobuf:"bytes,3,rep,name=git_config_options,json=gitConfigOptions" json:"git_config_options,omitempty"`
	// Git protocol version
	GitProtocol string `protobuf:"bytes,4,opt,name=git_protocol,json=gitProtocol" json:"git_protocol,omitempty"`
}

func (m *PostUploadPackRequest) Reset()                    { *m = PostUploadPackRequest{} }
func (m *PostUploadPackRequest) String() string            { return proto.CompactTextString(m) }
func (*PostUploadPackRequest) ProtoMessage()               {}
func (*PostUploadPackRequest) Descriptor() ([]byte, []int) { return fileDescriptor12, []int{2} }

func (m *PostUploadPackRequest) GetRepository() *Repository {
	if m != nil {
		return m.Repository
	}
	return nil
}

func (m *PostUploadPackRequest) GetData() []byte {
	if m != nil {
		return m.Data
	}
	return nil
}

func (m *PostUploadPackRequest) GetGitConfigOptions() []string {
	if m != nil {
		return m.GitConfigOptions
	}
	return nil
}

func (m *PostUploadPackRequest) GetGitProtocol() string {
	if m != nil {
		return m.GitProtocol
	}
	return ""
}

type PostUploadPackResponse struct {
	// Raw data from stdout of 'git upload-pack'
	Data []byte `protobuf:"bytes,1,opt,name=data,proto3" json:"data,omitempty"`
}

func (m *PostUploadPackResponse) Reset()                    { *m = PostUploadPackResponse{} }
func (m *PostUploadPackResponse) String() string            { return proto.CompactTextString(m) }
func (*PostUploadPackResponse) ProtoMessage()               {}
func (*PostUploadPackResponse) Descriptor() ([]byte, []int) { return fileDescriptor12, []int{3} }

func (m *PostUploadPackResponse) GetData() []byte {
	if m != nil {
		return m.Data
	}
	return nil
}

type PostReceivePackRequest struct {
	// repository should only be present in the first message of the stream
	Repository *Repository `protobuf:"bytes,1,opt,name=repository" json:"repository,omitempty"`
	// Raw data to be copied to stdin of 'git receive-pack'
	Data []byte `protobuf:"bytes,2,opt,name=data,proto3" json:"data,omitempty"`
	// gl_id, gl_repository, and gl_username become env variables, used by the Git {pre,post}-receive
	// hooks. They should only be present in the first message of the stream.
	GlId         string `protobuf:"bytes,3,opt,name=gl_id,json=glId" json:"gl_id,omitempty"`
	GlRepository string `protobuf:"bytes,4,opt,name=gl_repository,json=glRepository" json:"gl_repository,omitempty"`
	GlUsername   string `protobuf:"bytes,5,opt,name=gl_username,json=glUsername" json:"gl_username,omitempty"`
	// Git protocol version
	GitProtocol string `protobuf:"bytes,6,opt,name=git_protocol,json=gitProtocol" json:"git_protocol,omitempty"`
	// Parameters to use with git -c (key=value pairs)
	GitConfigOptions []string `protobuf:"bytes,7,rep,name=git_config_options,json=gitConfigOptions" json:"git_config_options,omitempty"`
}

func (m *PostReceivePackRequest) Reset()                    { *m = PostReceivePackRequest{} }
func (m *PostReceivePackRequest) String() string            { return proto.CompactTextString(m) }
func (*PostReceivePackRequest) ProtoMessage()               {}
func (*PostReceivePackRequest) Descriptor() ([]byte, []int) { return fileDescriptor12, []int{4} }

func (m *PostReceivePackRequest) GetRepository() *Repository {
	if m != nil {
		return m.Repository
	}
	return nil
}

func (m *PostReceivePackRequest) GetData() []byte {
	if m != nil {
		return m.Data
	}
	return nil
}

func (m *PostReceivePackRequest) GetGlId() string {
	if m != nil {
		return m.GlId
	}
	return ""
}

func (m *PostReceivePackRequest) GetGlRepository() string {
	if m != nil {
		return m.GlRepository
	}
	return ""
}

func (m *PostReceivePackRequest) GetGlUsername() string {
	if m != nil {
		return m.GlUsername
	}
	return ""
}

func (m *PostReceivePackRequest) GetGitProtocol() string {
	if m != nil {
		return m.GitProtocol
	}
	return ""
}

func (m *PostReceivePackRequest) GetGitConfigOptions() []string {
	if m != nil {
		return m.GitConfigOptions
	}
	return nil
}

type PostReceivePackResponse struct {
	// Raw data from stdout of 'git receive-pack'
	Data []byte `protobuf:"bytes,1,opt,name=data,proto3" json:"data,omitempty"`
}

func (m *PostReceivePackResponse) Reset()                    { *m = PostReceivePackResponse{} }
func (m *PostReceivePackResponse) String() string            { return proto.CompactTextString(m) }
func (*PostReceivePackResponse) ProtoMessage()               {}
func (*PostReceivePackResponse) Descriptor() ([]byte, []int) { return fileDescriptor12, []int{5} }

func (m *PostReceivePackResponse) GetData() []byte {
	if m != nil {
		return m.Data
	}
	return nil
}

func init() {
	proto.RegisterType((*InfoRefsRequest)(nil), "gitaly.InfoRefsRequest")
	proto.RegisterType((*InfoRefsResponse)(nil), "gitaly.InfoRefsResponse")
	proto.RegisterType((*PostUploadPackRequest)(nil), "gitaly.PostUploadPackRequest")
	proto.RegisterType((*PostUploadPackResponse)(nil), "gitaly.PostUploadPackResponse")
	proto.RegisterType((*PostReceivePackRequest)(nil), "gitaly.PostReceivePackRequest")
	proto.RegisterType((*PostReceivePackResponse)(nil), "gitaly.PostReceivePackResponse")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for SmartHTTPService service

type SmartHTTPServiceClient interface {
	// The response body for GET /info/refs?service=git-upload-pack
	InfoRefsUploadPack(ctx context.Context, in *InfoRefsRequest, opts ...grpc.CallOption) (SmartHTTPService_InfoRefsUploadPackClient, error)
	// The response body for GET /info/refs?service=git-receive-pack
	InfoRefsReceivePack(ctx context.Context, in *InfoRefsRequest, opts ...grpc.CallOption) (SmartHTTPService_InfoRefsReceivePackClient, error)
	// Request and response body for POST /upload-pack
	PostUploadPack(ctx context.Context, opts ...grpc.CallOption) (SmartHTTPService_PostUploadPackClient, error)
	// Request and response body for POST /receive-pack
	PostReceivePack(ctx context.Context, opts ...grpc.CallOption) (SmartHTTPService_PostReceivePackClient, error)
}

type smartHTTPServiceClient struct {
	cc *grpc.ClientConn
}

func NewSmartHTTPServiceClient(cc *grpc.ClientConn) SmartHTTPServiceClient {
	return &smartHTTPServiceClient{cc}
}

func (c *smartHTTPServiceClient) InfoRefsUploadPack(ctx context.Context, in *InfoRefsRequest, opts ...grpc.CallOption) (SmartHTTPService_InfoRefsUploadPackClient, error) {
	stream, err := grpc.NewClientStream(ctx, &_SmartHTTPService_serviceDesc.Streams[0], c.cc, "/gitaly.SmartHTTPService/InfoRefsUploadPack", opts...)
	if err != nil {
		return nil, err
	}
	x := &smartHTTPServiceInfoRefsUploadPackClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type SmartHTTPService_InfoRefsUploadPackClient interface {
	Recv() (*InfoRefsResponse, error)
	grpc.ClientStream
}

type smartHTTPServiceInfoRefsUploadPackClient struct {
	grpc.ClientStream
}

func (x *smartHTTPServiceInfoRefsUploadPackClient) Recv() (*InfoRefsResponse, error) {
	m := new(InfoRefsResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *smartHTTPServiceClient) InfoRefsReceivePack(ctx context.Context, in *InfoRefsRequest, opts ...grpc.CallOption) (SmartHTTPService_InfoRefsReceivePackClient, error) {
	stream, err := grpc.NewClientStream(ctx, &_SmartHTTPService_serviceDesc.Streams[1], c.cc, "/gitaly.SmartHTTPService/InfoRefsReceivePack", opts...)
	if err != nil {
		return nil, err
	}
	x := &smartHTTPServiceInfoRefsReceivePackClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type SmartHTTPService_InfoRefsReceivePackClient interface {
	Recv() (*InfoRefsResponse, error)
	grpc.ClientStream
}

type smartHTTPServiceInfoRefsReceivePackClient struct {
	grpc.ClientStream
}

func (x *smartHTTPServiceInfoRefsReceivePackClient) Recv() (*InfoRefsResponse, error) {
	m := new(InfoRefsResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *smartHTTPServiceClient) PostUploadPack(ctx context.Context, opts ...grpc.CallOption) (SmartHTTPService_PostUploadPackClient, error) {
	stream, err := grpc.NewClientStream(ctx, &_SmartHTTPService_serviceDesc.Streams[2], c.cc, "/gitaly.SmartHTTPService/PostUploadPack", opts...)
	if err != nil {
		return nil, err
	}
	x := &smartHTTPServicePostUploadPackClient{stream}
	return x, nil
}

type SmartHTTPService_PostUploadPackClient interface {
	Send(*PostUploadPackRequest) error
	Recv() (*PostUploadPackResponse, error)
	grpc.ClientStream
}

type smartHTTPServicePostUploadPackClient struct {
	grpc.ClientStream
}

func (x *smartHTTPServicePostUploadPackClient) Send(m *PostUploadPackRequest) error {
	return x.ClientStream.SendMsg(m)
}

func (x *smartHTTPServicePostUploadPackClient) Recv() (*PostUploadPackResponse, error) {
	m := new(PostUploadPackResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *smartHTTPServiceClient) PostReceivePack(ctx context.Context, opts ...grpc.CallOption) (SmartHTTPService_PostReceivePackClient, error) {
	stream, err := grpc.NewClientStream(ctx, &_SmartHTTPService_serviceDesc.Streams[3], c.cc, "/gitaly.SmartHTTPService/PostReceivePack", opts...)
	if err != nil {
		return nil, err
	}
	x := &smartHTTPServicePostReceivePackClient{stream}
	return x, nil
}

type SmartHTTPService_PostReceivePackClient interface {
	Send(*PostReceivePackRequest) error
	Recv() (*PostReceivePackResponse, error)
	grpc.ClientStream
}

type smartHTTPServicePostReceivePackClient struct {
	grpc.ClientStream
}

func (x *smartHTTPServicePostReceivePackClient) Send(m *PostReceivePackRequest) error {
	return x.ClientStream.SendMsg(m)
}

func (x *smartHTTPServicePostReceivePackClient) Recv() (*PostReceivePackResponse, error) {
	m := new(PostReceivePackResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// Server API for SmartHTTPService service

type SmartHTTPServiceServer interface {
	// The response body for GET /info/refs?service=git-upload-pack
	InfoRefsUploadPack(*InfoRefsRequest, SmartHTTPService_InfoRefsUploadPackServer) error
	// The response body for GET /info/refs?service=git-receive-pack
	InfoRefsReceivePack(*InfoRefsRequest, SmartHTTPService_InfoRefsReceivePackServer) error
	// Request and response body for POST /upload-pack
	PostUploadPack(SmartHTTPService_PostUploadPackServer) error
	// Request and response body for POST /receive-pack
	PostReceivePack(SmartHTTPService_PostReceivePackServer) error
}

func RegisterSmartHTTPServiceServer(s *grpc.Server, srv SmartHTTPServiceServer) {
	s.RegisterService(&_SmartHTTPService_serviceDesc, srv)
}

func _SmartHTTPService_InfoRefsUploadPack_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(InfoRefsRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(SmartHTTPServiceServer).InfoRefsUploadPack(m, &smartHTTPServiceInfoRefsUploadPackServer{stream})
}

type SmartHTTPService_InfoRefsUploadPackServer interface {
	Send(*InfoRefsResponse) error
	grpc.ServerStream
}

type smartHTTPServiceInfoRefsUploadPackServer struct {
	grpc.ServerStream
}

func (x *smartHTTPServiceInfoRefsUploadPackServer) Send(m *InfoRefsResponse) error {
	return x.ServerStream.SendMsg(m)
}

func _SmartHTTPService_InfoRefsReceivePack_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(InfoRefsRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(SmartHTTPServiceServer).InfoRefsReceivePack(m, &smartHTTPServiceInfoRefsReceivePackServer{stream})
}

type SmartHTTPService_InfoRefsReceivePackServer interface {
	Send(*InfoRefsResponse) error
	grpc.ServerStream
}

type smartHTTPServiceInfoRefsReceivePackServer struct {
	grpc.ServerStream
}

func (x *smartHTTPServiceInfoRefsReceivePackServer) Send(m *InfoRefsResponse) error {
	return x.ServerStream.SendMsg(m)
}

func _SmartHTTPService_PostUploadPack_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(SmartHTTPServiceServer).PostUploadPack(&smartHTTPServicePostUploadPackServer{stream})
}

type SmartHTTPService_PostUploadPackServer interface {
	Send(*PostUploadPackResponse) error
	Recv() (*PostUploadPackRequest, error)
	grpc.ServerStream
}

type smartHTTPServicePostUploadPackServer struct {
	grpc.ServerStream
}

func (x *smartHTTPServicePostUploadPackServer) Send(m *PostUploadPackResponse) error {
	return x.ServerStream.SendMsg(m)
}

func (x *smartHTTPServicePostUploadPackServer) Recv() (*PostUploadPackRequest, error) {
	m := new(PostUploadPackRequest)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _SmartHTTPService_PostReceivePack_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(SmartHTTPServiceServer).PostReceivePack(&smartHTTPServicePostReceivePackServer{stream})
}

type SmartHTTPService_PostReceivePackServer interface {
	Send(*PostReceivePackResponse) error
	Recv() (*PostReceivePackRequest, error)
	grpc.ServerStream
}

type smartHTTPServicePostReceivePackServer struct {
	grpc.ServerStream
}

func (x *smartHTTPServicePostReceivePackServer) Send(m *PostReceivePackResponse) error {
	return x.ServerStream.SendMsg(m)
}

func (x *smartHTTPServicePostReceivePackServer) Recv() (*PostReceivePackRequest, error) {
	m := new(PostReceivePackRequest)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

var _SmartHTTPService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "gitaly.SmartHTTPService",
	HandlerType: (*SmartHTTPServiceServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "InfoRefsUploadPack",
			Handler:       _SmartHTTPService_InfoRefsUploadPack_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "InfoRefsReceivePack",
			Handler:       _SmartHTTPService_InfoRefsReceivePack_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "PostUploadPack",
			Handler:       _SmartHTTPService_PostUploadPack_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
		{
			StreamName:    "PostReceivePack",
			Handler:       _SmartHTTPService_PostReceivePack_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "smarthttp.proto",
}

func init() { proto.RegisterFile("smarthttp.proto", fileDescriptor12) }

var fileDescriptor12 = []byte{
	// 423 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xb4, 0x53, 0xd1, 0x8a, 0xd3, 0x40,
	0x14, 0x75, 0xd2, 0x6e, 0x65, 0x6f, 0xa3, 0x2d, 0x77, 0xd1, 0x0d, 0x01, 0xdd, 0x1a, 0x41, 0xf2,
	0xb0, 0x96, 0xa5, 0x7e, 0x82, 0x2f, 0x2e, 0x0a, 0x86, 0xd9, 0x2d, 0xf8, 0x16, 0xc6, 0x64, 0x3a,
	0x3b, 0x38, 0x9b, 0x89, 0x99, 0xd9, 0x85, 0xfe, 0x83, 0xcf, 0x7e, 0x87, 0x5f, 0xe4, 0xb7, 0x48,
	0x93, 0xa6, 0x69, 0x9b, 0x46, 0x44, 0xf1, 0x2d, 0xdc, 0x7b, 0x72, 0xee, 0x39, 0xe7, 0xde, 0x81,
	0x91, 0xb9, 0x65, 0x85, 0xbd, 0xb1, 0x36, 0x9f, 0xe6, 0x85, 0xb6, 0x1a, 0x07, 0x42, 0x5a, 0xa6,
	0x96, 0xbe, 0x6b, 0x6e, 0x58, 0xc1, 0xd3, 0xaa, 0x1a, 0x7c, 0x27, 0x30, 0xba, 0xcc, 0x16, 0x9a,
	0xf2, 0x85, 0xa1, 0xfc, 0xeb, 0x1d, 0x37, 0x16, 0x67, 0x00, 0x05, 0xcf, 0xb5, 0x91, 0x56, 0x17,
	0x4b, 0x8f, 0x4c, 0x48, 0x38, 0x9c, 0xe1, 0xb4, 0xfa, 0x7d, 0x4a, 0x37, 0x1d, 0xba, 0x85, 0xc2,
	0x73, 0x40, 0x21, 0x6d, 0x9c, 0xe8, 0x6c, 0x21, 0x45, 0xac, 0x73, 0x2b, 0x75, 0x66, 0x3c, 0x67,
	0xd2, 0x0b, 0x8f, 0xe9, 0x58, 0x48, 0xfb, 0xb6, 0x6c, 0x7c, 0xac, 0xea, 0xf8, 0x02, 0xdc, 0x15,
	0xba, 0x94, 0x90, 0x68, 0xe5, 0xf5, 0x26, 0x24, 0x3c, 0xa6, 0x43, 0x21, 0x6d, 0xb4, 0x2e, 0x05,
	0xaf, 0x60, 0xdc, 0xe8, 0x32, 0xb9, 0xce, 0x0c, 0x47, 0x84, 0x7e, 0xca, 0x2c, 0x2b, 0x25, 0xb9,
	0xb4, 0xfc, 0x0e, 0x7e, 0x10, 0x78, 0x12, 0x69, 0x63, 0xe7, 0xb9, 0xd2, 0x2c, 0x8d, 0x58, 0xf2,
	0xe5, 0x5f, 0x6c, 0xd4, 0x13, 0x9c, 0x66, 0x42, 0x87, 0xb5, 0xde, 0x1f, 0x5a, 0xeb, 0xb7, 0xad,
	0x9d, 0xc3, 0xd3, 0x7d, 0xc5, 0xbf, 0x31, 0xf8, 0xcd, 0xa9, 0xe0, 0x94, 0x27, 0x5c, 0xde, 0xf3,
	0xff, 0xe1, 0xf0, 0x04, 0x8e, 0x84, 0x8a, 0x65, 0xba, 0xde, 0x43, 0x5f, 0xa8, 0xcb, 0x14, 0x5f,
	0xc2, 0x23, 0xa1, 0xe2, 0x2d, 0xfe, 0xca, 0x89, 0x2b, 0x54, 0xc3, 0x8c, 0x67, 0x30, 0x14, 0x2a,
	0xbe, 0x33, 0xbc, 0xc8, 0xd8, 0x2d, 0xf7, 0x8e, 0x4a, 0x08, 0x08, 0x35, 0x5f, 0x57, 0x5a, 0x71,
	0x0c, 0x5a, 0x71, 0x74, 0xe4, 0xfb, 0xf0, 0x70, 0xbe, 0xc1, 0x6b, 0x38, 0x6d, 0xa5, 0xd1, 0x9d,
	0xde, 0xec, 0xa7, 0x03, 0xe3, 0xab, 0xd5, 0x4b, 0x78, 0x77, 0x7d, 0x1d, 0x5d, 0xf1, 0xe2, 0x5e,
	0x26, 0x1c, 0xdf, 0x03, 0xd6, 0xb7, 0xd5, 0x2c, 0x01, 0x4f, 0xeb, 0xe4, 0xf6, 0xde, 0x83, 0xef,
	0xb5, 0x1b, 0xd5, 0xc4, 0xe0, 0xc1, 0x05, 0xc1, 0x0f, 0x70, 0xd2, 0xd4, 0x37, 0xa2, 0xfe, 0x96,
	0x6d, 0x0e, 0x8f, 0x77, 0x6f, 0x03, 0x9f, 0xd5, 0xf8, 0x83, 0x57, 0xee, 0x3f, 0xef, 0x6a, 0xd7,
	0xa4, 0x21, 0xb9, 0x20, 0xf8, 0x09, 0x46, 0x7b, 0xa9, 0xe1, 0xce, 0x8f, 0xed, 0xe3, 0xf2, 0xcf,
	0x3a, 0xfb, 0xdb, 0xcc, 0x9f, 0x07, 0xe5, 0x6a, 0xdf, 0xfc, 0x0a, 0x00, 0x00, 0xff, 0xff, 0xf1,
	0x46, 0x34, 0x66, 0x70, 0x04, 0x00, 0x00,
}
