// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        v4.24.4
// source: blockchainpb/synchronizerpb/synchronizerpb.proto

package synchronizerpb

import (
	blockchainpb "github.com/filecoin-project/mir/pkg/pb/blockchainpb"
	_ "github.com/filecoin-project/mir/pkg/pb/mir"
	_ "github.com/filecoin-project/mir/pkg/pb/net"
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

type Event struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Types that are assignable to Type:
	//
	//	*Event_SyncRequest
	Type isEvent_Type `protobuf_oneof:"type"`
}

func (x *Event) Reset() {
	*x = Event{}
	if protoimpl.UnsafeEnabled {
		mi := &file_blockchainpb_synchronizerpb_synchronizerpb_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Event) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Event) ProtoMessage() {}

func (x *Event) ProtoReflect() protoreflect.Message {
	mi := &file_blockchainpb_synchronizerpb_synchronizerpb_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Event.ProtoReflect.Descriptor instead.
func (*Event) Descriptor() ([]byte, []int) {
	return file_blockchainpb_synchronizerpb_synchronizerpb_proto_rawDescGZIP(), []int{0}
}

func (m *Event) GetType() isEvent_Type {
	if m != nil {
		return m.Type
	}
	return nil
}

func (x *Event) GetSyncRequest() *SyncRequest {
	if x, ok := x.GetType().(*Event_SyncRequest); ok {
		return x.SyncRequest
	}
	return nil
}

type isEvent_Type interface {
	isEvent_Type()
}

type Event_SyncRequest struct {
	SyncRequest *SyncRequest `protobuf:"bytes,1,opt,name=sync_request,json=syncRequest,proto3,oneof"`
}

func (*Event_SyncRequest) isEvent_Type() {}

type SyncRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	OrphanBlock *blockchainpb.Block `protobuf:"bytes,1,opt,name=orphan_block,json=orphanBlock,proto3" json:"orphan_block,omitempty"`
	LeaveIds    []uint64            `protobuf:"varint,2,rep,packed,name=leave_ids,json=leaveIds,proto3" json:"leave_ids,omitempty"`
}

func (x *SyncRequest) Reset() {
	*x = SyncRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_blockchainpb_synchronizerpb_synchronizerpb_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SyncRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SyncRequest) ProtoMessage() {}

func (x *SyncRequest) ProtoReflect() protoreflect.Message {
	mi := &file_blockchainpb_synchronizerpb_synchronizerpb_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SyncRequest.ProtoReflect.Descriptor instead.
func (*SyncRequest) Descriptor() ([]byte, []int) {
	return file_blockchainpb_synchronizerpb_synchronizerpb_proto_rawDescGZIP(), []int{1}
}

func (x *SyncRequest) GetOrphanBlock() *blockchainpb.Block {
	if x != nil {
		return x.OrphanBlock
	}
	return nil
}

func (x *SyncRequest) GetLeaveIds() []uint64 {
	if x != nil {
		return x.LeaveIds
	}
	return nil
}

type Message struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Types that are assignable to Type:
	//
	//	*Message_BlockRequest
	//	*Message_BlockResponse
	Type isMessage_Type `protobuf_oneof:"type"`
}

func (x *Message) Reset() {
	*x = Message{}
	if protoimpl.UnsafeEnabled {
		mi := &file_blockchainpb_synchronizerpb_synchronizerpb_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Message) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Message) ProtoMessage() {}

func (x *Message) ProtoReflect() protoreflect.Message {
	mi := &file_blockchainpb_synchronizerpb_synchronizerpb_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Message.ProtoReflect.Descriptor instead.
func (*Message) Descriptor() ([]byte, []int) {
	return file_blockchainpb_synchronizerpb_synchronizerpb_proto_rawDescGZIP(), []int{2}
}

func (m *Message) GetType() isMessage_Type {
	if m != nil {
		return m.Type
	}
	return nil
}

func (x *Message) GetBlockRequest() *BlockRequest {
	if x, ok := x.GetType().(*Message_BlockRequest); ok {
		return x.BlockRequest
	}
	return nil
}

func (x *Message) GetBlockResponse() *BlockResponse {
	if x, ok := x.GetType().(*Message_BlockResponse); ok {
		return x.BlockResponse
	}
	return nil
}

type isMessage_Type interface {
	isMessage_Type()
}

type Message_BlockRequest struct {
	BlockRequest *BlockRequest `protobuf:"bytes,1,opt,name=block_request,json=blockRequest,proto3,oneof"`
}

type Message_BlockResponse struct {
	BlockResponse *BlockResponse `protobuf:"bytes,2,opt,name=block_response,json=blockResponse,proto3,oneof"`
}

func (*Message_BlockRequest) isMessage_Type() {}

func (*Message_BlockResponse) isMessage_Type() {}

type BlockRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	RequestId uint64 `protobuf:"varint,1,opt,name=request_id,json=requestId,proto3" json:"request_id,omitempty"`
	BlockId   uint64 `protobuf:"varint,2,opt,name=block_id,json=blockId,proto3" json:"block_id,omitempty"`
}

func (x *BlockRequest) Reset() {
	*x = BlockRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_blockchainpb_synchronizerpb_synchronizerpb_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BlockRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BlockRequest) ProtoMessage() {}

func (x *BlockRequest) ProtoReflect() protoreflect.Message {
	mi := &file_blockchainpb_synchronizerpb_synchronizerpb_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BlockRequest.ProtoReflect.Descriptor instead.
func (*BlockRequest) Descriptor() ([]byte, []int) {
	return file_blockchainpb_synchronizerpb_synchronizerpb_proto_rawDescGZIP(), []int{3}
}

func (x *BlockRequest) GetRequestId() uint64 {
	if x != nil {
		return x.RequestId
	}
	return 0
}

func (x *BlockRequest) GetBlockId() uint64 {
	if x != nil {
		return x.BlockId
	}
	return 0
}

type BlockResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	RequestId uint64              `protobuf:"varint,1,opt,name=request_id,json=requestId,proto3" json:"request_id,omitempty"`
	Found     bool                `protobuf:"varint,2,opt,name=found,proto3" json:"found,omitempty"`
	Block     *blockchainpb.Block `protobuf:"bytes,3,opt,name=block,proto3" json:"block,omitempty"` // possibly undefined (proto3 spec no-longer supports optional/required...)
}

func (x *BlockResponse) Reset() {
	*x = BlockResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_blockchainpb_synchronizerpb_synchronizerpb_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BlockResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BlockResponse) ProtoMessage() {}

func (x *BlockResponse) ProtoReflect() protoreflect.Message {
	mi := &file_blockchainpb_synchronizerpb_synchronizerpb_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BlockResponse.ProtoReflect.Descriptor instead.
func (*BlockResponse) Descriptor() ([]byte, []int) {
	return file_blockchainpb_synchronizerpb_synchronizerpb_proto_rawDescGZIP(), []int{4}
}

func (x *BlockResponse) GetRequestId() uint64 {
	if x != nil {
		return x.RequestId
	}
	return 0
}

func (x *BlockResponse) GetFound() bool {
	if x != nil {
		return x.Found
	}
	return false
}

func (x *BlockResponse) GetBlock() *blockchainpb.Block {
	if x != nil {
		return x.Block
	}
	return nil
}

var File_blockchainpb_synchronizerpb_synchronizerpb_proto protoreflect.FileDescriptor

var file_blockchainpb_synchronizerpb_synchronizerpb_proto_rawDesc = []byte{
	0x0a, 0x30, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x70, 0x62, 0x2f, 0x73,
	0x79, 0x6e, 0x63, 0x68, 0x72, 0x6f, 0x6e, 0x69, 0x7a, 0x65, 0x72, 0x70, 0x62, 0x2f, 0x73, 0x79,
	0x6e, 0x63, 0x68, 0x72, 0x6f, 0x6e, 0x69, 0x7a, 0x65, 0x72, 0x70, 0x62, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x12, 0x0e, 0x73, 0x79, 0x6e, 0x63, 0x68, 0x72, 0x6f, 0x6e, 0x69, 0x7a, 0x65, 0x72,
	0x70, 0x62, 0x1a, 0x1f, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x70, 0x62,
	0x2f, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x70, 0x62, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x1a, 0x1c, 0x6d, 0x69, 0x72, 0x2f, 0x63, 0x6f, 0x64, 0x65, 0x67, 0x65, 0x6e,
	0x5f, 0x65, 0x78, 0x74, 0x65, 0x6e, 0x73, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x1a, 0x1c, 0x6e, 0x65, 0x74, 0x2f, 0x63, 0x6f, 0x64, 0x65, 0x67, 0x65, 0x6e, 0x5f, 0x65,
	0x78, 0x74, 0x65, 0x6e, 0x73, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22,
	0x5d, 0x0a, 0x05, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x12, 0x40, 0x0a, 0x0c, 0x73, 0x79, 0x6e, 0x63,
	0x5f, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1b,
	0x2e, 0x73, 0x79, 0x6e, 0x63, 0x68, 0x72, 0x6f, 0x6e, 0x69, 0x7a, 0x65, 0x72, 0x70, 0x62, 0x2e,
	0x53, 0x79, 0x6e, 0x63, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x48, 0x00, 0x52, 0x0b, 0x73,
	0x79, 0x6e, 0x63, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x3a, 0x04, 0x90, 0xa6, 0x1d, 0x01,
	0x42, 0x0c, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65, 0x12, 0x04, 0x80, 0xa6, 0x1d, 0x01, 0x22, 0x68,
	0x0a, 0x0b, 0x53, 0x79, 0x6e, 0x63, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x36, 0x0a,
	0x0c, 0x6f, 0x72, 0x70, 0x68, 0x61, 0x6e, 0x5f, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x13, 0x2e, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x63, 0x68, 0x61, 0x69, 0x6e,
	0x70, 0x62, 0x2e, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x52, 0x0b, 0x6f, 0x72, 0x70, 0x68, 0x61, 0x6e,
	0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x12, 0x1b, 0x0a, 0x09, 0x6c, 0x65, 0x61, 0x76, 0x65, 0x5f, 0x69,
	0x64, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x04, 0x52, 0x08, 0x6c, 0x65, 0x61, 0x76, 0x65, 0x49,
	0x64, 0x73, 0x3a, 0x04, 0x98, 0xa6, 0x1d, 0x01, 0x22, 0xaa, 0x01, 0x0a, 0x07, 0x4d, 0x65, 0x73,
	0x73, 0x61, 0x67, 0x65, 0x12, 0x43, 0x0a, 0x0d, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x5f, 0x72, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1c, 0x2e, 0x73, 0x79,
	0x6e, 0x63, 0x68, 0x72, 0x6f, 0x6e, 0x69, 0x7a, 0x65, 0x72, 0x70, 0x62, 0x2e, 0x42, 0x6c, 0x6f,
	0x63, 0x6b, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x48, 0x00, 0x52, 0x0c, 0x62, 0x6c, 0x6f,
	0x63, 0x6b, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x46, 0x0a, 0x0e, 0x62, 0x6c, 0x6f,
	0x63, 0x6b, 0x5f, 0x72, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x1d, 0x2e, 0x73, 0x79, 0x6e, 0x63, 0x68, 0x72, 0x6f, 0x6e, 0x69, 0x7a, 0x65, 0x72,
	0x70, 0x62, 0x2e, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x48, 0x00, 0x52, 0x0d, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x3a, 0x04, 0xc8, 0xe4, 0x1d, 0x01, 0x42, 0x0c, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65, 0x12,
	0x04, 0xc8, 0xe4, 0x1d, 0x01, 0x22, 0x4e, 0x0a, 0x0c, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x1d, 0x0a, 0x0a, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x04, 0x52, 0x09, 0x72, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x49, 0x64, 0x12, 0x19, 0x0a, 0x08, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x5f, 0x69, 0x64,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x04, 0x52, 0x07, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x49, 0x64, 0x3a,
	0x04, 0xd0, 0xe4, 0x1d, 0x01, 0x22, 0x75, 0x0a, 0x0d, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x1d, 0x0a, 0x0a, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x04, 0x52, 0x09, 0x72, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x49, 0x64, 0x12, 0x14, 0x0a, 0x05, 0x66, 0x6f, 0x75, 0x6e, 0x64, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x08, 0x52, 0x05, 0x66, 0x6f, 0x75, 0x6e, 0x64, 0x12, 0x29, 0x0a, 0x05, 0x62,
	0x6c, 0x6f, 0x63, 0x6b, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x13, 0x2e, 0x62, 0x6c, 0x6f,
	0x63, 0x6b, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x70, 0x62, 0x2e, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x52,
	0x05, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x3a, 0x04, 0xd0, 0xe4, 0x1d, 0x01, 0x42, 0x44, 0x5a, 0x42,
	0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x66, 0x69, 0x6c, 0x65, 0x63,
	0x6f, 0x69, 0x6e, 0x2d, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x2f, 0x6d, 0x69, 0x72, 0x2f,
	0x70, 0x6b, 0x67, 0x2f, 0x70, 0x62, 0x2f, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x63, 0x68, 0x61, 0x69,
	0x6e, 0x70, 0x62, 0x2f, 0x73, 0x79, 0x6e, 0x63, 0x68, 0x72, 0x6f, 0x6e, 0x69, 0x7a, 0x65, 0x72,
	0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_blockchainpb_synchronizerpb_synchronizerpb_proto_rawDescOnce sync.Once
	file_blockchainpb_synchronizerpb_synchronizerpb_proto_rawDescData = file_blockchainpb_synchronizerpb_synchronizerpb_proto_rawDesc
)

func file_blockchainpb_synchronizerpb_synchronizerpb_proto_rawDescGZIP() []byte {
	file_blockchainpb_synchronizerpb_synchronizerpb_proto_rawDescOnce.Do(func() {
		file_blockchainpb_synchronizerpb_synchronizerpb_proto_rawDescData = protoimpl.X.CompressGZIP(file_blockchainpb_synchronizerpb_synchronizerpb_proto_rawDescData)
	})
	return file_blockchainpb_synchronizerpb_synchronizerpb_proto_rawDescData
}

var file_blockchainpb_synchronizerpb_synchronizerpb_proto_msgTypes = make([]protoimpl.MessageInfo, 5)
var file_blockchainpb_synchronizerpb_synchronizerpb_proto_goTypes = []interface{}{
	(*Event)(nil),              // 0: synchronizerpb.Event
	(*SyncRequest)(nil),        // 1: synchronizerpb.SyncRequest
	(*Message)(nil),            // 2: synchronizerpb.Message
	(*BlockRequest)(nil),       // 3: synchronizerpb.BlockRequest
	(*BlockResponse)(nil),      // 4: synchronizerpb.BlockResponse
	(*blockchainpb.Block)(nil), // 5: blockchainpb.Block
}
var file_blockchainpb_synchronizerpb_synchronizerpb_proto_depIdxs = []int32{
	1, // 0: synchronizerpb.Event.sync_request:type_name -> synchronizerpb.SyncRequest
	5, // 1: synchronizerpb.SyncRequest.orphan_block:type_name -> blockchainpb.Block
	3, // 2: synchronizerpb.Message.block_request:type_name -> synchronizerpb.BlockRequest
	4, // 3: synchronizerpb.Message.block_response:type_name -> synchronizerpb.BlockResponse
	5, // 4: synchronizerpb.BlockResponse.block:type_name -> blockchainpb.Block
	5, // [5:5] is the sub-list for method output_type
	5, // [5:5] is the sub-list for method input_type
	5, // [5:5] is the sub-list for extension type_name
	5, // [5:5] is the sub-list for extension extendee
	0, // [0:5] is the sub-list for field type_name
}

func init() { file_blockchainpb_synchronizerpb_synchronizerpb_proto_init() }
func file_blockchainpb_synchronizerpb_synchronizerpb_proto_init() {
	if File_blockchainpb_synchronizerpb_synchronizerpb_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_blockchainpb_synchronizerpb_synchronizerpb_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Event); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_blockchainpb_synchronizerpb_synchronizerpb_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SyncRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_blockchainpb_synchronizerpb_synchronizerpb_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Message); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_blockchainpb_synchronizerpb_synchronizerpb_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BlockRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_blockchainpb_synchronizerpb_synchronizerpb_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BlockResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	file_blockchainpb_synchronizerpb_synchronizerpb_proto_msgTypes[0].OneofWrappers = []interface{}{
		(*Event_SyncRequest)(nil),
	}
	file_blockchainpb_synchronizerpb_synchronizerpb_proto_msgTypes[2].OneofWrappers = []interface{}{
		(*Message_BlockRequest)(nil),
		(*Message_BlockResponse)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_blockchainpb_synchronizerpb_synchronizerpb_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   5,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_blockchainpb_synchronizerpb_synchronizerpb_proto_goTypes,
		DependencyIndexes: file_blockchainpb_synchronizerpb_synchronizerpb_proto_depIdxs,
		MessageInfos:      file_blockchainpb_synchronizerpb_synchronizerpb_proto_msgTypes,
	}.Build()
	File_blockchainpb_synchronizerpb_synchronizerpb_proto = out.File
	file_blockchainpb_synchronizerpb_synchronizerpb_proto_rawDesc = nil
	file_blockchainpb_synchronizerpb_synchronizerpb_proto_goTypes = nil
	file_blockchainpb_synchronizerpb_synchronizerpb_proto_depIdxs = nil
}
