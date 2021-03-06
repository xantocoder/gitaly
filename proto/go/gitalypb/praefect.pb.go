// Code generated by protoc-gen-go. DO NOT EDIT.
// source: praefect.proto

package gitalypb

import (
	context "context"
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
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

type SetAuthoritativeStorageRequest struct {
	VirtualStorage       string   `protobuf:"bytes,1,opt,name=virtual_storage,json=virtualStorage,proto3" json:"virtual_storage,omitempty"`
	RelativePath         string   `protobuf:"bytes,2,opt,name=relative_path,json=relativePath,proto3" json:"relative_path,omitempty"`
	AuthoritativeStorage string   `protobuf:"bytes,3,opt,name=authoritative_storage,json=authoritativeStorage,proto3" json:"authoritative_storage,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *SetAuthoritativeStorageRequest) Reset()         { *m = SetAuthoritativeStorageRequest{} }
func (m *SetAuthoritativeStorageRequest) String() string { return proto.CompactTextString(m) }
func (*SetAuthoritativeStorageRequest) ProtoMessage()    {}
func (*SetAuthoritativeStorageRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_d32bf44842ead735, []int{0}
}

func (m *SetAuthoritativeStorageRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SetAuthoritativeStorageRequest.Unmarshal(m, b)
}
func (m *SetAuthoritativeStorageRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SetAuthoritativeStorageRequest.Marshal(b, m, deterministic)
}
func (m *SetAuthoritativeStorageRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SetAuthoritativeStorageRequest.Merge(m, src)
}
func (m *SetAuthoritativeStorageRequest) XXX_Size() int {
	return xxx_messageInfo_SetAuthoritativeStorageRequest.Size(m)
}
func (m *SetAuthoritativeStorageRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_SetAuthoritativeStorageRequest.DiscardUnknown(m)
}

var xxx_messageInfo_SetAuthoritativeStorageRequest proto.InternalMessageInfo

func (m *SetAuthoritativeStorageRequest) GetVirtualStorage() string {
	if m != nil {
		return m.VirtualStorage
	}
	return ""
}

func (m *SetAuthoritativeStorageRequest) GetRelativePath() string {
	if m != nil {
		return m.RelativePath
	}
	return ""
}

func (m *SetAuthoritativeStorageRequest) GetAuthoritativeStorage() string {
	if m != nil {
		return m.AuthoritativeStorage
	}
	return ""
}

type SetAuthoritativeStorageResponse struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *SetAuthoritativeStorageResponse) Reset()         { *m = SetAuthoritativeStorageResponse{} }
func (m *SetAuthoritativeStorageResponse) String() string { return proto.CompactTextString(m) }
func (*SetAuthoritativeStorageResponse) ProtoMessage()    {}
func (*SetAuthoritativeStorageResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_d32bf44842ead735, []int{1}
}

func (m *SetAuthoritativeStorageResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SetAuthoritativeStorageResponse.Unmarshal(m, b)
}
func (m *SetAuthoritativeStorageResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SetAuthoritativeStorageResponse.Marshal(b, m, deterministic)
}
func (m *SetAuthoritativeStorageResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SetAuthoritativeStorageResponse.Merge(m, src)
}
func (m *SetAuthoritativeStorageResponse) XXX_Size() int {
	return xxx_messageInfo_SetAuthoritativeStorageResponse.Size(m)
}
func (m *SetAuthoritativeStorageResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_SetAuthoritativeStorageResponse.DiscardUnknown(m)
}

var xxx_messageInfo_SetAuthoritativeStorageResponse proto.InternalMessageInfo

type DatalossCheckRequest struct {
	VirtualStorage string `protobuf:"bytes,1,opt,name=virtual_storage,json=virtualStorage,proto3" json:"virtual_storage,omitempty"`
	// include_partially_replicated decides whether to include repositories which are fully up to date
	// on the primary but are outdated on some secondaries. Such repositories are still writable and do
	// not suffer from data loss. The data on the primary is not fully replicated which increases the
	// chances of data loss following a failover.
	IncludePartiallyReplicated bool     `protobuf:"varint,2,opt,name=include_partially_replicated,json=includePartiallyReplicated,proto3" json:"include_partially_replicated,omitempty"`
	XXX_NoUnkeyedLiteral       struct{} `json:"-"`
	XXX_unrecognized           []byte   `json:"-"`
	XXX_sizecache              int32    `json:"-"`
}

func (m *DatalossCheckRequest) Reset()         { *m = DatalossCheckRequest{} }
func (m *DatalossCheckRequest) String() string { return proto.CompactTextString(m) }
func (*DatalossCheckRequest) ProtoMessage()    {}
func (*DatalossCheckRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_d32bf44842ead735, []int{2}
}

func (m *DatalossCheckRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_DatalossCheckRequest.Unmarshal(m, b)
}
func (m *DatalossCheckRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_DatalossCheckRequest.Marshal(b, m, deterministic)
}
func (m *DatalossCheckRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_DatalossCheckRequest.Merge(m, src)
}
func (m *DatalossCheckRequest) XXX_Size() int {
	return xxx_messageInfo_DatalossCheckRequest.Size(m)
}
func (m *DatalossCheckRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_DatalossCheckRequest.DiscardUnknown(m)
}

var xxx_messageInfo_DatalossCheckRequest proto.InternalMessageInfo

func (m *DatalossCheckRequest) GetVirtualStorage() string {
	if m != nil {
		return m.VirtualStorage
	}
	return ""
}

func (m *DatalossCheckRequest) GetIncludePartiallyReplicated() bool {
	if m != nil {
		return m.IncludePartiallyReplicated
	}
	return false
}

type DatalossCheckResponse struct {
	// current primary storage
	Primary string `protobuf:"bytes,1,opt,name=primary,proto3" json:"primary,omitempty"`
	// repositories with data loss
	Repositories         []*DatalossCheckResponse_Repository `protobuf:"bytes,2,rep,name=repositories,proto3" json:"repositories,omitempty"`
	XXX_NoUnkeyedLiteral struct{}                            `json:"-"`
	XXX_unrecognized     []byte                              `json:"-"`
	XXX_sizecache        int32                               `json:"-"`
}

func (m *DatalossCheckResponse) Reset()         { *m = DatalossCheckResponse{} }
func (m *DatalossCheckResponse) String() string { return proto.CompactTextString(m) }
func (*DatalossCheckResponse) ProtoMessage()    {}
func (*DatalossCheckResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_d32bf44842ead735, []int{3}
}

func (m *DatalossCheckResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_DatalossCheckResponse.Unmarshal(m, b)
}
func (m *DatalossCheckResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_DatalossCheckResponse.Marshal(b, m, deterministic)
}
func (m *DatalossCheckResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_DatalossCheckResponse.Merge(m, src)
}
func (m *DatalossCheckResponse) XXX_Size() int {
	return xxx_messageInfo_DatalossCheckResponse.Size(m)
}
func (m *DatalossCheckResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_DatalossCheckResponse.DiscardUnknown(m)
}

var xxx_messageInfo_DatalossCheckResponse proto.InternalMessageInfo

func (m *DatalossCheckResponse) GetPrimary() string {
	if m != nil {
		return m.Primary
	}
	return ""
}

func (m *DatalossCheckResponse) GetRepositories() []*DatalossCheckResponse_Repository {
	if m != nil {
		return m.Repositories
	}
	return nil
}

type DatalossCheckResponse_Repository struct {
	// relative path of the repository with outdated replicas
	RelativePath string `protobuf:"bytes,1,opt,name=relative_path,json=relativePath,proto3" json:"relative_path,omitempty"`
	// storages on which the repository is outdated
	Storages []*DatalossCheckResponse_Repository_Storage `protobuf:"bytes,2,rep,name=storages,proto3" json:"storages,omitempty"`
	// read_only indicates whether the repository is in read-only mode.
	ReadOnly             bool     `protobuf:"varint,3,opt,name=read_only,json=readOnly,proto3" json:"read_only,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *DatalossCheckResponse_Repository) Reset()         { *m = DatalossCheckResponse_Repository{} }
func (m *DatalossCheckResponse_Repository) String() string { return proto.CompactTextString(m) }
func (*DatalossCheckResponse_Repository) ProtoMessage()    {}
func (*DatalossCheckResponse_Repository) Descriptor() ([]byte, []int) {
	return fileDescriptor_d32bf44842ead735, []int{3, 0}
}

func (m *DatalossCheckResponse_Repository) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_DatalossCheckResponse_Repository.Unmarshal(m, b)
}
func (m *DatalossCheckResponse_Repository) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_DatalossCheckResponse_Repository.Marshal(b, m, deterministic)
}
func (m *DatalossCheckResponse_Repository) XXX_Merge(src proto.Message) {
	xxx_messageInfo_DatalossCheckResponse_Repository.Merge(m, src)
}
func (m *DatalossCheckResponse_Repository) XXX_Size() int {
	return xxx_messageInfo_DatalossCheckResponse_Repository.Size(m)
}
func (m *DatalossCheckResponse_Repository) XXX_DiscardUnknown() {
	xxx_messageInfo_DatalossCheckResponse_Repository.DiscardUnknown(m)
}

var xxx_messageInfo_DatalossCheckResponse_Repository proto.InternalMessageInfo

func (m *DatalossCheckResponse_Repository) GetRelativePath() string {
	if m != nil {
		return m.RelativePath
	}
	return ""
}

func (m *DatalossCheckResponse_Repository) GetStorages() []*DatalossCheckResponse_Repository_Storage {
	if m != nil {
		return m.Storages
	}
	return nil
}

func (m *DatalossCheckResponse_Repository) GetReadOnly() bool {
	if m != nil {
		return m.ReadOnly
	}
	return false
}

type DatalossCheckResponse_Repository_Storage struct {
	// name of the storage
	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	// behind_by indicates how many generations this storage is behind.
	BehindBy             int64    `protobuf:"varint,2,opt,name=behind_by,json=behindBy,proto3" json:"behind_by,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *DatalossCheckResponse_Repository_Storage) Reset() {
	*m = DatalossCheckResponse_Repository_Storage{}
}
func (m *DatalossCheckResponse_Repository_Storage) String() string { return proto.CompactTextString(m) }
func (*DatalossCheckResponse_Repository_Storage) ProtoMessage()    {}
func (*DatalossCheckResponse_Repository_Storage) Descriptor() ([]byte, []int) {
	return fileDescriptor_d32bf44842ead735, []int{3, 0, 0}
}

func (m *DatalossCheckResponse_Repository_Storage) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_DatalossCheckResponse_Repository_Storage.Unmarshal(m, b)
}
func (m *DatalossCheckResponse_Repository_Storage) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_DatalossCheckResponse_Repository_Storage.Marshal(b, m, deterministic)
}
func (m *DatalossCheckResponse_Repository_Storage) XXX_Merge(src proto.Message) {
	xxx_messageInfo_DatalossCheckResponse_Repository_Storage.Merge(m, src)
}
func (m *DatalossCheckResponse_Repository_Storage) XXX_Size() int {
	return xxx_messageInfo_DatalossCheckResponse_Repository_Storage.Size(m)
}
func (m *DatalossCheckResponse_Repository_Storage) XXX_DiscardUnknown() {
	xxx_messageInfo_DatalossCheckResponse_Repository_Storage.DiscardUnknown(m)
}

var xxx_messageInfo_DatalossCheckResponse_Repository_Storage proto.InternalMessageInfo

func (m *DatalossCheckResponse_Repository_Storage) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *DatalossCheckResponse_Repository_Storage) GetBehindBy() int64 {
	if m != nil {
		return m.BehindBy
	}
	return 0
}

type RepositoryReplicasRequest struct {
	Repository           *Repository `protobuf:"bytes,1,opt,name=repository,proto3" json:"repository,omitempty"`
	XXX_NoUnkeyedLiteral struct{}    `json:"-"`
	XXX_unrecognized     []byte      `json:"-"`
	XXX_sizecache        int32       `json:"-"`
}

func (m *RepositoryReplicasRequest) Reset()         { *m = RepositoryReplicasRequest{} }
func (m *RepositoryReplicasRequest) String() string { return proto.CompactTextString(m) }
func (*RepositoryReplicasRequest) ProtoMessage()    {}
func (*RepositoryReplicasRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_d32bf44842ead735, []int{4}
}

func (m *RepositoryReplicasRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_RepositoryReplicasRequest.Unmarshal(m, b)
}
func (m *RepositoryReplicasRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_RepositoryReplicasRequest.Marshal(b, m, deterministic)
}
func (m *RepositoryReplicasRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_RepositoryReplicasRequest.Merge(m, src)
}
func (m *RepositoryReplicasRequest) XXX_Size() int {
	return xxx_messageInfo_RepositoryReplicasRequest.Size(m)
}
func (m *RepositoryReplicasRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_RepositoryReplicasRequest.DiscardUnknown(m)
}

var xxx_messageInfo_RepositoryReplicasRequest proto.InternalMessageInfo

func (m *RepositoryReplicasRequest) GetRepository() *Repository {
	if m != nil {
		return m.Repository
	}
	return nil
}

type RepositoryReplicasResponse struct {
	Primary              *RepositoryReplicasResponse_RepositoryDetails   `protobuf:"bytes,1,opt,name=primary,proto3" json:"primary,omitempty"`
	Replicas             []*RepositoryReplicasResponse_RepositoryDetails `protobuf:"bytes,2,rep,name=replicas,proto3" json:"replicas,omitempty"`
	XXX_NoUnkeyedLiteral struct{}                                        `json:"-"`
	XXX_unrecognized     []byte                                          `json:"-"`
	XXX_sizecache        int32                                           `json:"-"`
}

func (m *RepositoryReplicasResponse) Reset()         { *m = RepositoryReplicasResponse{} }
func (m *RepositoryReplicasResponse) String() string { return proto.CompactTextString(m) }
func (*RepositoryReplicasResponse) ProtoMessage()    {}
func (*RepositoryReplicasResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_d32bf44842ead735, []int{5}
}

func (m *RepositoryReplicasResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_RepositoryReplicasResponse.Unmarshal(m, b)
}
func (m *RepositoryReplicasResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_RepositoryReplicasResponse.Marshal(b, m, deterministic)
}
func (m *RepositoryReplicasResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_RepositoryReplicasResponse.Merge(m, src)
}
func (m *RepositoryReplicasResponse) XXX_Size() int {
	return xxx_messageInfo_RepositoryReplicasResponse.Size(m)
}
func (m *RepositoryReplicasResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_RepositoryReplicasResponse.DiscardUnknown(m)
}

var xxx_messageInfo_RepositoryReplicasResponse proto.InternalMessageInfo

func (m *RepositoryReplicasResponse) GetPrimary() *RepositoryReplicasResponse_RepositoryDetails {
	if m != nil {
		return m.Primary
	}
	return nil
}

func (m *RepositoryReplicasResponse) GetReplicas() []*RepositoryReplicasResponse_RepositoryDetails {
	if m != nil {
		return m.Replicas
	}
	return nil
}

type RepositoryReplicasResponse_RepositoryDetails struct {
	Repository           *Repository `protobuf:"bytes,1,opt,name=repository,proto3" json:"repository,omitempty"`
	Checksum             string      `protobuf:"bytes,2,opt,name=checksum,proto3" json:"checksum,omitempty"`
	XXX_NoUnkeyedLiteral struct{}    `json:"-"`
	XXX_unrecognized     []byte      `json:"-"`
	XXX_sizecache        int32       `json:"-"`
}

func (m *RepositoryReplicasResponse_RepositoryDetails) Reset() {
	*m = RepositoryReplicasResponse_RepositoryDetails{}
}
func (m *RepositoryReplicasResponse_RepositoryDetails) String() string {
	return proto.CompactTextString(m)
}
func (*RepositoryReplicasResponse_RepositoryDetails) ProtoMessage() {}
func (*RepositoryReplicasResponse_RepositoryDetails) Descriptor() ([]byte, []int) {
	return fileDescriptor_d32bf44842ead735, []int{5, 0}
}

func (m *RepositoryReplicasResponse_RepositoryDetails) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_RepositoryReplicasResponse_RepositoryDetails.Unmarshal(m, b)
}
func (m *RepositoryReplicasResponse_RepositoryDetails) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_RepositoryReplicasResponse_RepositoryDetails.Marshal(b, m, deterministic)
}
func (m *RepositoryReplicasResponse_RepositoryDetails) XXX_Merge(src proto.Message) {
	xxx_messageInfo_RepositoryReplicasResponse_RepositoryDetails.Merge(m, src)
}
func (m *RepositoryReplicasResponse_RepositoryDetails) XXX_Size() int {
	return xxx_messageInfo_RepositoryReplicasResponse_RepositoryDetails.Size(m)
}
func (m *RepositoryReplicasResponse_RepositoryDetails) XXX_DiscardUnknown() {
	xxx_messageInfo_RepositoryReplicasResponse_RepositoryDetails.DiscardUnknown(m)
}

var xxx_messageInfo_RepositoryReplicasResponse_RepositoryDetails proto.InternalMessageInfo

func (m *RepositoryReplicasResponse_RepositoryDetails) GetRepository() *Repository {
	if m != nil {
		return m.Repository
	}
	return nil
}

func (m *RepositoryReplicasResponse_RepositoryDetails) GetChecksum() string {
	if m != nil {
		return m.Checksum
	}
	return ""
}

type ConsistencyCheckRequest struct {
	VirtualStorage string `protobuf:"bytes,1,opt,name=virtual_storage,json=virtualStorage,proto3" json:"virtual_storage,omitempty"`
	// The target storage is the storage you wish to check for inconsistencies
	// against a reference storage (typically the current primary).
	TargetStorage string `protobuf:"bytes,2,opt,name=target_storage,json=targetStorage,proto3" json:"target_storage,omitempty"`
	// Optionally provide a reference storage to compare the target storage
	// against. If a reference storage is omitted, the current primary will be
	// used.
	ReferenceStorage string `protobuf:"bytes,3,opt,name=reference_storage,json=referenceStorage,proto3" json:"reference_storage,omitempty"`
	// Be default, reconcilliation is enabled. Disabling reconcilliation will
	// make the request side-effect free.
	DisableReconcilliation bool     `protobuf:"varint,4,opt,name=disable_reconcilliation,json=disableReconcilliation,proto3" json:"disable_reconcilliation,omitempty"`
	XXX_NoUnkeyedLiteral   struct{} `json:"-"`
	XXX_unrecognized       []byte   `json:"-"`
	XXX_sizecache          int32    `json:"-"`
}

func (m *ConsistencyCheckRequest) Reset()         { *m = ConsistencyCheckRequest{} }
func (m *ConsistencyCheckRequest) String() string { return proto.CompactTextString(m) }
func (*ConsistencyCheckRequest) ProtoMessage()    {}
func (*ConsistencyCheckRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_d32bf44842ead735, []int{6}
}

func (m *ConsistencyCheckRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ConsistencyCheckRequest.Unmarshal(m, b)
}
func (m *ConsistencyCheckRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ConsistencyCheckRequest.Marshal(b, m, deterministic)
}
func (m *ConsistencyCheckRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ConsistencyCheckRequest.Merge(m, src)
}
func (m *ConsistencyCheckRequest) XXX_Size() int {
	return xxx_messageInfo_ConsistencyCheckRequest.Size(m)
}
func (m *ConsistencyCheckRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_ConsistencyCheckRequest.DiscardUnknown(m)
}

var xxx_messageInfo_ConsistencyCheckRequest proto.InternalMessageInfo

func (m *ConsistencyCheckRequest) GetVirtualStorage() string {
	if m != nil {
		return m.VirtualStorage
	}
	return ""
}

func (m *ConsistencyCheckRequest) GetTargetStorage() string {
	if m != nil {
		return m.TargetStorage
	}
	return ""
}

func (m *ConsistencyCheckRequest) GetReferenceStorage() string {
	if m != nil {
		return m.ReferenceStorage
	}
	return ""
}

func (m *ConsistencyCheckRequest) GetDisableReconcilliation() bool {
	if m != nil {
		return m.DisableReconcilliation
	}
	return false
}

type ConsistencyCheckResponse struct {
	RepoRelativePath  string `protobuf:"bytes,1,opt,name=repo_relative_path,json=repoRelativePath,proto3" json:"repo_relative_path,omitempty"`
	TargetChecksum    string `protobuf:"bytes,2,opt,name=target_checksum,json=targetChecksum,proto3" json:"target_checksum,omitempty"`
	ReferenceChecksum string `protobuf:"bytes,3,opt,name=reference_checksum,json=referenceChecksum,proto3" json:"reference_checksum,omitempty"`
	// If resync was enabled, then each inconsistency will schedule a replication
	// job. A replication ID is returned to track the corresponding job.
	ReplJobId uint64 `protobuf:"varint,4,opt,name=repl_job_id,json=replJobId,proto3" json:"repl_job_id,omitempty"`
	// If the reference storage was not specified, reply with the reference used
	ReferenceStorage     string   `protobuf:"bytes,5,opt,name=reference_storage,json=referenceStorage,proto3" json:"reference_storage,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ConsistencyCheckResponse) Reset()         { *m = ConsistencyCheckResponse{} }
func (m *ConsistencyCheckResponse) String() string { return proto.CompactTextString(m) }
func (*ConsistencyCheckResponse) ProtoMessage()    {}
func (*ConsistencyCheckResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_d32bf44842ead735, []int{7}
}

func (m *ConsistencyCheckResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ConsistencyCheckResponse.Unmarshal(m, b)
}
func (m *ConsistencyCheckResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ConsistencyCheckResponse.Marshal(b, m, deterministic)
}
func (m *ConsistencyCheckResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ConsistencyCheckResponse.Merge(m, src)
}
func (m *ConsistencyCheckResponse) XXX_Size() int {
	return xxx_messageInfo_ConsistencyCheckResponse.Size(m)
}
func (m *ConsistencyCheckResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_ConsistencyCheckResponse.DiscardUnknown(m)
}

var xxx_messageInfo_ConsistencyCheckResponse proto.InternalMessageInfo

func (m *ConsistencyCheckResponse) GetRepoRelativePath() string {
	if m != nil {
		return m.RepoRelativePath
	}
	return ""
}

func (m *ConsistencyCheckResponse) GetTargetChecksum() string {
	if m != nil {
		return m.TargetChecksum
	}
	return ""
}

func (m *ConsistencyCheckResponse) GetReferenceChecksum() string {
	if m != nil {
		return m.ReferenceChecksum
	}
	return ""
}

func (m *ConsistencyCheckResponse) GetReplJobId() uint64 {
	if m != nil {
		return m.ReplJobId
	}
	return 0
}

func (m *ConsistencyCheckResponse) GetReferenceStorage() string {
	if m != nil {
		return m.ReferenceStorage
	}
	return ""
}

func init() {
	proto.RegisterType((*SetAuthoritativeStorageRequest)(nil), "gitaly.SetAuthoritativeStorageRequest")
	proto.RegisterType((*SetAuthoritativeStorageResponse)(nil), "gitaly.SetAuthoritativeStorageResponse")
	proto.RegisterType((*DatalossCheckRequest)(nil), "gitaly.DatalossCheckRequest")
	proto.RegisterType((*DatalossCheckResponse)(nil), "gitaly.DatalossCheckResponse")
	proto.RegisterType((*DatalossCheckResponse_Repository)(nil), "gitaly.DatalossCheckResponse.Repository")
	proto.RegisterType((*DatalossCheckResponse_Repository_Storage)(nil), "gitaly.DatalossCheckResponse.Repository.Storage")
	proto.RegisterType((*RepositoryReplicasRequest)(nil), "gitaly.RepositoryReplicasRequest")
	proto.RegisterType((*RepositoryReplicasResponse)(nil), "gitaly.RepositoryReplicasResponse")
	proto.RegisterType((*RepositoryReplicasResponse_RepositoryDetails)(nil), "gitaly.RepositoryReplicasResponse.RepositoryDetails")
	proto.RegisterType((*ConsistencyCheckRequest)(nil), "gitaly.ConsistencyCheckRequest")
	proto.RegisterType((*ConsistencyCheckResponse)(nil), "gitaly.ConsistencyCheckResponse")
}

func init() { proto.RegisterFile("praefect.proto", fileDescriptor_d32bf44842ead735) }

var fileDescriptor_d32bf44842ead735 = []byte{
	// 745 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x9c, 0x55, 0xcd, 0x6e, 0xd3, 0x4a,
	0x14, 0x96, 0x93, 0xdc, 0x36, 0x3d, 0xe9, 0xef, 0xdc, 0xf6, 0x36, 0xd7, 0x94, 0xfe, 0x18, 0x41,
	0x23, 0x41, 0x93, 0x2a, 0x45, 0x42, 0x62, 0x05, 0x6d, 0x37, 0x45, 0x15, 0x8d, 0xdc, 0x05, 0x12,
	0x2c, 0xac, 0xb1, 0x3d, 0x4d, 0xa6, 0x4c, 0x3c, 0x66, 0x3c, 0xa9, 0xe4, 0x25, 0x6b, 0x1e, 0x80,
	0x07, 0xe8, 0x03, 0xb1, 0x45, 0xbc, 0x03, 0x12, 0x8f, 0x80, 0x6c, 0xcf, 0xb8, 0xf9, 0x71, 0x5a,
	0xe8, 0xce, 0x3e, 0xe7, 0x3b, 0xdf, 0xcc, 0xf9, 0xbe, 0x33, 0x33, 0xb0, 0x18, 0x0a, 0x4c, 0x2e,
	0x88, 0x27, 0x9b, 0xa1, 0xe0, 0x92, 0xa3, 0x99, 0x2e, 0x95, 0x98, 0xc5, 0x26, 0x30, 0x1a, 0xa8,
	0x98, 0x39, 0x1f, 0xf5, 0xb0, 0x20, 0x7e, 0xf6, 0x67, 0x5d, 0x1b, 0xb0, 0x79, 0x4e, 0xe4, 0xeb,
	0x81, 0xec, 0x71, 0x41, 0x25, 0x96, 0xf4, 0x8a, 0x9c, 0x4b, 0x2e, 0x70, 0x97, 0xd8, 0xe4, 0xd3,
	0x80, 0x44, 0x12, 0xed, 0xc2, 0xd2, 0x15, 0x15, 0x72, 0x80, 0x99, 0x13, 0x65, 0x99, 0xba, 0xb1,
	0x6d, 0x34, 0xe6, 0xec, 0x45, 0x15, 0x56, 0x78, 0xf4, 0x08, 0x16, 0x04, 0x61, 0x29, 0x85, 0x13,
	0x62, 0xd9, 0xab, 0x97, 0x52, 0xd8, 0xbc, 0x0e, 0x76, 0xb0, 0xec, 0xa1, 0x03, 0x58, 0xc3, 0xc3,
	0x8b, 0xe5, 0x9c, 0xe5, 0x14, 0xbc, 0x8a, 0x0b, 0x76, 0x62, 0xed, 0xc0, 0xd6, 0xd4, 0x4d, 0x46,
	0x21, 0x0f, 0x22, 0x62, 0x7d, 0x36, 0x60, 0xf5, 0x18, 0x4b, 0xcc, 0x78, 0x14, 0x1d, 0xf5, 0x88,
	0xf7, 0xf1, 0xaf, 0xb7, 0xff, 0x0a, 0x36, 0x68, 0xe0, 0xb1, 0x81, 0x9f, 0xec, 0x5e, 0x48, 0x8a,
	0x19, 0x8b, 0x1d, 0x41, 0x42, 0x46, 0x3d, 0x2c, 0x89, 0x9f, 0x76, 0x53, 0xb5, 0x4d, 0x85, 0xe9,
	0x68, 0x88, 0x9d, 0x23, 0xac, 0x1f, 0x25, 0x58, 0x1b, 0xdb, 0x43, 0xb6, 0x3b, 0x54, 0x87, 0xd9,
	0x50, 0xd0, 0x3e, 0x16, 0xb1, 0x5a, 0x5c, 0xff, 0xa2, 0x53, 0x98, 0x17, 0x24, 0xe4, 0x11, 0x95,
	0x5c, 0x50, 0x12, 0xd5, 0x4b, 0xdb, 0xe5, 0x46, 0xad, 0xdd, 0x68, 0x66, 0xce, 0x35, 0x0b, 0xe9,
	0x9a, 0xb6, 0xae, 0x88, 0xed, 0x91, 0x6a, 0xf3, 0xbb, 0x01, 0x70, 0x93, 0x9c, 0x74, 0xc4, 0x28,
	0x70, 0xe4, 0x14, 0xaa, 0x4a, 0x18, 0xbd, 0xfa, 0xfe, 0x9f, 0xae, 0xde, 0xd4, 0x2e, 0xe4, 0x0c,
	0xe8, 0x01, 0xcc, 0x09, 0x82, 0x7d, 0x87, 0x07, 0x2c, 0x4e, 0x3d, 0xad, 0xda, 0xd5, 0x24, 0x70,
	0x16, 0xb0, 0xd8, 0x7c, 0x09, 0xb3, 0x5a, 0x6d, 0x04, 0x95, 0x00, 0xf7, 0xb5, 0x17, 0xe9, 0x77,
	0x52, 0xeb, 0x92, 0x1e, 0x0d, 0x7c, 0xc7, 0x8d, 0x53, 0xb9, 0xcb, 0x76, 0x35, 0x0b, 0x1c, 0xc6,
	0xd6, 0x19, 0xfc, 0x3f, 0xd4, 0x76, 0x26, 0x7a, 0xa4, 0x4d, 0x6e, 0x03, 0xe4, 0x3a, 0x64, 0x12,
	0xd7, 0xda, 0x48, 0x77, 0x31, 0x54, 0x36, 0x84, 0xb2, 0xae, 0x4b, 0x60, 0x16, 0x31, 0x2a, 0xcb,
	0xde, 0x8e, 0x5a, 0x56, 0x6b, 0x3f, 0x2f, 0xe0, 0x1b, 0x2b, 0x1a, 0x4a, 0x1d, 0x13, 0x89, 0x29,
	0x8b, 0x6e, 0x8c, 0xee, 0x40, 0x55, 0x0d, 0x93, 0x96, 0xf9, 0x7e, 0x84, 0x39, 0x8b, 0xe9, 0xc1,
	0xca, 0x44, 0xfa, 0x3e, 0x4a, 0x20, 0x13, 0xaa, 0x5e, 0xe2, 0x70, 0x34, 0xe8, 0xab, 0x33, 0x9b,
	0xff, 0x5b, 0xdf, 0x0c, 0x58, 0x3f, 0xe2, 0x41, 0x44, 0x23, 0x49, 0x02, 0x2f, 0xbe, 0xdf, 0xd1,
	0x7a, 0x0c, 0x8b, 0x12, 0x8b, 0x2e, 0x91, 0x39, 0x2e, 0x5b, 0x66, 0x21, 0x8b, 0x6a, 0xd8, 0x53,
	0x58, 0x11, 0xe4, 0x82, 0x08, 0x12, 0x78, 0xe3, 0xf7, 0xc2, 0x72, 0x9e, 0xd0, 0xe0, 0x17, 0xb0,
	0xee, 0xd3, 0x08, 0xbb, 0x8c, 0x38, 0x82, 0x78, 0x3c, 0xf0, 0x28, 0x63, 0x14, 0x4b, 0xca, 0x83,
	0x7a, 0x25, 0x1d, 0xbb, 0xff, 0x54, 0xda, 0x1e, 0xcd, 0x5a, 0x3f, 0x0d, 0xa8, 0x4f, 0x76, 0xa4,
	0x5c, 0x7f, 0x06, 0x28, 0x11, 0xc6, 0x29, 0x3a, 0x36, 0xcb, 0x49, 0xc6, 0x1e, 0x3e, 0x3a, 0xbb,
	0xb0, 0xa4, 0xfa, 0x1a, 0xd3, 0x4f, 0xb5, 0x7b, 0xa4, 0xa2, 0x68, 0x2f, 0xa1, 0xd5, 0x9d, 0xe5,
	0xd8, 0xac, 0xb5, 0x9b, 0x9e, 0x73, 0xf8, 0x26, 0xd4, 0x12, 0x97, 0x9d, 0x4b, 0xee, 0x3a, 0xd4,
	0x4f, 0xfb, 0xa9, 0xd8, 0x73, 0x49, 0xe8, 0x0d, 0x77, 0x4f, 0xfc, 0x62, 0xa1, 0xfe, 0x29, 0x16,
	0xaa, 0xfd, 0xa5, 0x0c, 0xff, 0x76, 0xd4, 0xbb, 0x70, 0x12, 0x5c, 0xf0, 0x73, 0x22, 0xae, 0xa8,
	0x47, 0xd0, 0x07, 0x40, 0x93, 0x83, 0x87, 0x76, 0x6e, 0x1b, 0xca, 0xd4, 0x76, 0xd3, 0xba, 0x7b,
	0x6e, 0xd1, 0x3b, 0x58, 0x1e, 0xd7, 0x18, 0x6d, 0xe9, 0xba, 0x29, 0xf3, 0x64, 0x6e, 0x4f, 0x07,
	0x64, 0xb4, 0xfb, 0x06, 0x3a, 0x85, 0x85, 0x91, 0x5b, 0x09, 0x6d, 0x4c, 0xb9, 0xac, 0x32, 0xca,
	0x87, 0xb7, 0x5e, 0x65, 0xe8, 0x12, 0xd6, 0xa7, 0x3c, 0x2c, 0xe8, 0x89, 0xae, 0xbc, 0xfd, 0x79,
	0x34, 0x77, 0xef, 0xc4, 0x65, 0x6b, 0x99, 0x95, 0x5f, 0x5f, 0x1b, 0xc6, 0xe1, 0xfe, 0xfb, 0x04,
	0xcf, 0xb0, 0xdb, 0xf4, 0x78, 0xbf, 0x95, 0x7d, 0xee, 0x71, 0xd1, 0x6d, 0x65, 0x2c, 0xad, 0xf4,
	0x59, 0x6e, 0x75, 0xb9, 0xfa, 0x0f, 0x5d, 0x77, 0x26, 0x0d, 0x1d, 0xfc, 0x0e, 0x00, 0x00, 0xff,
	0xff, 0xaf, 0x1f, 0xf3, 0xad, 0xdd, 0x07, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// PraefectInfoServiceClient is the client API for PraefectInfoService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type PraefectInfoServiceClient interface {
	RepositoryReplicas(ctx context.Context, in *RepositoryReplicasRequest, opts ...grpc.CallOption) (*RepositoryReplicasResponse, error)
	// ConsistencyCheck will perform a consistency check on the requested
	// virtual storage backend. A stream of repository statuses will be sent
	// back indicating which repos are consistent with the primary and which ones
	// need repair.
	ConsistencyCheck(ctx context.Context, in *ConsistencyCheckRequest, opts ...grpc.CallOption) (PraefectInfoService_ConsistencyCheckClient, error)
	// DatalossCheck checks for outdated repository replicas.
	DatalossCheck(ctx context.Context, in *DatalossCheckRequest, opts ...grpc.CallOption) (*DatalossCheckResponse, error)
	// SetAuthoritativeStorage sets the authoritative storage for a repository on a given virtual storage.
	// This causes the current version of the repository on the authoritative storage to be considered the
	// latest and overwrite any other version on the virtual storage.
	SetAuthoritativeStorage(ctx context.Context, in *SetAuthoritativeStorageRequest, opts ...grpc.CallOption) (*SetAuthoritativeStorageResponse, error)
}

type praefectInfoServiceClient struct {
	cc *grpc.ClientConn
}

func NewPraefectInfoServiceClient(cc *grpc.ClientConn) PraefectInfoServiceClient {
	return &praefectInfoServiceClient{cc}
}

func (c *praefectInfoServiceClient) RepositoryReplicas(ctx context.Context, in *RepositoryReplicasRequest, opts ...grpc.CallOption) (*RepositoryReplicasResponse, error) {
	out := new(RepositoryReplicasResponse)
	err := c.cc.Invoke(ctx, "/gitaly.PraefectInfoService/RepositoryReplicas", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *praefectInfoServiceClient) ConsistencyCheck(ctx context.Context, in *ConsistencyCheckRequest, opts ...grpc.CallOption) (PraefectInfoService_ConsistencyCheckClient, error) {
	stream, err := c.cc.NewStream(ctx, &_PraefectInfoService_serviceDesc.Streams[0], "/gitaly.PraefectInfoService/ConsistencyCheck", opts...)
	if err != nil {
		return nil, err
	}
	x := &praefectInfoServiceConsistencyCheckClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type PraefectInfoService_ConsistencyCheckClient interface {
	Recv() (*ConsistencyCheckResponse, error)
	grpc.ClientStream
}

type praefectInfoServiceConsistencyCheckClient struct {
	grpc.ClientStream
}

func (x *praefectInfoServiceConsistencyCheckClient) Recv() (*ConsistencyCheckResponse, error) {
	m := new(ConsistencyCheckResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *praefectInfoServiceClient) DatalossCheck(ctx context.Context, in *DatalossCheckRequest, opts ...grpc.CallOption) (*DatalossCheckResponse, error) {
	out := new(DatalossCheckResponse)
	err := c.cc.Invoke(ctx, "/gitaly.PraefectInfoService/DatalossCheck", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *praefectInfoServiceClient) SetAuthoritativeStorage(ctx context.Context, in *SetAuthoritativeStorageRequest, opts ...grpc.CallOption) (*SetAuthoritativeStorageResponse, error) {
	out := new(SetAuthoritativeStorageResponse)
	err := c.cc.Invoke(ctx, "/gitaly.PraefectInfoService/SetAuthoritativeStorage", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// PraefectInfoServiceServer is the server API for PraefectInfoService service.
type PraefectInfoServiceServer interface {
	RepositoryReplicas(context.Context, *RepositoryReplicasRequest) (*RepositoryReplicasResponse, error)
	// ConsistencyCheck will perform a consistency check on the requested
	// virtual storage backend. A stream of repository statuses will be sent
	// back indicating which repos are consistent with the primary and which ones
	// need repair.
	ConsistencyCheck(*ConsistencyCheckRequest, PraefectInfoService_ConsistencyCheckServer) error
	// DatalossCheck checks for outdated repository replicas.
	DatalossCheck(context.Context, *DatalossCheckRequest) (*DatalossCheckResponse, error)
	// SetAuthoritativeStorage sets the authoritative storage for a repository on a given virtual storage.
	// This causes the current version of the repository on the authoritative storage to be considered the
	// latest and overwrite any other version on the virtual storage.
	SetAuthoritativeStorage(context.Context, *SetAuthoritativeStorageRequest) (*SetAuthoritativeStorageResponse, error)
}

// UnimplementedPraefectInfoServiceServer can be embedded to have forward compatible implementations.
type UnimplementedPraefectInfoServiceServer struct {
}

func (*UnimplementedPraefectInfoServiceServer) RepositoryReplicas(ctx context.Context, req *RepositoryReplicasRequest) (*RepositoryReplicasResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RepositoryReplicas not implemented")
}
func (*UnimplementedPraefectInfoServiceServer) ConsistencyCheck(req *ConsistencyCheckRequest, srv PraefectInfoService_ConsistencyCheckServer) error {
	return status.Errorf(codes.Unimplemented, "method ConsistencyCheck not implemented")
}
func (*UnimplementedPraefectInfoServiceServer) DatalossCheck(ctx context.Context, req *DatalossCheckRequest) (*DatalossCheckResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DatalossCheck not implemented")
}
func (*UnimplementedPraefectInfoServiceServer) SetAuthoritativeStorage(ctx context.Context, req *SetAuthoritativeStorageRequest) (*SetAuthoritativeStorageResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SetAuthoritativeStorage not implemented")
}

func RegisterPraefectInfoServiceServer(s *grpc.Server, srv PraefectInfoServiceServer) {
	s.RegisterService(&_PraefectInfoService_serviceDesc, srv)
}

func _PraefectInfoService_RepositoryReplicas_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RepositoryReplicasRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PraefectInfoServiceServer).RepositoryReplicas(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/gitaly.PraefectInfoService/RepositoryReplicas",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PraefectInfoServiceServer).RepositoryReplicas(ctx, req.(*RepositoryReplicasRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _PraefectInfoService_ConsistencyCheck_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(ConsistencyCheckRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(PraefectInfoServiceServer).ConsistencyCheck(m, &praefectInfoServiceConsistencyCheckServer{stream})
}

type PraefectInfoService_ConsistencyCheckServer interface {
	Send(*ConsistencyCheckResponse) error
	grpc.ServerStream
}

type praefectInfoServiceConsistencyCheckServer struct {
	grpc.ServerStream
}

func (x *praefectInfoServiceConsistencyCheckServer) Send(m *ConsistencyCheckResponse) error {
	return x.ServerStream.SendMsg(m)
}

func _PraefectInfoService_DatalossCheck_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DatalossCheckRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PraefectInfoServiceServer).DatalossCheck(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/gitaly.PraefectInfoService/DatalossCheck",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PraefectInfoServiceServer).DatalossCheck(ctx, req.(*DatalossCheckRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _PraefectInfoService_SetAuthoritativeStorage_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SetAuthoritativeStorageRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PraefectInfoServiceServer).SetAuthoritativeStorage(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/gitaly.PraefectInfoService/SetAuthoritativeStorage",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PraefectInfoServiceServer).SetAuthoritativeStorage(ctx, req.(*SetAuthoritativeStorageRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _PraefectInfoService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "gitaly.PraefectInfoService",
	HandlerType: (*PraefectInfoServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "RepositoryReplicas",
			Handler:    _PraefectInfoService_RepositoryReplicas_Handler,
		},
		{
			MethodName: "DatalossCheck",
			Handler:    _PraefectInfoService_DatalossCheck_Handler,
		},
		{
			MethodName: "SetAuthoritativeStorage",
			Handler:    _PraefectInfoService_SetAuthoritativeStorage_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "ConsistencyCheck",
			Handler:       _PraefectInfoService_ConsistencyCheck_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "praefect.proto",
}
