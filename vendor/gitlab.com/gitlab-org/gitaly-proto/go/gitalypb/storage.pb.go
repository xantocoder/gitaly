// Code generated by protoc-gen-go. DO NOT EDIT.
// source: storage.proto

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

type ListDirectoriesRequest struct {
	StorageName          string   `protobuf:"bytes,1,opt,name=storage_name,json=storageName,proto3" json:"storage_name,omitempty"`
	Depth                uint32   `protobuf:"varint,2,opt,name=depth,proto3" json:"depth,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ListDirectoriesRequest) Reset()         { *m = ListDirectoriesRequest{} }
func (m *ListDirectoriesRequest) String() string { return proto.CompactTextString(m) }
func (*ListDirectoriesRequest) ProtoMessage()    {}
func (*ListDirectoriesRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_storage_71656ad921f7d066, []int{0}
}
func (m *ListDirectoriesRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ListDirectoriesRequest.Unmarshal(m, b)
}
func (m *ListDirectoriesRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ListDirectoriesRequest.Marshal(b, m, deterministic)
}
func (dst *ListDirectoriesRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ListDirectoriesRequest.Merge(dst, src)
}
func (m *ListDirectoriesRequest) XXX_Size() int {
	return xxx_messageInfo_ListDirectoriesRequest.Size(m)
}
func (m *ListDirectoriesRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_ListDirectoriesRequest.DiscardUnknown(m)
}

var xxx_messageInfo_ListDirectoriesRequest proto.InternalMessageInfo

func (m *ListDirectoriesRequest) GetStorageName() string {
	if m != nil {
		return m.StorageName
	}
	return ""
}

func (m *ListDirectoriesRequest) GetDepth() uint32 {
	if m != nil {
		return m.Depth
	}
	return 0
}

type ListDirectoriesResponse struct {
	Paths                []string `protobuf:"bytes,1,rep,name=paths,proto3" json:"paths,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ListDirectoriesResponse) Reset()         { *m = ListDirectoriesResponse{} }
func (m *ListDirectoriesResponse) String() string { return proto.CompactTextString(m) }
func (*ListDirectoriesResponse) ProtoMessage()    {}
func (*ListDirectoriesResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_storage_71656ad921f7d066, []int{1}
}
func (m *ListDirectoriesResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ListDirectoriesResponse.Unmarshal(m, b)
}
func (m *ListDirectoriesResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ListDirectoriesResponse.Marshal(b, m, deterministic)
}
func (dst *ListDirectoriesResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ListDirectoriesResponse.Merge(dst, src)
}
func (m *ListDirectoriesResponse) XXX_Size() int {
	return xxx_messageInfo_ListDirectoriesResponse.Size(m)
}
func (m *ListDirectoriesResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_ListDirectoriesResponse.DiscardUnknown(m)
}

var xxx_messageInfo_ListDirectoriesResponse proto.InternalMessageInfo

func (m *ListDirectoriesResponse) GetPaths() []string {
	if m != nil {
		return m.Paths
	}
	return nil
}

type DeleteAllRepositoriesRequest struct {
	StorageName          string   `protobuf:"bytes,1,opt,name=storage_name,json=storageName,proto3" json:"storage_name,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *DeleteAllRepositoriesRequest) Reset()         { *m = DeleteAllRepositoriesRequest{} }
func (m *DeleteAllRepositoriesRequest) String() string { return proto.CompactTextString(m) }
func (*DeleteAllRepositoriesRequest) ProtoMessage()    {}
func (*DeleteAllRepositoriesRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_storage_71656ad921f7d066, []int{2}
}
func (m *DeleteAllRepositoriesRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_DeleteAllRepositoriesRequest.Unmarshal(m, b)
}
func (m *DeleteAllRepositoriesRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_DeleteAllRepositoriesRequest.Marshal(b, m, deterministic)
}
func (dst *DeleteAllRepositoriesRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_DeleteAllRepositoriesRequest.Merge(dst, src)
}
func (m *DeleteAllRepositoriesRequest) XXX_Size() int {
	return xxx_messageInfo_DeleteAllRepositoriesRequest.Size(m)
}
func (m *DeleteAllRepositoriesRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_DeleteAllRepositoriesRequest.DiscardUnknown(m)
}

var xxx_messageInfo_DeleteAllRepositoriesRequest proto.InternalMessageInfo

func (m *DeleteAllRepositoriesRequest) GetStorageName() string {
	if m != nil {
		return m.StorageName
	}
	return ""
}

type DeleteAllRepositoriesResponse struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *DeleteAllRepositoriesResponse) Reset()         { *m = DeleteAllRepositoriesResponse{} }
func (m *DeleteAllRepositoriesResponse) String() string { return proto.CompactTextString(m) }
func (*DeleteAllRepositoriesResponse) ProtoMessage()    {}
func (*DeleteAllRepositoriesResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_storage_71656ad921f7d066, []int{3}
}
func (m *DeleteAllRepositoriesResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_DeleteAllRepositoriesResponse.Unmarshal(m, b)
}
func (m *DeleteAllRepositoriesResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_DeleteAllRepositoriesResponse.Marshal(b, m, deterministic)
}
func (dst *DeleteAllRepositoriesResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_DeleteAllRepositoriesResponse.Merge(dst, src)
}
func (m *DeleteAllRepositoriesResponse) XXX_Size() int {
	return xxx_messageInfo_DeleteAllRepositoriesResponse.Size(m)
}
func (m *DeleteAllRepositoriesResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_DeleteAllRepositoriesResponse.DiscardUnknown(m)
}

var xxx_messageInfo_DeleteAllRepositoriesResponse proto.InternalMessageInfo

func init() {
	proto.RegisterType((*ListDirectoriesRequest)(nil), "gitaly.ListDirectoriesRequest")
	proto.RegisterType((*ListDirectoriesResponse)(nil), "gitaly.ListDirectoriesResponse")
	proto.RegisterType((*DeleteAllRepositoriesRequest)(nil), "gitaly.DeleteAllRepositoriesRequest")
	proto.RegisterType((*DeleteAllRepositoriesResponse)(nil), "gitaly.DeleteAllRepositoriesResponse")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// StorageServiceClient is the client API for StorageService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type StorageServiceClient interface {
	ListDirectories(ctx context.Context, in *ListDirectoriesRequest, opts ...grpc.CallOption) (StorageService_ListDirectoriesClient, error)
	DeleteAllRepositories(ctx context.Context, in *DeleteAllRepositoriesRequest, opts ...grpc.CallOption) (*DeleteAllRepositoriesResponse, error)
}

type storageServiceClient struct {
	cc *grpc.ClientConn
}

func NewStorageServiceClient(cc *grpc.ClientConn) StorageServiceClient {
	return &storageServiceClient{cc}
}

func (c *storageServiceClient) ListDirectories(ctx context.Context, in *ListDirectoriesRequest, opts ...grpc.CallOption) (StorageService_ListDirectoriesClient, error) {
	stream, err := c.cc.NewStream(ctx, &_StorageService_serviceDesc.Streams[0], "/gitaly.StorageService/ListDirectories", opts...)
	if err != nil {
		return nil, err
	}
	x := &storageServiceListDirectoriesClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type StorageService_ListDirectoriesClient interface {
	Recv() (*ListDirectoriesResponse, error)
	grpc.ClientStream
}

type storageServiceListDirectoriesClient struct {
	grpc.ClientStream
}

func (x *storageServiceListDirectoriesClient) Recv() (*ListDirectoriesResponse, error) {
	m := new(ListDirectoriesResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *storageServiceClient) DeleteAllRepositories(ctx context.Context, in *DeleteAllRepositoriesRequest, opts ...grpc.CallOption) (*DeleteAllRepositoriesResponse, error) {
	out := new(DeleteAllRepositoriesResponse)
	err := c.cc.Invoke(ctx, "/gitaly.StorageService/DeleteAllRepositories", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// StorageServiceServer is the server API for StorageService service.
type StorageServiceServer interface {
	ListDirectories(*ListDirectoriesRequest, StorageService_ListDirectoriesServer) error
	DeleteAllRepositories(context.Context, *DeleteAllRepositoriesRequest) (*DeleteAllRepositoriesResponse, error)
}

func RegisterStorageServiceServer(s *grpc.Server, srv StorageServiceServer) {
	s.RegisterService(&_StorageService_serviceDesc, srv)
}

func _StorageService_ListDirectories_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(ListDirectoriesRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(StorageServiceServer).ListDirectories(m, &storageServiceListDirectoriesServer{stream})
}

type StorageService_ListDirectoriesServer interface {
	Send(*ListDirectoriesResponse) error
	grpc.ServerStream
}

type storageServiceListDirectoriesServer struct {
	grpc.ServerStream
}

func (x *storageServiceListDirectoriesServer) Send(m *ListDirectoriesResponse) error {
	return x.ServerStream.SendMsg(m)
}

func _StorageService_DeleteAllRepositories_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteAllRepositoriesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StorageServiceServer).DeleteAllRepositories(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/gitaly.StorageService/DeleteAllRepositories",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StorageServiceServer).DeleteAllRepositories(ctx, req.(*DeleteAllRepositoriesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _StorageService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "gitaly.StorageService",
	HandlerType: (*StorageServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "DeleteAllRepositories",
			Handler:    _StorageService_DeleteAllRepositories_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "ListDirectories",
			Handler:       _StorageService_ListDirectories_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "storage.proto",
}

func init() { proto.RegisterFile("storage.proto", fileDescriptor_storage_71656ad921f7d066) }

var fileDescriptor_storage_71656ad921f7d066 = []byte{
	// 285 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x94, 0x91, 0x5f, 0x4b, 0xf3, 0x30,
	0x14, 0xc6, 0xc9, 0x5e, 0xde, 0xe2, 0x8e, 0x9b, 0x42, 0xf0, 0x4f, 0x29, 0xea, 0x6a, 0x51, 0xe8,
	0xcd, 0xda, 0xa1, 0x9f, 0x60, 0xb2, 0x4b, 0x11, 0xcc, 0xee, 0x44, 0x90, 0xb4, 0x3b, 0xb4, 0x81,
	0x74, 0x89, 0x49, 0x14, 0xfc, 0x24, 0x7e, 0x39, 0x3f, 0x89, 0x57, 0x62, 0x53, 0x2f, 0xd4, 0x4d,
	0xf1, 0x2e, 0xe7, 0xc9, 0x39, 0xbf, 0xe7, 0xc9, 0x09, 0x0c, 0xad, 0x53, 0x86, 0x57, 0x98, 0x69,
	0xa3, 0x9c, 0xa2, 0x41, 0x25, 0x1c, 0x97, 0x4f, 0xd1, 0xc0, 0xd6, 0xdc, 0xe0, 0xc2, 0xab, 0xc9,
	0x35, 0xec, 0x5d, 0x0a, 0xeb, 0x66, 0xc2, 0x60, 0xe9, 0x94, 0x11, 0x68, 0x19, 0xde, 0x3f, 0xa0,
	0x75, 0xf4, 0x18, 0x06, 0x1d, 0xe0, 0x6e, 0xc9, 0x1b, 0x0c, 0x49, 0x4c, 0xd2, 0x3e, 0xdb, 0xec,
	0xb4, 0x2b, 0xde, 0x20, 0xdd, 0x81, 0xff, 0x0b, 0xd4, 0xae, 0x0e, 0x7b, 0x31, 0x49, 0x87, 0xcc,
	0x17, 0x49, 0x0e, 0xfb, 0xdf, 0x90, 0x56, 0xab, 0xa5, 0x6d, 0x07, 0x34, 0x77, 0xb5, 0x0d, 0x49,
	0xfc, 0x2f, 0xed, 0x33, 0x5f, 0x24, 0x53, 0x38, 0x98, 0xa1, 0x44, 0x87, 0x53, 0x29, 0x19, 0x6a,
	0x65, 0xc5, 0x5f, 0x93, 0x24, 0x23, 0x38, 0x5c, 0x83, 0xf0, 0xce, 0x67, 0x2f, 0x04, 0xb6, 0xe6,
	0x7e, 0x60, 0x8e, 0xe6, 0x51, 0x94, 0x48, 0x6f, 0x61, 0xfb, 0x4b, 0x4e, 0x7a, 0x94, 0xf9, 0x25,
	0x65, 0xab, 0x77, 0x12, 0x8d, 0xd6, 0xde, 0x7b, 0x9b, 0x24, 0x78, 0x7d, 0x4e, 0x7b, 0x1b, 0x64,
	0x42, 0xa8, 0x84, 0xdd, 0x95, 0x89, 0xe8, 0xc9, 0x07, 0xe3, 0xa7, 0x37, 0x47, 0xa7, 0xbf, 0x74,
	0x7d, 0xf6, 0xbb, 0x98, 0xdc, 0xbc, 0xf7, 0x4b, 0x5e, 0x64, 0xa5, 0x6a, 0x72, 0x7f, 0x1c, 0x2b,
	0x53, 0xe5, 0x9e, 0x32, 0x6e, 0x3f, 0x3b, 0xaf, 0x54, 0x57, 0xeb, 0xa2, 0x08, 0x5a, 0xe9, 0xfc,
	0x2d, 0x00, 0x00, 0xff, 0xff, 0xb7, 0xc3, 0x8d, 0x13, 0x26, 0x02, 0x00, 0x00,
}
