// Code generated by protoc-gen-go. DO NOT EDIT.
// source: shared.proto

package gitalypb

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import google_protobuf "github.com/golang/protobuf/ptypes/timestamp"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

type Repository struct {
	StorageName  string `protobuf:"bytes,2,opt,name=storage_name,json=storageName" json:"storage_name,omitempty"`
	RelativePath string `protobuf:"bytes,3,opt,name=relative_path,json=relativePath" json:"relative_path,omitempty"`
	// Sets the GIT_OBJECT_DIRECTORY envvar on git commands to the value of this field.
	// It influences the object storage directory the SHA1 directories are created underneath.
	GitObjectDirectory string `protobuf:"bytes,4,opt,name=git_object_directory,json=gitObjectDirectory" json:"git_object_directory,omitempty"`
	// Sets the GIT_ALTERNATE_OBJECT_DIRECTORIES envvar on git commands to the values of this field.
	// It influences the list of Git object directories which can be used to search for Git objects.
	GitAlternateObjectDirectories []string `protobuf:"bytes,5,rep,name=git_alternate_object_directories,json=gitAlternateObjectDirectories" json:"git_alternate_object_directories,omitempty"`
	// Used in callbacks to GitLab so that it knows what repository the event is
	// associated with. May be left empty on RPC's that do not perform callbacks.
	GlRepository string `protobuf:"bytes,6,opt,name=gl_repository,json=glRepository" json:"gl_repository,omitempty"`
}

func (m *Repository) Reset()                    { *m = Repository{} }
func (m *Repository) String() string            { return proto.CompactTextString(m) }
func (*Repository) ProtoMessage()               {}
func (*Repository) Descriptor() ([]byte, []int) { return fileDescriptor12, []int{0} }

func (m *Repository) GetStorageName() string {
	if m != nil {
		return m.StorageName
	}
	return ""
}

func (m *Repository) GetRelativePath() string {
	if m != nil {
		return m.RelativePath
	}
	return ""
}

func (m *Repository) GetGitObjectDirectory() string {
	if m != nil {
		return m.GitObjectDirectory
	}
	return ""
}

func (m *Repository) GetGitAlternateObjectDirectories() []string {
	if m != nil {
		return m.GitAlternateObjectDirectories
	}
	return nil
}

func (m *Repository) GetGlRepository() string {
	if m != nil {
		return m.GlRepository
	}
	return ""
}

// Corresponds to Gitlab::Git::Commit
type GitCommit struct {
	Id        string        `protobuf:"bytes,1,opt,name=id" json:"id,omitempty"`
	Subject   []byte        `protobuf:"bytes,2,opt,name=subject,proto3" json:"subject,omitempty"`
	Body      []byte        `protobuf:"bytes,3,opt,name=body,proto3" json:"body,omitempty"`
	Author    *CommitAuthor `protobuf:"bytes,4,opt,name=author" json:"author,omitempty"`
	Committer *CommitAuthor `protobuf:"bytes,5,opt,name=committer" json:"committer,omitempty"`
	ParentIds []string      `protobuf:"bytes,6,rep,name=parent_ids,json=parentIds" json:"parent_ids,omitempty"`
	// If body exceeds a certain threshold, it will be nullified,
	// but its size will be set in body_size so we can know if
	// a commit had a body in the first place.
	BodySize int64 `protobuf:"varint,7,opt,name=body_size,json=bodySize" json:"body_size,omitempty"`
}

func (m *GitCommit) Reset()                    { *m = GitCommit{} }
func (m *GitCommit) String() string            { return proto.CompactTextString(m) }
func (*GitCommit) ProtoMessage()               {}
func (*GitCommit) Descriptor() ([]byte, []int) { return fileDescriptor12, []int{1} }

func (m *GitCommit) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *GitCommit) GetSubject() []byte {
	if m != nil {
		return m.Subject
	}
	return nil
}

func (m *GitCommit) GetBody() []byte {
	if m != nil {
		return m.Body
	}
	return nil
}

func (m *GitCommit) GetAuthor() *CommitAuthor {
	if m != nil {
		return m.Author
	}
	return nil
}

func (m *GitCommit) GetCommitter() *CommitAuthor {
	if m != nil {
		return m.Committer
	}
	return nil
}

func (m *GitCommit) GetParentIds() []string {
	if m != nil {
		return m.ParentIds
	}
	return nil
}

func (m *GitCommit) GetBodySize() int64 {
	if m != nil {
		return m.BodySize
	}
	return 0
}

type CommitAuthor struct {
	Name  []byte                     `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Email []byte                     `protobuf:"bytes,2,opt,name=email,proto3" json:"email,omitempty"`
	Date  *google_protobuf.Timestamp `protobuf:"bytes,3,opt,name=date" json:"date,omitempty"`
}

func (m *CommitAuthor) Reset()                    { *m = CommitAuthor{} }
func (m *CommitAuthor) String() string            { return proto.CompactTextString(m) }
func (*CommitAuthor) ProtoMessage()               {}
func (*CommitAuthor) Descriptor() ([]byte, []int) { return fileDescriptor12, []int{2} }

func (m *CommitAuthor) GetName() []byte {
	if m != nil {
		return m.Name
	}
	return nil
}

func (m *CommitAuthor) GetEmail() []byte {
	if m != nil {
		return m.Email
	}
	return nil
}

func (m *CommitAuthor) GetDate() *google_protobuf.Timestamp {
	if m != nil {
		return m.Date
	}
	return nil
}

type ExitStatus struct {
	Value int32 `protobuf:"varint,1,opt,name=value" json:"value,omitempty"`
}

func (m *ExitStatus) Reset()                    { *m = ExitStatus{} }
func (m *ExitStatus) String() string            { return proto.CompactTextString(m) }
func (*ExitStatus) ProtoMessage()               {}
func (*ExitStatus) Descriptor() ([]byte, []int) { return fileDescriptor12, []int{3} }

func (m *ExitStatus) GetValue() int32 {
	if m != nil {
		return m.Value
	}
	return 0
}

// Corresponds to Gitlab::Git::Branch
type Branch struct {
	Name         []byte     `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	TargetCommit *GitCommit `protobuf:"bytes,2,opt,name=target_commit,json=targetCommit" json:"target_commit,omitempty"`
}

func (m *Branch) Reset()                    { *m = Branch{} }
func (m *Branch) String() string            { return proto.CompactTextString(m) }
func (*Branch) ProtoMessage()               {}
func (*Branch) Descriptor() ([]byte, []int) { return fileDescriptor12, []int{4} }

func (m *Branch) GetName() []byte {
	if m != nil {
		return m.Name
	}
	return nil
}

func (m *Branch) GetTargetCommit() *GitCommit {
	if m != nil {
		return m.TargetCommit
	}
	return nil
}

type Tag struct {
	Name         []byte     `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Id           string     `protobuf:"bytes,2,opt,name=id" json:"id,omitempty"`
	TargetCommit *GitCommit `protobuf:"bytes,3,opt,name=target_commit,json=targetCommit" json:"target_commit,omitempty"`
	// If message exceeds a certain threshold, it will be nullified,
	// but its size will be set in message_size so we can know if
	// a tag had a message in the first place.
	Message     []byte        `protobuf:"bytes,4,opt,name=message,proto3" json:"message,omitempty"`
	MessageSize int64         `protobuf:"varint,5,opt,name=message_size,json=messageSize" json:"message_size,omitempty"`
	Tagger      *CommitAuthor `protobuf:"bytes,6,opt,name=tagger" json:"tagger,omitempty"`
}

func (m *Tag) Reset()                    { *m = Tag{} }
func (m *Tag) String() string            { return proto.CompactTextString(m) }
func (*Tag) ProtoMessage()               {}
func (*Tag) Descriptor() ([]byte, []int) { return fileDescriptor12, []int{5} }

func (m *Tag) GetName() []byte {
	if m != nil {
		return m.Name
	}
	return nil
}

func (m *Tag) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *Tag) GetTargetCommit() *GitCommit {
	if m != nil {
		return m.TargetCommit
	}
	return nil
}

func (m *Tag) GetMessage() []byte {
	if m != nil {
		return m.Message
	}
	return nil
}

func (m *Tag) GetMessageSize() int64 {
	if m != nil {
		return m.MessageSize
	}
	return 0
}

func (m *Tag) GetTagger() *CommitAuthor {
	if m != nil {
		return m.Tagger
	}
	return nil
}

type User struct {
	GlId       string `protobuf:"bytes,1,opt,name=gl_id,json=glId" json:"gl_id,omitempty"`
	Name       []byte `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	Email      []byte `protobuf:"bytes,3,opt,name=email,proto3" json:"email,omitempty"`
	GlUsername string `protobuf:"bytes,4,opt,name=gl_username,json=glUsername" json:"gl_username,omitempty"`
}

func (m *User) Reset()                    { *m = User{} }
func (m *User) String() string            { return proto.CompactTextString(m) }
func (*User) ProtoMessage()               {}
func (*User) Descriptor() ([]byte, []int) { return fileDescriptor12, []int{6} }

func (m *User) GetGlId() string {
	if m != nil {
		return m.GlId
	}
	return ""
}

func (m *User) GetName() []byte {
	if m != nil {
		return m.Name
	}
	return nil
}

func (m *User) GetEmail() []byte {
	if m != nil {
		return m.Email
	}
	return nil
}

func (m *User) GetGlUsername() string {
	if m != nil {
		return m.GlUsername
	}
	return ""
}

func init() {
	proto.RegisterType((*Repository)(nil), "gitaly.Repository")
	proto.RegisterType((*GitCommit)(nil), "gitaly.GitCommit")
	proto.RegisterType((*CommitAuthor)(nil), "gitaly.CommitAuthor")
	proto.RegisterType((*ExitStatus)(nil), "gitaly.ExitStatus")
	proto.RegisterType((*Branch)(nil), "gitaly.Branch")
	proto.RegisterType((*Tag)(nil), "gitaly.Tag")
	proto.RegisterType((*User)(nil), "gitaly.User")
}

func init() { proto.RegisterFile("shared.proto", fileDescriptor12) }

var fileDescriptor12 = []byte{
	// 576 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x53, 0x41, 0x6f, 0xd4, 0x3c,
	0x10, 0x55, 0xb2, 0xd9, 0x6d, 0x77, 0x36, 0xfd, 0xf4, 0x61, 0x7a, 0x88, 0x8a, 0xaa, 0x2e, 0xe1,
	0xd2, 0x03, 0x4a, 0xd1, 0x22, 0x71, 0x6f, 0x01, 0x55, 0xe5, 0x00, 0xc8, 0x6d, 0x2f, 0x5c, 0x22,
	0xef, 0x66, 0xea, 0x35, 0x72, 0x36, 0x2b, 0x7b, 0x52, 0xd1, 0x9e, 0xf8, 0x81, 0xfc, 0x0f, 0xfe,
	0x06, 0xb2, 0x9d, 0x6c, 0x0b, 0x14, 0xc4, 0xcd, 0xf3, 0xfc, 0x3c, 0x9e, 0xe7, 0xf7, 0x0c, 0xa9,
	0x5d, 0x0a, 0x83, 0x55, 0xb1, 0x36, 0x0d, 0x35, 0x6c, 0x24, 0x15, 0x09, 0x7d, 0xb3, 0x77, 0x20,
	0x9b, 0x46, 0x6a, 0x3c, 0xf2, 0xe8, 0xbc, 0xbd, 0x3a, 0x22, 0x55, 0xa3, 0x25, 0x51, 0xaf, 0x03,
	0x31, 0xff, 0x1a, 0x03, 0x70, 0x5c, 0x37, 0x56, 0x51, 0x63, 0x6e, 0xd8, 0x53, 0x48, 0x2d, 0x35,
	0x46, 0x48, 0x2c, 0x57, 0xa2, 0xc6, 0x2c, 0x9e, 0x46, 0x87, 0x63, 0x3e, 0xe9, 0xb0, 0xf7, 0xa2,
	0x46, 0xf6, 0x0c, 0x76, 0x0c, 0x6a, 0x41, 0xea, 0x1a, 0xcb, 0xb5, 0xa0, 0x65, 0x36, 0xf0, 0x9c,
	0xb4, 0x07, 0x3f, 0x0a, 0x5a, 0xb2, 0x17, 0xb0, 0x2b, 0x15, 0x95, 0xcd, 0xfc, 0x33, 0x2e, 0xa8,
	0xac, 0x94, 0xc1, 0x85, 0xeb, 0x9f, 0x25, 0x9e, 0xcb, 0xa4, 0xa2, 0x0f, 0x7e, 0xeb, 0x4d, 0xbf,
	0xc3, 0x4e, 0x61, 0xea, 0x4e, 0x08, 0x4d, 0x68, 0x56, 0x82, 0xf0, 0xd7, 0xb3, 0x0a, 0x6d, 0x36,
	0x9c, 0x0e, 0x0e, 0xc7, 0x7c, 0x5f, 0x2a, 0x3a, 0xee, 0x69, 0x3f, 0xb7, 0x51, 0x68, 0xdd, 0x7c,
	0x52, 0x97, 0x66, 0xa3, 0x29, 0x1b, 0x85, 0xf9, 0xa4, 0xbe, 0xd3, 0xf9, 0x2e, 0xd9, 0x8e, 0xfe,
	0x8f, 0x79, 0xe2, 0xe6, 0xcf, 0xbf, 0x47, 0x30, 0x3e, 0x55, 0xf4, 0xba, 0xa9, 0x6b, 0x45, 0xec,
	0x3f, 0x88, 0x55, 0x95, 0x45, 0xfe, 0x4c, 0xac, 0x2a, 0x96, 0xc1, 0x96, 0x6d, 0xfd, 0x25, 0xfe,
	0x31, 0x52, 0xde, 0x97, 0x8c, 0x41, 0x32, 0x6f, 0xaa, 0x1b, 0xaf, 0x3f, 0xe5, 0x7e, 0xcd, 0x9e,
	0xc3, 0x48, 0xb4, 0xb4, 0x6c, 0x8c, 0x57, 0x3a, 0x99, 0xed, 0x16, 0xc1, 0x88, 0x22, 0x74, 0x3f,
	0xf6, 0x7b, 0xbc, 0xe3, 0xb0, 0x19, 0x8c, 0x17, 0x1e, 0x27, 0x34, 0xd9, 0xf0, 0x2f, 0x07, 0xee,
	0x68, 0x6c, 0x1f, 0x60, 0x2d, 0x0c, 0xae, 0xa8, 0x54, 0x95, 0xcd, 0x46, 0xfe, 0x45, 0xc6, 0x01,
	0x39, 0xab, 0x2c, 0x7b, 0x02, 0x63, 0x37, 0x48, 0x69, 0xd5, 0x2d, 0x66, 0x5b, 0xd3, 0xe8, 0x70,
	0xc0, 0xb7, 0x1d, 0x70, 0xae, 0x6e, 0x31, 0x5f, 0x42, 0x7a, 0xbf, 0xad, 0x53, 0xe0, 0x5d, 0x8e,
	0x82, 0x02, 0xb7, 0x66, 0xbb, 0x30, 0xc4, 0x5a, 0x28, 0xdd, 0xa9, 0x0d, 0x05, 0x2b, 0x20, 0xa9,
	0x04, 0xa1, 0xd7, 0x3a, 0x99, 0xed, 0x15, 0x21, 0x56, 0x45, 0x1f, 0xab, 0xe2, 0xa2, 0x8f, 0x15,
	0xf7, 0xbc, 0x3c, 0x07, 0x78, 0xfb, 0x45, 0xd1, 0x39, 0x09, 0x6a, 0xad, 0xeb, 0x79, 0x2d, 0x74,
	0x1b, 0x2e, 0x1a, 0xf2, 0x50, 0xe4, 0x17, 0x30, 0x3a, 0x31, 0x62, 0xb5, 0x58, 0x3e, 0x38, 0xc7,
	0x2b, 0xd8, 0x21, 0x61, 0x24, 0x52, 0x19, 0xb4, 0xfb, 0x79, 0x26, 0xb3, 0x47, 0xfd, 0xfb, 0x6c,
	0x1c, 0xe3, 0x69, 0xe0, 0x85, 0x2a, 0xff, 0x16, 0xc1, 0xe0, 0x42, 0xc8, 0x07, 0x7b, 0x06, 0x6f,
	0xe3, 0x8d, 0xb7, 0xbf, 0xdd, 0x31, 0xf8, 0xa7, 0x3b, 0x5c, 0x26, 0x6a, 0xb4, 0x56, 0x48, 0xf4,
	0x36, 0xa7, 0xbc, 0x2f, 0xdd, 0xff, 0xe9, 0x96, 0xc1, 0x81, 0xa1, 0x77, 0x60, 0xd2, 0x61, 0xce,
	0x04, 0x17, 0x11, 0x12, 0x52, 0xa2, 0xf1, 0xc1, 0xfc, 0x63, 0x44, 0x02, 0x27, 0xbf, 0x82, 0xe4,
	0xd2, 0xa2, 0x61, 0x8f, 0x61, 0x28, 0x75, 0xb9, 0x49, 0x66, 0x22, 0xf5, 0x59, 0xb5, 0xd1, 0x18,
	0x3f, 0xe4, 0xdf, 0xe0, 0xbe, 0x7f, 0x07, 0x30, 0x91, 0xba, 0x6c, 0xad, 0xfb, 0x34, 0x35, 0x76,
	0xdf, 0x10, 0xa4, 0xbe, 0xec, 0x90, 0x13, 0xf8, 0xb4, 0x1d, 0xc6, 0x58, 0xcf, 0xe7, 0x23, 0x6f,
	0xeb, 0xcb, 0x1f, 0x01, 0x00, 0x00, 0xff, 0xff, 0x1d, 0xed, 0xe0, 0x22, 0x53, 0x04, 0x00, 0x00,
}
