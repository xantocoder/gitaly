// Code generated by protoc-gen-go. DO NOT EDIT.
// source: lint.proto

package gitalypb

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	descriptor "github.com/golang/protobuf/protoc-gen-go/descriptor"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

type OperationMsg_Operation int32

const (
	OperationMsg_UNKNOWN  OperationMsg_Operation = 0
	OperationMsg_MUTATOR  OperationMsg_Operation = 1
	OperationMsg_ACCESSOR OperationMsg_Operation = 2
)

var OperationMsg_Operation_name = map[int32]string{
	0: "UNKNOWN",
	1: "MUTATOR",
	2: "ACCESSOR",
}

var OperationMsg_Operation_value = map[string]int32{
	"UNKNOWN":  0,
	"MUTATOR":  1,
	"ACCESSOR": 2,
}

func (x OperationMsg_Operation) String() string {
	return proto.EnumName(OperationMsg_Operation_name, int32(x))
}

func (OperationMsg_Operation) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_1612d42a10b555ca, []int{0, 0}
}

type OperationMsg_Scope int32

const (
	OperationMsg_REPOSITORY OperationMsg_Scope = 0
	OperationMsg_STORAGE    OperationMsg_Scope = 2
)

var OperationMsg_Scope_name = map[int32]string{
	0: "REPOSITORY",
	2: "STORAGE",
}

var OperationMsg_Scope_value = map[string]int32{
	"REPOSITORY": 0,
	"STORAGE":    2,
}

func (x OperationMsg_Scope) String() string {
	return proto.EnumName(OperationMsg_Scope_name, int32(x))
}

func (OperationMsg_Scope) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_1612d42a10b555ca, []int{0, 1}
}

type OperationMsg struct {
	Op OperationMsg_Operation `protobuf:"varint,1,opt,name=op,proto3,enum=gitaly.OperationMsg_Operation" json:"op,omitempty"`
	// Scope level indicates what level an RPC interacts with a server:
	//   - REPOSITORY: scoped to only a single repo
	//   - SERVER: affects the entire server and potentially all repos
	//   - STORAGE: scoped to a specific storage location and all repos within
	ScopeLevel           OperationMsg_Scope `protobuf:"varint,2,opt,name=scope_level,json=scopeLevel,proto3,enum=gitaly.OperationMsg_Scope" json:"scope_level,omitempty"`
	XXX_NoUnkeyedLiteral struct{}           `json:"-"`
	XXX_unrecognized     []byte             `json:"-"`
	XXX_sizecache        int32              `json:"-"`
}

func (m *OperationMsg) Reset()         { *m = OperationMsg{} }
func (m *OperationMsg) String() string { return proto.CompactTextString(m) }
func (*OperationMsg) ProtoMessage()    {}
func (*OperationMsg) Descriptor() ([]byte, []int) {
	return fileDescriptor_1612d42a10b555ca, []int{0}
}

func (m *OperationMsg) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_OperationMsg.Unmarshal(m, b)
}
func (m *OperationMsg) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_OperationMsg.Marshal(b, m, deterministic)
}
func (m *OperationMsg) XXX_Merge(src proto.Message) {
	xxx_messageInfo_OperationMsg.Merge(m, src)
}
func (m *OperationMsg) XXX_Size() int {
	return xxx_messageInfo_OperationMsg.Size(m)
}
func (m *OperationMsg) XXX_DiscardUnknown() {
	xxx_messageInfo_OperationMsg.DiscardUnknown(m)
}

var xxx_messageInfo_OperationMsg proto.InternalMessageInfo

func (m *OperationMsg) GetOp() OperationMsg_Operation {
	if m != nil {
		return m.Op
	}
	return OperationMsg_UNKNOWN
}

func (m *OperationMsg) GetScopeLevel() OperationMsg_Scope {
	if m != nil {
		return m.ScopeLevel
	}
	return OperationMsg_REPOSITORY
}

var E_Intercepted = &proto.ExtensionDesc{
	ExtendedType:  (*descriptor.ServiceOptions)(nil),
	ExtensionType: (*bool)(nil),
	Field:         82302,
	Name:          "gitaly.intercepted",
	Tag:           "varint,82302,opt,name=intercepted",
	Filename:      "lint.proto",
}

var E_OpType = &proto.ExtensionDesc{
	ExtendedType:  (*descriptor.MethodOptions)(nil),
	ExtensionType: (*OperationMsg)(nil),
	Field:         82303,
	Name:          "gitaly.op_type",
	Tag:           "bytes,82303,opt,name=op_type",
	Filename:      "lint.proto",
}

var E_Storage = &proto.ExtensionDesc{
	ExtendedType:  (*descriptor.FieldOptions)(nil),
	ExtensionType: (*bool)(nil),
	Field:         91233,
	Name:          "gitaly.storage",
	Tag:           "varint,91233,opt,name=storage",
	Filename:      "lint.proto",
}

var E_Repository = &proto.ExtensionDesc{
	ExtendedType:  (*descriptor.FieldOptions)(nil),
	ExtensionType: (*bool)(nil),
	Field:         91234,
	Name:          "gitaly.repository",
	Tag:           "varint,91234,opt,name=repository",
	Filename:      "lint.proto",
}

var E_TargetRepository = &proto.ExtensionDesc{
	ExtendedType:  (*descriptor.FieldOptions)(nil),
	ExtensionType: (*bool)(nil),
	Field:         91235,
	Name:          "gitaly.target_repository",
	Tag:           "varint,91235,opt,name=target_repository",
	Filename:      "lint.proto",
}

var E_AdditionalRepository = &proto.ExtensionDesc{
	ExtendedType:  (*descriptor.FieldOptions)(nil),
	ExtensionType: (*bool)(nil),
	Field:         91236,
	Name:          "gitaly.additional_repository",
	Tag:           "varint,91236,opt,name=additional_repository",
	Filename:      "lint.proto",
}

func init() {
	proto.RegisterEnum("gitaly.OperationMsg_Operation", OperationMsg_Operation_name, OperationMsg_Operation_value)
	proto.RegisterEnum("gitaly.OperationMsg_Scope", OperationMsg_Scope_name, OperationMsg_Scope_value)
	proto.RegisterType((*OperationMsg)(nil), "gitaly.OperationMsg")
	proto.RegisterExtension(E_Intercepted)
	proto.RegisterExtension(E_OpType)
	proto.RegisterExtension(E_Storage)
	proto.RegisterExtension(E_Repository)
	proto.RegisterExtension(E_TargetRepository)
	proto.RegisterExtension(E_AdditionalRepository)
}

func init() { proto.RegisterFile("lint.proto", fileDescriptor_1612d42a10b555ca) }

var fileDescriptor_1612d42a10b555ca = []byte{
	// 426 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x94, 0x92, 0xd1, 0x8a, 0xd3, 0x40,
	0x14, 0x86, 0x4d, 0x70, 0xdb, 0x7a, 0xba, 0x2c, 0x71, 0x58, 0xa1, 0x2c, 0xb8, 0x96, 0x5e, 0x2d,
	0x82, 0x89, 0x74, 0xaf, 0x8c, 0x17, 0x52, 0x4b, 0x14, 0x71, 0xdb, 0x91, 0x49, 0x56, 0xd1, 0x9b,
	0x92, 0x26, 0xc7, 0x38, 0x10, 0x7b, 0x86, 0xc9, 0xb8, 0xd0, 0x5b, 0x9f, 0xce, 0xd7, 0x50, 0xf7,
	0x39, 0x54, 0x92, 0x69, 0x37, 0x85, 0x5d, 0x50, 0xef, 0xe6, 0x1c, 0xfe, 0xef, 0xe3, 0xf0, 0x33,
	0x00, 0xa5, 0x5c, 0x19, 0x5f, 0x69, 0x32, 0xc4, 0x3a, 0x85, 0x34, 0x69, 0xb9, 0x3e, 0x1a, 0x16,
	0x44, 0x45, 0x89, 0x41, 0xb3, 0x5d, 0x7e, 0xf9, 0x18, 0xe4, 0x58, 0x65, 0x5a, 0x2a, 0x43, 0xda,
	0x26, 0x47, 0x97, 0x0e, 0xec, 0x73, 0x85, 0x3a, 0x35, 0x92, 0x56, 0xb3, 0xaa, 0x60, 0x3e, 0xb8,
	0xa4, 0x06, 0xce, 0xd0, 0x39, 0x39, 0x18, 0x1f, 0xfb, 0xd6, 0xe3, 0xef, 0x26, 0xda, 0x41, 0xb8,
	0xa4, 0xd8, 0x53, 0xe8, 0x57, 0x19, 0x29, 0x5c, 0x94, 0x78, 0x81, 0xe5, 0xc0, 0x6d, 0xc0, 0xa3,
	0x1b, 0xc1, 0xb8, 0xce, 0x09, 0x68, 0xe2, 0x67, 0x75, 0x7a, 0x74, 0x0a, 0x77, 0xae, 0x12, 0xac,
	0x0f, 0xdd, 0xf3, 0xf9, 0xeb, 0x39, 0x7f, 0x37, 0xf7, 0x6e, 0xd5, 0xc3, 0xec, 0x3c, 0x99, 0x24,
	0x5c, 0x78, 0x0e, 0xdb, 0x87, 0xde, 0x64, 0x3a, 0x8d, 0xe2, 0x98, 0x0b, 0xcf, 0x1d, 0x8d, 0x61,
	0xaf, 0x31, 0xb1, 0x03, 0x00, 0x11, 0xbd, 0xe1, 0xf1, 0xab, 0x84, 0x8b, 0xf7, 0x96, 0x89, 0x13,
	0x2e, 0x26, 0x2f, 0x23, 0xcf, 0x1d, 0xdd, 0xee, 0x39, 0x9e, 0xf3, 0xb0, 0x13, 0x47, 0xe2, 0x6d,
	0x24, 0xc2, 0x29, 0xf4, 0xe5, 0xca, 0xa0, 0xce, 0x50, 0x19, 0xcc, 0xd9, 0x03, 0xdf, 0x16, 0xe3,
	0x6f, 0x8b, 0xf1, 0x63, 0xd4, 0x17, 0x32, 0x43, 0xae, 0xea, 0x53, 0xaa, 0xc1, 0xaf, 0xaf, 0x7b,
	0x43, 0xe7, 0xa4, 0x27, 0x76, 0xa9, 0x90, 0x43, 0x97, 0xd4, 0xc2, 0xac, 0x15, 0xb2, 0xe3, 0x6b,
	0x82, 0x19, 0x9a, 0x4f, 0x94, 0x6f, 0xf9, 0xdf, 0x0d, 0xdf, 0x1f, 0x1f, 0xde, 0x54, 0x84, 0xe8,
	0x90, 0x4a, 0xd6, 0x0a, 0xc3, 0x27, 0xd0, 0xad, 0x0c, 0xe9, 0xb4, 0x40, 0x76, 0xff, 0x9a, 0xf0,
	0x85, 0xc4, 0xf2, 0xca, 0xf7, 0xfd, 0x9b, 0xbd, 0x67, 0x9b, 0x0f, 0x9f, 0x01, 0x68, 0x54, 0x54,
	0x49, 0x43, 0x7a, 0xfd, 0x37, 0xfa, 0xc7, 0x86, 0xde, 0x41, 0xc2, 0x33, 0xb8, 0x6b, 0x52, 0x5d,
	0xa0, 0x59, 0xfc, 0xbb, 0xe7, 0xe7, 0xc6, 0xe3, 0x59, 0x52, 0xb4, 0xb6, 0x04, 0xee, 0xa5, 0x79,
	0x2e, 0xeb, 0x58, 0x5a, 0xfe, 0x87, 0xf1, 0x72, 0x63, 0x3c, 0x6c, 0xe9, 0xd6, 0xfa, 0xfc, 0xf1,
	0x87, 0xba, 0xbe, 0x32, 0x5d, 0xfa, 0x19, 0x7d, 0x0e, 0xec, 0xf3, 0x11, 0xe9, 0x22, 0xb0, 0xa5,
	0xda, 0x6f, 0x1d, 0x14, 0xb4, 0x99, 0xd5, 0x72, 0xd9, 0x69, 0x56, 0xa7, 0x7f, 0x02, 0x00, 0x00,
	0xff, 0xff, 0x80, 0x8b, 0x42, 0x14, 0x0d, 0x03, 0x00, 0x00,
}
