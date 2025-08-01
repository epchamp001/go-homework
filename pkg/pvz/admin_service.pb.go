// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.6
// 	protoc        v6.30.1
// source: pvz/admin_service.proto

package pvzpb

import (
	reflect "reflect"
	sync "sync"
	unsafe "unsafe"

	_ "github.com/envoyproxy/protoc-gen-validate/validate"
	_ "github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2/options"
	_ "google.golang.org/genproto/googleapis/api/annotations"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type ResizeRequest struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// Новое количество воркеров (1 … 100)
	Size          uint32 `protobuf:"varint,1,opt,name=size,proto3" json:"size,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ResizeRequest) Reset() {
	*x = ResizeRequest{}
	mi := &file_pvz_admin_service_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ResizeRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ResizeRequest) ProtoMessage() {}

func (x *ResizeRequest) ProtoReflect() protoreflect.Message {
	mi := &file_pvz_admin_service_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ResizeRequest.ProtoReflect.Descriptor instead.
func (*ResizeRequest) Descriptor() ([]byte, []int) {
	return file_pvz_admin_service_proto_rawDescGZIP(), []int{0}
}

func (x *ResizeRequest) GetSize() uint32 {
	if x != nil {
		return x.Size
	}
	return 0
}

var File_pvz_admin_service_proto protoreflect.FileDescriptor

const file_pvz_admin_service_proto_rawDesc = "" +
	"\n" +
	"\x17pvz/admin_service.proto\x12\x05admin\x1a\x1bgoogle/protobuf/empty.proto\x1a\x1cgoogle/api/annotations.proto\x1a\x17validate/validate.proto\x1a.protoc-gen-openapiv2/options/annotations.proto\"B\n" +
	"\rResizeRequest\x12\x1d\n" +
	"\x04size\x18\x01 \x01(\rB\t\xfaB\x06*\x04\x18d \x00R\x04size:\x12\x92A\x0f\n" +
	"\rJ\v{\"size\":16}2\xad\x01\n" +
	"\fAdminService\x12\x9c\x01\n" +
	"\n" +
	"ResizePool\x12\x14.admin.ResizeRequest\x1a\x16.google.protobuf.Empty\"`\x92A>\n" +
	"\x05admin\x12\x12Resize worker-pool*\x10Admin_ResizePoolb\x0f\n" +
	"\r\n" +
	"\tbasicAuth\x12\x00\x82\xd3\xe4\x93\x02\x19:\x01*\"\x14/v1/admin/resizePoolBG\x92A-\x12\x14\n" +
	"\rAdmin Service2\x031.0*\x02\x01\x02Z\x11\n" +
	"\x0f\n" +
	"\tbasicAuth\x12\x02\b\x01Z\x15pvz-cli/pkg/pvz;pvzpbb\x06proto3"

var (
	file_pvz_admin_service_proto_rawDescOnce sync.Once
	file_pvz_admin_service_proto_rawDescData []byte
)

func file_pvz_admin_service_proto_rawDescGZIP() []byte {
	file_pvz_admin_service_proto_rawDescOnce.Do(func() {
		file_pvz_admin_service_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_pvz_admin_service_proto_rawDesc), len(file_pvz_admin_service_proto_rawDesc)))
	})
	return file_pvz_admin_service_proto_rawDescData
}

var file_pvz_admin_service_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_pvz_admin_service_proto_goTypes = []any{
	(*ResizeRequest)(nil), // 0: admin.ResizeRequest
	(*emptypb.Empty)(nil), // 1: google.protobuf.Empty
}
var file_pvz_admin_service_proto_depIdxs = []int32{
	0, // 0: admin.AdminService.ResizePool:input_type -> admin.ResizeRequest
	1, // 1: admin.AdminService.ResizePool:output_type -> google.protobuf.Empty
	1, // [1:2] is the sub-list for method output_type
	0, // [0:1] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_pvz_admin_service_proto_init() }
func file_pvz_admin_service_proto_init() {
	if File_pvz_admin_service_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_pvz_admin_service_proto_rawDesc), len(file_pvz_admin_service_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_pvz_admin_service_proto_goTypes,
		DependencyIndexes: file_pvz_admin_service_proto_depIdxs,
		MessageInfos:      file_pvz_admin_service_proto_msgTypes,
	}.Build()
	File_pvz_admin_service_proto = out.File
	file_pvz_admin_service_proto_goTypes = nil
	file_pvz_admin_service_proto_depIdxs = nil
}
