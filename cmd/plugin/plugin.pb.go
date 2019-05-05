// Code generated by protoc-gen-go. DO NOT EDIT.
// source: plugin.proto

package plugin

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	ast "github.com/gqlc/graphql/ast"
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

// An encoded PluginRequest is written to the plugin's stdin.
type PluginRequest struct {
	// The .gqlc/.graphql files to generate.
	FileToGenerate []string `protobuf:"bytes,1,rep,name=file_to_generate,json=fileToGenerate,proto3" json:"file_to_generate,omitempty"`
	// The generator parameter passed on the command-line.
	Parameter string `protobuf:"bytes,2,opt,name=parameter,proto3" json:"parameter,omitempty"`
	// Documents are all the parsed documents to be generated.
	Documents            []*ast.Document `protobuf:"bytes,3,rep,name=documents,proto3" json:"documents,omitempty"`
	XXX_NoUnkeyedLiteral struct{}        `json:"-"`
	XXX_unrecognized     []byte          `json:"-"`
	XXX_sizecache        int32           `json:"-"`
}

func (m *PluginRequest) Reset()         { *m = PluginRequest{} }
func (m *PluginRequest) String() string { return proto.CompactTextString(m) }
func (*PluginRequest) ProtoMessage()    {}
func (*PluginRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_22a625af4bc1cc87, []int{0}
}

func (m *PluginRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PluginRequest.Unmarshal(m, b)
}
func (m *PluginRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PluginRequest.Marshal(b, m, deterministic)
}
func (m *PluginRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PluginRequest.Merge(m, src)
}
func (m *PluginRequest) XXX_Size() int {
	return xxx_messageInfo_PluginRequest.Size(m)
}
func (m *PluginRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_PluginRequest.DiscardUnknown(m)
}

var xxx_messageInfo_PluginRequest proto.InternalMessageInfo

func (m *PluginRequest) GetFileToGenerate() []string {
	if m != nil {
		return m.FileToGenerate
	}
	return nil
}

func (m *PluginRequest) GetParameter() string {
	if m != nil {
		return m.Parameter
	}
	return ""
}

func (m *PluginRequest) GetDocuments() []*ast.Document {
	if m != nil {
		return m.Documents
	}
	return nil
}

// The plugin writes an encoded PluginResponse to stdout.
type PluginResponse struct {
	// Error message. If non-empty code generation failed. The plugin
	// process should exit with status code zero even if it reports
	// an error in this way.
	//
	Error                string                 `protobuf:"bytes,1,opt,name=error,proto3" json:"error,omitempty"`
	File                 []*PluginResponse_File `protobuf:"bytes,2,rep,name=file,proto3" json:"file,omitempty"`
	XXX_NoUnkeyedLiteral struct{}               `json:"-"`
	XXX_unrecognized     []byte                 `json:"-"`
	XXX_sizecache        int32                  `json:"-"`
}

func (m *PluginResponse) Reset()         { *m = PluginResponse{} }
func (m *PluginResponse) String() string { return proto.CompactTextString(m) }
func (*PluginResponse) ProtoMessage()    {}
func (*PluginResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_22a625af4bc1cc87, []int{1}
}

func (m *PluginResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PluginResponse.Unmarshal(m, b)
}
func (m *PluginResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PluginResponse.Marshal(b, m, deterministic)
}
func (m *PluginResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PluginResponse.Merge(m, src)
}
func (m *PluginResponse) XXX_Size() int {
	return xxx_messageInfo_PluginResponse.Size(m)
}
func (m *PluginResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_PluginResponse.DiscardUnknown(m)
}

var xxx_messageInfo_PluginResponse proto.InternalMessageInfo

func (m *PluginResponse) GetError() string {
	if m != nil {
		return m.Error
	}
	return ""
}

func (m *PluginResponse) GetFile() []*PluginResponse_File {
	if m != nil {
		return m.File
	}
	return nil
}

// Represents a single generated file.
type PluginResponse_File struct {
	// The file name, relative to the output directory. The name must not
	// contain "." or ".." components and must be relative, not be absolute (so,
	// the file cannot lie outside the output directory). "/" must be used as
	// the path separator, not "\".
	//
	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	// The file contents.
	Content              string   `protobuf:"bytes,15,opt,name=content,proto3" json:"content,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PluginResponse_File) Reset()         { *m = PluginResponse_File{} }
func (m *PluginResponse_File) String() string { return proto.CompactTextString(m) }
func (*PluginResponse_File) ProtoMessage()    {}
func (*PluginResponse_File) Descriptor() ([]byte, []int) {
	return fileDescriptor_22a625af4bc1cc87, []int{1, 0}
}

func (m *PluginResponse_File) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PluginResponse_File.Unmarshal(m, b)
}
func (m *PluginResponse_File) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PluginResponse_File.Marshal(b, m, deterministic)
}
func (m *PluginResponse_File) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PluginResponse_File.Merge(m, src)
}
func (m *PluginResponse_File) XXX_Size() int {
	return xxx_messageInfo_PluginResponse_File.Size(m)
}
func (m *PluginResponse_File) XXX_DiscardUnknown() {
	xxx_messageInfo_PluginResponse_File.DiscardUnknown(m)
}

var xxx_messageInfo_PluginResponse_File proto.InternalMessageInfo

func (m *PluginResponse_File) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *PluginResponse_File) GetContent() string {
	if m != nil {
		return m.Content
	}
	return ""
}

func init() {
	proto.RegisterType((*PluginRequest)(nil), "PluginRequest")
	proto.RegisterType((*PluginResponse)(nil), "PluginResponse")
	proto.RegisterType((*PluginResponse_File)(nil), "PluginResponse.File")
}

func init() { proto.RegisterFile("plugin.proto", fileDescriptor_22a625af4bc1cc87) }

var fileDescriptor_22a625af4bc1cc87 = []byte{
	// 234 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x54, 0x90, 0x41, 0x4b, 0xc4, 0x30,
	0x10, 0x85, 0xc9, 0xb6, 0x2a, 0x19, 0x75, 0x95, 0xb0, 0x42, 0x58, 0x3c, 0x94, 0xbd, 0x98, 0x53,
	0x05, 0xf5, 0x27, 0x88, 0x5e, 0x25, 0x78, 0x5f, 0xe2, 0x3a, 0xd6, 0x42, 0x9b, 0x64, 0x27, 0xd3,
	0xbb, 0x27, 0x7f, 0xb7, 0x34, 0xdb, 0xa5, 0x78, 0x08, 0xe4, 0x7d, 0xf3, 0xe6, 0xf1, 0x18, 0xb8,
	0x88, 0xdd, 0xd0, 0xb4, 0xbe, 0x8e, 0x14, 0x38, 0xac, 0x6f, 0x1a, 0x72, 0xf1, 0x7b, 0xdf, 0xdd,
	0xbb, 0xc4, 0xe3, 0x3b, 0xe0, 0xcd, 0x8f, 0x80, 0xcb, 0xb7, 0xec, 0xb3, 0xb8, 0x1f, 0x30, 0xb1,
	0x32, 0x70, 0xfd, 0xd5, 0x76, 0xb8, 0xe5, 0xb0, 0x6d, 0xd0, 0x23, 0x39, 0x46, 0x2d, 0xaa, 0xc2,
	0x48, 0xbb, 0x1c, 0xf9, 0x7b, 0x78, 0x9d, 0xa8, 0xba, 0x05, 0x19, 0x1d, 0xb9, 0x1e, 0x19, 0x49,
	0x2f, 0x2a, 0x61, 0xa4, 0x9d, 0x81, 0xba, 0x03, 0xf9, 0x19, 0x76, 0x43, 0x8f, 0x9e, 0x93, 0x2e,
	0xaa, 0xc2, 0x9c, 0x3f, 0xc8, 0xfa, 0x79, 0x22, 0x76, 0x9e, 0x6d, 0x7e, 0x05, 0x2c, 0x8f, 0x15,
	0x52, 0x0c, 0x3e, 0xa1, 0x5a, 0xc1, 0x09, 0x12, 0x05, 0xd2, 0x22, 0xa7, 0x1e, 0x84, 0x32, 0x50,
	0x8e, 0x0d, 0xf4, 0x22, 0x87, 0xad, 0xea, 0xff, 0x4b, 0xf5, 0x4b, 0xdb, 0xa1, 0xcd, 0x8e, 0xf5,
	0x13, 0x94, 0xa3, 0x52, 0x0a, 0x4a, 0xef, 0x7a, 0x9c, 0x62, 0xf2, 0x5f, 0x69, 0x38, 0xdb, 0x05,
	0xcf, 0xe8, 0x59, 0x5f, 0x65, 0x7c, 0x94, 0x1f, 0xa7, 0xf9, 0x24, 0x8f, 0x7f, 0x01, 0x00, 0x00,
	0xff, 0xff, 0xd1, 0x8e, 0xc1, 0xd1, 0x39, 0x01, 0x00, 0x00,
}