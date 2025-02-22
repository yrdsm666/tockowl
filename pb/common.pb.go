// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.3
// 	protoc        v5.29.3
// source: pb/common.proto

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

type PacketType int32

const (
	PacketType_P2PPACKET    PacketType = 0
	PacketType_CLIENTPACKET PacketType = 1
)

// Enum value maps for PacketType.
var (
	PacketType_name = map[int32]string{
		0: "P2PPACKET",
		1: "CLIENTPACKET",
	}
	PacketType_value = map[string]int32{
		"P2PPACKET":    0,
		"CLIENTPACKET": 1,
	}
)

func (x PacketType) Enum() *PacketType {
	p := new(PacketType)
	*p = x
	return p
}

func (x PacketType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (PacketType) Descriptor() protoreflect.EnumDescriptor {
	return file_pb_common_proto_enumTypes[0].Descriptor()
}

func (PacketType) Type() protoreflect.EnumType {
	return &file_pb_common_proto_enumTypes[0]
}

func (x PacketType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use PacketType.Descriptor instead.
func (PacketType) EnumDescriptor() ([]byte, []int) {
	return file_pb_common_proto_rawDescGZIP(), []int{0}
}

type TransactionType int32

const (
	TransactionType_NORMAL   TransactionType = 0
	TransactionType_UPGRADE  TransactionType = 1
	TransactionType_TIMEVOTE TransactionType = 2
	TransactionType_LOCK     TransactionType = 3
)

// Enum value maps for TransactionType.
var (
	TransactionType_name = map[int32]string{
		0: "NORMAL",
		1: "UPGRADE",
		2: "TIMEVOTE",
		3: "LOCK",
	}
	TransactionType_value = map[string]int32{
		"NORMAL":   0,
		"UPGRADE":  1,
		"TIMEVOTE": 2,
		"LOCK":     3,
	}
)

func (x TransactionType) Enum() *TransactionType {
	p := new(TransactionType)
	*p = x
	return p
}

func (x TransactionType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (TransactionType) Descriptor() protoreflect.EnumDescriptor {
	return file_pb_common_proto_enumTypes[1].Descriptor()
}

func (TransactionType) Type() protoreflect.EnumType {
	return &file_pb_common_proto_enumTypes[1]
}

func (x TransactionType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use TransactionType.Descriptor instead.
func (TransactionType) EnumDescriptor() ([]byte, []int) {
	return file_pb_common_proto_rawDescGZIP(), []int{1}
}

type Empty struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Empty) Reset() {
	*x = Empty{}
	mi := &file_pb_common_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Empty) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Empty) ProtoMessage() {}

func (x *Empty) ProtoReflect() protoreflect.Message {
	mi := &file_pb_common_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Empty.ProtoReflect.Descriptor instead.
func (*Empty) Descriptor() ([]byte, []int) {
	return file_pb_common_proto_rawDescGZIP(), []int{0}
}

type Request struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Tx            []byte                 `protobuf:"bytes,1,opt,name=tx,proto3" json:"tx,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Request) Reset() {
	*x = Request{}
	mi := &file_pb_common_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Request) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Request) ProtoMessage() {}

func (x *Request) ProtoReflect() protoreflect.Message {
	mi := &file_pb_common_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Request.ProtoReflect.Descriptor instead.
func (*Request) Descriptor() ([]byte, []int) {
	return file_pb_common_proto_rawDescGZIP(), []int{1}
}

func (x *Request) GetTx() []byte {
	if x != nil {
		return x.Tx
	}
	return nil
}

type Reply struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Tx            []byte                 `protobuf:"bytes,1,opt,name=tx,proto3" json:"tx,omitempty"`
	Receipt       []byte                 `protobuf:"bytes,2,opt,name=receipt,proto3" json:"receipt,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Reply) Reset() {
	*x = Reply{}
	mi := &file_pb_common_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Reply) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Reply) ProtoMessage() {}

func (x *Reply) ProtoReflect() protoreflect.Message {
	mi := &file_pb_common_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Reply.ProtoReflect.Descriptor instead.
func (*Reply) Descriptor() ([]byte, []int) {
	return file_pb_common_proto_rawDescGZIP(), []int{2}
}

func (x *Reply) GetTx() []byte {
	if x != nil {
		return x.Tx
	}
	return nil
}

func (x *Reply) GetReceipt() []byte {
	if x != nil {
		return x.Receipt
	}
	return nil
}

type Header struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Height        uint64                 `protobuf:"varint,1,opt,name=Height,proto3" json:"Height,omitempty"`
	ParentHash    []byte                 `protobuf:"bytes,2,opt,name=ParentHash,proto3" json:"ParentHash,omitempty"`
	UncleHash     [][]byte               `protobuf:"bytes,3,rep,name=UncleHash,proto3" json:"UncleHash,omitempty"`
	Mixdigest     []byte                 `protobuf:"bytes,4,opt,name=Mixdigest,proto3" json:"Mixdigest,omitempty"`
	Difficulty    []byte                 `protobuf:"bytes,5,opt,name=Difficulty,proto3" json:"Difficulty,omitempty"`
	Nonce         int64                  `protobuf:"varint,6,opt,name=Nonce,proto3" json:"Nonce,omitempty"`
	Timestamp     []byte                 `protobuf:"bytes,7,opt,name=Timestamp,proto3" json:"Timestamp,omitempty"`
	PoTProof      [][]byte               `protobuf:"bytes,8,rep,name=PoTProof,proto3" json:"PoTProof,omitempty"`
	Address       int64                  `protobuf:"varint,9,opt,name=Address,proto3" json:"Address,omitempty"`
	Hashes        []byte                 `protobuf:"bytes,10,opt,name=Hashes,proto3" json:"Hashes,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Header) Reset() {
	*x = Header{}
	mi := &file_pb_common_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Header) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Header) ProtoMessage() {}

func (x *Header) ProtoReflect() protoreflect.Message {
	mi := &file_pb_common_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Header.ProtoReflect.Descriptor instead.
func (*Header) Descriptor() ([]byte, []int) {
	return file_pb_common_proto_rawDescGZIP(), []int{3}
}

func (x *Header) GetHeight() uint64 {
	if x != nil {
		return x.Height
	}
	return 0
}

func (x *Header) GetParentHash() []byte {
	if x != nil {
		return x.ParentHash
	}
	return nil
}

func (x *Header) GetUncleHash() [][]byte {
	if x != nil {
		return x.UncleHash
	}
	return nil
}

func (x *Header) GetMixdigest() []byte {
	if x != nil {
		return x.Mixdigest
	}
	return nil
}

func (x *Header) GetDifficulty() []byte {
	if x != nil {
		return x.Difficulty
	}
	return nil
}

func (x *Header) GetNonce() int64 {
	if x != nil {
		return x.Nonce
	}
	return 0
}

func (x *Header) GetTimestamp() []byte {
	if x != nil {
		return x.Timestamp
	}
	return nil
}

func (x *Header) GetPoTProof() [][]byte {
	if x != nil {
		return x.PoTProof
	}
	return nil
}

func (x *Header) GetAddress() int64 {
	if x != nil {
		return x.Address
	}
	return 0
}

func (x *Header) GetHashes() []byte {
	if x != nil {
		return x.Hashes
	}
	return nil
}

type HeaderRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Height        uint64                 `protobuf:"varint,1,opt,name=Height,proto3" json:"Height,omitempty"`
	Hashes        []byte                 `protobuf:"bytes,2,opt,name=Hashes,proto3" json:"Hashes,omitempty"`
	Address       int64                  `protobuf:"varint,3,opt,name=Address,proto3" json:"Address,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *HeaderRequest) Reset() {
	*x = HeaderRequest{}
	mi := &file_pb_common_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *HeaderRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*HeaderRequest) ProtoMessage() {}

func (x *HeaderRequest) ProtoReflect() protoreflect.Message {
	mi := &file_pb_common_proto_msgTypes[4]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use HeaderRequest.ProtoReflect.Descriptor instead.
func (*HeaderRequest) Descriptor() ([]byte, []int) {
	return file_pb_common_proto_rawDescGZIP(), []int{4}
}

func (x *HeaderRequest) GetHeight() uint64 {
	if x != nil {
		return x.Height
	}
	return 0
}

func (x *HeaderRequest) GetHashes() []byte {
	if x != nil {
		return x.Hashes
	}
	return nil
}

func (x *HeaderRequest) GetAddress() int64 {
	if x != nil {
		return x.Address
	}
	return 0
}

type Msg struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// Types that are valid to be assigned to Payload:
	//
	//	*Msg_Request
	//	*Msg_Reply
	Payload       isMsg_Payload `protobuf_oneof:"Payload"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Msg) Reset() {
	*x = Msg{}
	mi := &file_pb_common_proto_msgTypes[5]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Msg) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Msg) ProtoMessage() {}

func (x *Msg) ProtoReflect() protoreflect.Message {
	mi := &file_pb_common_proto_msgTypes[5]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Msg.ProtoReflect.Descriptor instead.
func (*Msg) Descriptor() ([]byte, []int) {
	return file_pb_common_proto_rawDescGZIP(), []int{5}
}

func (x *Msg) GetPayload() isMsg_Payload {
	if x != nil {
		return x.Payload
	}
	return nil
}

func (x *Msg) GetRequest() *Request {
	if x != nil {
		if x, ok := x.Payload.(*Msg_Request); ok {
			return x.Request
		}
	}
	return nil
}

func (x *Msg) GetReply() *Reply {
	if x != nil {
		if x, ok := x.Payload.(*Msg_Reply); ok {
			return x.Reply
		}
	}
	return nil
}

type isMsg_Payload interface {
	isMsg_Payload()
}

type Msg_Request struct {
	Request *Request `protobuf:"bytes,1,opt,name=request,proto3,oneof"`
}

type Msg_Reply struct {
	Reply *Reply `protobuf:"bytes,2,opt,name=reply,proto3,oneof"`
}

func (*Msg_Request) isMsg_Payload() {}

func (*Msg_Reply) isMsg_Payload() {}

type Packet struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Msg           []byte                 `protobuf:"bytes,1,opt,name=msg,proto3" json:"msg,omitempty"`
	ConsensusID   int64                  `protobuf:"varint,2,opt,name=consensusID,proto3" json:"consensusID,omitempty"`
	Epoch         int64                  `protobuf:"varint,3,opt,name=epoch,proto3" json:"epoch,omitempty"`
	Type          PacketType             `protobuf:"varint,4,opt,name=type,proto3,enum=pb.PacketType" json:"type,omitempty"`
	Partition     string                 `protobuf:"bytes,5,opt,name=partition,proto3" json:"partition,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Packet) Reset() {
	*x = Packet{}
	mi := &file_pb_common_proto_msgTypes[6]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Packet) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Packet) ProtoMessage() {}

func (x *Packet) ProtoReflect() protoreflect.Message {
	mi := &file_pb_common_proto_msgTypes[6]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Packet.ProtoReflect.Descriptor instead.
func (*Packet) Descriptor() ([]byte, []int) {
	return file_pb_common_proto_rawDescGZIP(), []int{6}
}

func (x *Packet) GetMsg() []byte {
	if x != nil {
		return x.Msg
	}
	return nil
}

func (x *Packet) GetConsensusID() int64 {
	if x != nil {
		return x.ConsensusID
	}
	return 0
}

func (x *Packet) GetEpoch() int64 {
	if x != nil {
		return x.Epoch
	}
	return 0
}

func (x *Packet) GetType() PacketType {
	if x != nil {
		return x.Type
	}
	return PacketType_P2PPACKET
}

func (x *Packet) GetPartition() string {
	if x != nil {
		return x.Partition
	}
	return ""
}

type Transaction struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Type          TransactionType        `protobuf:"varint,1,opt,name=type,proto3,enum=pb.TransactionType" json:"type,omitempty"`
	Payload       []byte                 `protobuf:"bytes,2,opt,name=payload,proto3" json:"payload,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Transaction) Reset() {
	*x = Transaction{}
	mi := &file_pb_common_proto_msgTypes[7]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Transaction) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Transaction) ProtoMessage() {}

func (x *Transaction) ProtoReflect() protoreflect.Message {
	mi := &file_pb_common_proto_msgTypes[7]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Transaction.ProtoReflect.Descriptor instead.
func (*Transaction) Descriptor() ([]byte, []int) {
	return file_pb_common_proto_rawDescGZIP(), []int{7}
}

func (x *Transaction) GetType() TransactionType {
	if x != nil {
		return x.Type
	}
	return TransactionType_NORMAL
}

func (x *Transaction) GetPayload() []byte {
	if x != nil {
		return x.Payload
	}
	return nil
}

var File_pb_common_proto protoreflect.FileDescriptor

var file_pb_common_proto_rawDesc = []byte{
	0x0a, 0x0f, 0x70, 0x62, 0x2f, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x12, 0x02, 0x70, 0x62, 0x22, 0x07, 0x0a, 0x05, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x22, 0x19,
	0x0a, 0x07, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x0e, 0x0a, 0x02, 0x74, 0x78, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x02, 0x74, 0x78, 0x22, 0x31, 0x0a, 0x05, 0x52, 0x65, 0x70,
	0x6c, 0x79, 0x12, 0x0e, 0x0a, 0x02, 0x74, 0x78, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x02,
	0x74, 0x78, 0x12, 0x18, 0x0a, 0x07, 0x72, 0x65, 0x63, 0x65, 0x69, 0x70, 0x74, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x0c, 0x52, 0x07, 0x72, 0x65, 0x63, 0x65, 0x69, 0x70, 0x74, 0x22, 0x9e, 0x02, 0x0a,
	0x06, 0x48, 0x65, 0x61, 0x64, 0x65, 0x72, 0x12, 0x16, 0x0a, 0x06, 0x48, 0x65, 0x69, 0x67, 0x68,
	0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x04, 0x52, 0x06, 0x48, 0x65, 0x69, 0x67, 0x68, 0x74, 0x12,
	0x1e, 0x0a, 0x0a, 0x50, 0x61, 0x72, 0x65, 0x6e, 0x74, 0x48, 0x61, 0x73, 0x68, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x0c, 0x52, 0x0a, 0x50, 0x61, 0x72, 0x65, 0x6e, 0x74, 0x48, 0x61, 0x73, 0x68, 0x12,
	0x1c, 0x0a, 0x09, 0x55, 0x6e, 0x63, 0x6c, 0x65, 0x48, 0x61, 0x73, 0x68, 0x18, 0x03, 0x20, 0x03,
	0x28, 0x0c, 0x52, 0x09, 0x55, 0x6e, 0x63, 0x6c, 0x65, 0x48, 0x61, 0x73, 0x68, 0x12, 0x1c, 0x0a,
	0x09, 0x4d, 0x69, 0x78, 0x64, 0x69, 0x67, 0x65, 0x73, 0x74, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0c,
	0x52, 0x09, 0x4d, 0x69, 0x78, 0x64, 0x69, 0x67, 0x65, 0x73, 0x74, 0x12, 0x1e, 0x0a, 0x0a, 0x44,
	0x69, 0x66, 0x66, 0x69, 0x63, 0x75, 0x6c, 0x74, 0x79, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0c, 0x52,
	0x0a, 0x44, 0x69, 0x66, 0x66, 0x69, 0x63, 0x75, 0x6c, 0x74, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x4e,
	0x6f, 0x6e, 0x63, 0x65, 0x18, 0x06, 0x20, 0x01, 0x28, 0x03, 0x52, 0x05, 0x4e, 0x6f, 0x6e, 0x63,
	0x65, 0x12, 0x1c, 0x0a, 0x09, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x18, 0x07,
	0x20, 0x01, 0x28, 0x0c, 0x52, 0x09, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x12,
	0x1a, 0x0a, 0x08, 0x50, 0x6f, 0x54, 0x50, 0x72, 0x6f, 0x6f, 0x66, 0x18, 0x08, 0x20, 0x03, 0x28,
	0x0c, 0x52, 0x08, 0x50, 0x6f, 0x54, 0x50, 0x72, 0x6f, 0x6f, 0x66, 0x12, 0x18, 0x0a, 0x07, 0x41,
	0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x18, 0x09, 0x20, 0x01, 0x28, 0x03, 0x52, 0x07, 0x41, 0x64,
	0x64, 0x72, 0x65, 0x73, 0x73, 0x12, 0x16, 0x0a, 0x06, 0x48, 0x61, 0x73, 0x68, 0x65, 0x73, 0x18,
	0x0a, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x06, 0x48, 0x61, 0x73, 0x68, 0x65, 0x73, 0x22, 0x59, 0x0a,
	0x0d, 0x48, 0x65, 0x61, 0x64, 0x65, 0x72, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x16,
	0x0a, 0x06, 0x48, 0x65, 0x69, 0x67, 0x68, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x04, 0x52, 0x06,
	0x48, 0x65, 0x69, 0x67, 0x68, 0x74, 0x12, 0x16, 0x0a, 0x06, 0x48, 0x61, 0x73, 0x68, 0x65, 0x73,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x06, 0x48, 0x61, 0x73, 0x68, 0x65, 0x73, 0x12, 0x18,
	0x0a, 0x07, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x18, 0x03, 0x20, 0x01, 0x28, 0x03, 0x52,
	0x07, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x22, 0x5c, 0x0a, 0x03, 0x4d, 0x73, 0x67, 0x12,
	0x27, 0x0a, 0x07, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x0b, 0x2e, 0x70, 0x62, 0x2e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x48, 0x00, 0x52,
	0x07, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x21, 0x0a, 0x05, 0x72, 0x65, 0x70, 0x6c,
	0x79, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x09, 0x2e, 0x70, 0x62, 0x2e, 0x52, 0x65, 0x70,
	0x6c, 0x79, 0x48, 0x00, 0x52, 0x05, 0x72, 0x65, 0x70, 0x6c, 0x79, 0x42, 0x09, 0x0a, 0x07, 0x50,
	0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x22, 0x94, 0x01, 0x0a, 0x06, 0x50, 0x61, 0x63, 0x6b, 0x65,
	0x74, 0x12, 0x10, 0x0a, 0x03, 0x6d, 0x73, 0x67, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x03,
	0x6d, 0x73, 0x67, 0x12, 0x20, 0x0a, 0x0b, 0x63, 0x6f, 0x6e, 0x73, 0x65, 0x6e, 0x73, 0x75, 0x73,
	0x49, 0x44, 0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x0b, 0x63, 0x6f, 0x6e, 0x73, 0x65, 0x6e,
	0x73, 0x75, 0x73, 0x49, 0x44, 0x12, 0x14, 0x0a, 0x05, 0x65, 0x70, 0x6f, 0x63, 0x68, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x03, 0x52, 0x05, 0x65, 0x70, 0x6f, 0x63, 0x68, 0x12, 0x22, 0x0a, 0x04, 0x74,
	0x79, 0x70, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x0e, 0x2e, 0x70, 0x62, 0x2e, 0x50,
	0x61, 0x63, 0x6b, 0x65, 0x74, 0x54, 0x79, 0x70, 0x65, 0x52, 0x04, 0x74, 0x79, 0x70, 0x65, 0x12,
	0x1c, 0x0a, 0x09, 0x70, 0x61, 0x72, 0x74, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x05, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x09, 0x70, 0x61, 0x72, 0x74, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x22, 0x50, 0x0a,
	0x0b, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x27, 0x0a, 0x04,
	0x74, 0x79, 0x70, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x13, 0x2e, 0x70, 0x62, 0x2e,
	0x54, 0x72, 0x61, 0x6e, 0x73, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x54, 0x79, 0x70, 0x65, 0x52,
	0x04, 0x74, 0x79, 0x70, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x70, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x07, 0x70, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x2a,
	0x2d, 0x0a, 0x0a, 0x50, 0x61, 0x63, 0x6b, 0x65, 0x74, 0x54, 0x79, 0x70, 0x65, 0x12, 0x0d, 0x0a,
	0x09, 0x50, 0x32, 0x50, 0x50, 0x41, 0x43, 0x4b, 0x45, 0x54, 0x10, 0x00, 0x12, 0x10, 0x0a, 0x0c,
	0x43, 0x4c, 0x49, 0x45, 0x4e, 0x54, 0x50, 0x41, 0x43, 0x4b, 0x45, 0x54, 0x10, 0x01, 0x2a, 0x42,
	0x0a, 0x0f, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x54, 0x79, 0x70,
	0x65, 0x12, 0x0a, 0x0a, 0x06, 0x4e, 0x4f, 0x52, 0x4d, 0x41, 0x4c, 0x10, 0x00, 0x12, 0x0b, 0x0a,
	0x07, 0x55, 0x50, 0x47, 0x52, 0x41, 0x44, 0x45, 0x10, 0x01, 0x12, 0x0c, 0x0a, 0x08, 0x54, 0x49,
	0x4d, 0x45, 0x56, 0x4f, 0x54, 0x45, 0x10, 0x02, 0x12, 0x08, 0x0a, 0x04, 0x4c, 0x4f, 0x43, 0x4b,
	0x10, 0x03, 0x32, 0x52, 0x0a, 0x03, 0x50, 0x32, 0x50, 0x12, 0x1f, 0x0a, 0x04, 0x53, 0x65, 0x6e,
	0x64, 0x12, 0x0a, 0x2e, 0x70, 0x62, 0x2e, 0x50, 0x61, 0x63, 0x6b, 0x65, 0x74, 0x1a, 0x09, 0x2e,
	0x70, 0x62, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x22, 0x00, 0x12, 0x2a, 0x0a, 0x07, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x11, 0x2e, 0x70, 0x62, 0x2e, 0x48, 0x65, 0x61, 0x64, 0x65,
	0x72, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x0a, 0x2e, 0x70, 0x62, 0x2e, 0x48, 0x65,
	0x61, 0x64, 0x65, 0x72, 0x22, 0x00, 0x42, 0x06, 0x5a, 0x04, 0x2e, 0x2f, 0x70, 0x62, 0x62, 0x06,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_pb_common_proto_rawDescOnce sync.Once
	file_pb_common_proto_rawDescData = file_pb_common_proto_rawDesc
)

func file_pb_common_proto_rawDescGZIP() []byte {
	file_pb_common_proto_rawDescOnce.Do(func() {
		file_pb_common_proto_rawDescData = protoimpl.X.CompressGZIP(file_pb_common_proto_rawDescData)
	})
	return file_pb_common_proto_rawDescData
}

var file_pb_common_proto_enumTypes = make([]protoimpl.EnumInfo, 2)
var file_pb_common_proto_msgTypes = make([]protoimpl.MessageInfo, 8)
var file_pb_common_proto_goTypes = []any{
	(PacketType)(0),       // 0: pb.PacketType
	(TransactionType)(0),  // 1: pb.TransactionType
	(*Empty)(nil),         // 2: pb.Empty
	(*Request)(nil),       // 3: pb.Request
	(*Reply)(nil),         // 4: pb.Reply
	(*Header)(nil),        // 5: pb.Header
	(*HeaderRequest)(nil), // 6: pb.HeaderRequest
	(*Msg)(nil),           // 7: pb.Msg
	(*Packet)(nil),        // 8: pb.Packet
	(*Transaction)(nil),   // 9: pb.Transaction
}
var file_pb_common_proto_depIdxs = []int32{
	3, // 0: pb.Msg.request:type_name -> pb.Request
	4, // 1: pb.Msg.reply:type_name -> pb.Reply
	0, // 2: pb.Packet.type:type_name -> pb.PacketType
	1, // 3: pb.Transaction.type:type_name -> pb.TransactionType
	8, // 4: pb.P2P.Send:input_type -> pb.Packet
	6, // 5: pb.P2P.Request:input_type -> pb.HeaderRequest
	2, // 6: pb.P2P.Send:output_type -> pb.Empty
	5, // 7: pb.P2P.Request:output_type -> pb.Header
	6, // [6:8] is the sub-list for method output_type
	4, // [4:6] is the sub-list for method input_type
	4, // [4:4] is the sub-list for extension type_name
	4, // [4:4] is the sub-list for extension extendee
	0, // [0:4] is the sub-list for field type_name
}

func init() { file_pb_common_proto_init() }
func file_pb_common_proto_init() {
	if File_pb_common_proto != nil {
		return
	}
	file_pb_common_proto_msgTypes[5].OneofWrappers = []any{
		(*Msg_Request)(nil),
		(*Msg_Reply)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_pb_common_proto_rawDesc,
			NumEnums:      2,
			NumMessages:   8,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_pb_common_proto_goTypes,
		DependencyIndexes: file_pb_common_proto_depIdxs,
		EnumInfos:         file_pb_common_proto_enumTypes,
		MessageInfos:      file_pb_common_proto_msgTypes,
	}.Build()
	File_pb_common_proto = out.File
	file_pb_common_proto_rawDesc = nil
	file_pb_common_proto_goTypes = nil
	file_pb_common_proto_depIdxs = nil
}
