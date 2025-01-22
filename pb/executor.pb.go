// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.3
// 	protoc        v5.29.3
// source: pb/executor.proto

package pb

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type ExecBlock struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Txs           [][]byte               `protobuf:"bytes,1,rep,name=txs,proto3" json:"txs,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ExecBlock) Reset() {
	*x = ExecBlock{}
	mi := &file_pb_executor_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ExecBlock) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ExecBlock) ProtoMessage() {}

func (x *ExecBlock) ProtoReflect() protoreflect.Message {
	mi := &file_pb_executor_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ExecBlock.ProtoReflect.Descriptor instead.
func (*ExecBlock) Descriptor() ([]byte, []int) {
	return file_pb_executor_proto_rawDescGZIP(), []int{0}
}

func (x *ExecBlock) GetTxs() [][]byte {
	if x != nil {
		return x.Txs
	}
	return nil
}

type Result struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Success       bool                   `protobuf:"varint,1,opt,name=success,proto3" json:"success,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Result) Reset() {
	*x = Result{}
	mi := &file_pb_executor_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Result) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Result) ProtoMessage() {}

func (x *Result) ProtoReflect() protoreflect.Message {
	mi := &file_pb_executor_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Result.ProtoReflect.Descriptor instead.
func (*Result) Descriptor() ([]byte, []int) {
	return file_pb_executor_proto_rawDescGZIP(), []int{1}
}

func (x *Result) GetSuccess() bool {
	if x != nil {
		return x.Success
	}
	return false
}

var File_pb_executor_proto protoreflect.FileDescriptor

var file_pb_executor_proto_rawDesc = []byte{
	0x0a, 0x11, 0x70, 0x62, 0x2f, 0x65, 0x78, 0x65, 0x63, 0x75, 0x74, 0x6f, 0x72, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x12, 0x02, 0x70, 0x62, 0x1a, 0x0f, 0x70, 0x62, 0x2f, 0x63, 0x6f, 0x6d, 0x6d,
	0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x1d, 0x0a, 0x09, 0x45, 0x78, 0x65, 0x63,
	0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x12, 0x10, 0x0a, 0x03, 0x74, 0x78, 0x73, 0x18, 0x01, 0x20, 0x03,
	0x28, 0x0c, 0x52, 0x03, 0x74, 0x78, 0x73, 0x22, 0x22, 0x0a, 0x06, 0x52, 0x65, 0x73, 0x75, 0x6c,
	0x74, 0x12, 0x18, 0x0a, 0x07, 0x73, 0x75, 0x63, 0x63, 0x65, 0x73, 0x73, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x08, 0x52, 0x07, 0x73, 0x75, 0x63, 0x63, 0x65, 0x73, 0x73, 0x32, 0x60, 0x0a, 0x08, 0x45,
	0x78, 0x65, 0x63, 0x75, 0x74, 0x6f, 0x72, 0x12, 0x29, 0x0a, 0x0b, 0x43, 0x6f, 0x6d, 0x6d, 0x69,
	0x74, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x12, 0x0d, 0x2e, 0x70, 0x62, 0x2e, 0x45, 0x78, 0x65, 0x63,
	0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x1a, 0x09, 0x2e, 0x70, 0x62, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79,
	0x22, 0x00, 0x12, 0x29, 0x0a, 0x08, 0x56, 0x65, 0x72, 0x69, 0x66, 0x79, 0x54, 0x78, 0x12, 0x0f,
	0x2e, 0x70, 0x62, 0x2e, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x1a,
	0x0a, 0x2e, 0x70, 0x62, 0x2e, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x22, 0x00, 0x42, 0x06, 0x5a,
	0x04, 0x2e, 0x2f, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_pb_executor_proto_rawDescOnce sync.Once
	file_pb_executor_proto_rawDescData = file_pb_executor_proto_rawDesc
)

func file_pb_executor_proto_rawDescGZIP() []byte {
	file_pb_executor_proto_rawDescOnce.Do(func() {
		file_pb_executor_proto_rawDescData = protoimpl.X.CompressGZIP(file_pb_executor_proto_rawDescData)
	})
	return file_pb_executor_proto_rawDescData
}

var file_pb_executor_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_pb_executor_proto_goTypes = []any{
	(*ExecBlock)(nil),   // 0: pb.ExecBlock
	(*Result)(nil),      // 1: pb.Result
	(*Transaction)(nil), // 2: pb.Transaction
	(*Empty)(nil),       // 3: pb.Empty
}
var file_pb_executor_proto_depIdxs = []int32{
	0, // 0: pb.Executor.CommitBlock:input_type -> pb.ExecBlock
	2, // 1: pb.Executor.VerifyTx:input_type -> pb.Transaction
	3, // 2: pb.Executor.CommitBlock:output_type -> pb.Empty
	1, // 3: pb.Executor.VerifyTx:output_type -> pb.Result
	2, // [2:4] is the sub-list for method output_type
	0, // [0:2] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_pb_executor_proto_init() }
func file_pb_executor_proto_init() {
	if File_pb_executor_proto != nil {
		return
	}
	file_pb_common_proto_init()
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_pb_executor_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_pb_executor_proto_goTypes,
		DependencyIndexes: file_pb_executor_proto_depIdxs,
		MessageInfos:      file_pb_executor_proto_msgTypes,
	}.Build()
	File_pb_executor_proto = out.File
	file_pb_executor_proto_rawDesc = nil
	file_pb_executor_proto_goTypes = nil
	file_pb_executor_proto_depIdxs = nil
}
