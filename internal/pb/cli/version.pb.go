// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.6
// 	protoc        v5.28.3
// source: config/cli/version.proto

package cli

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	unsafe "unsafe"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type Version struct {
	state                        protoimpl.MessageState `protogen:"opaque.v1"`
	xxx_hidden_ProductionVersion string                 `protobuf:"bytes,1,opt,name=production_version,json=productionVersion,proto3"`
	unknownFields                protoimpl.UnknownFields
	sizeCache                    protoimpl.SizeCache
}

func (x *Version) Reset() {
	*x = Version{}
	mi := &file_config_cli_version_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Version) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Version) ProtoMessage() {}

func (x *Version) ProtoReflect() protoreflect.Message {
	mi := &file_config_cli_version_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (x *Version) GetProductionVersion() string {
	if x != nil {
		return x.xxx_hidden_ProductionVersion
	}
	return ""
}

func (x *Version) SetProductionVersion(v string) {
	x.xxx_hidden_ProductionVersion = v
}

type Version_builder struct {
	_ [0]func() // Prevents comparability and use of unkeyed literals for the builder.

	ProductionVersion string
}

func (b0 Version_builder) Build() *Version {
	m0 := &Version{}
	b, x := &b0, m0
	_, _ = b, x
	x.xxx_hidden_ProductionVersion = b.ProductionVersion
	return m0
}

var File_config_cli_version_proto protoreflect.FileDescriptor

const file_config_cli_version_proto_rawDesc = "" +
	"\n" +
	"\x18config/cli/version.proto\x12\x05proto\"8\n" +
	"\aVersion\x12-\n" +
	"\x12production_version\x18\x01 \x01(\tR\x11productionVersionB3Z1github.com/devplaninc/devplan-cli/internal/pb/clib\x06proto3"

var file_config_cli_version_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_config_cli_version_proto_goTypes = []any{
	(*Version)(nil), // 0: proto.Version
}
var file_config_cli_version_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_config_cli_version_proto_init() }
func file_config_cli_version_proto_init() {
	if File_config_cli_version_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_config_cli_version_proto_rawDesc), len(file_config_cli_version_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_config_cli_version_proto_goTypes,
		DependencyIndexes: file_config_cli_version_proto_depIdxs,
		MessageInfos:      file_config_cli_version_proto_msgTypes,
	}.Build()
	File_config_cli_version_proto = out.File
	file_config_cli_version_proto_goTypes = nil
	file_config_cli_version_proto_depIdxs = nil
}
