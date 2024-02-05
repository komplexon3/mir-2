// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        v4.24.4
// source: protos-blockchain/minerpb/minerpb.proto

package minerpb

import (
	payloadpb "github.com/filecoin-project/mir/pkg-blockchain/pb/payloadpb"
	_ "github.com/filecoin-project/mir/pkg/pb/mir"
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
	//	*Event_BlockRequest
	//	*Event_NewHead
	Type isEvent_Type `protobuf_oneof:"type"`
}

func (x *Event) Reset() {
	*x = Event{}
	if protoimpl.UnsafeEnabled {
		mi := &file_protos_blockchain_minerpb_minerpb_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Event) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Event) ProtoMessage() {}

func (x *Event) ProtoReflect() protoreflect.Message {
	mi := &file_protos_blockchain_minerpb_minerpb_proto_msgTypes[0]
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
	return file_protos_blockchain_minerpb_minerpb_proto_rawDescGZIP(), []int{0}
}

func (m *Event) GetType() isEvent_Type {
	if m != nil {
		return m.Type
	}
	return nil
}

func (x *Event) GetBlockRequest() *BlockRequest {
	if x, ok := x.GetType().(*Event_BlockRequest); ok {
		return x.BlockRequest
	}
	return nil
}

func (x *Event) GetNewHead() *NewHead {
	if x, ok := x.GetType().(*Event_NewHead); ok {
		return x.NewHead
	}
	return nil
}

type isEvent_Type interface {
	isEvent_Type()
}

type Event_BlockRequest struct {
	BlockRequest *BlockRequest `protobuf:"bytes,1,opt,name=block_request,json=blockRequest,proto3,oneof"`
}

type Event_NewHead struct {
	NewHead *NewHead `protobuf:"bytes,2,opt,name=new_head,json=newHead,proto3,oneof"`
}

func (*Event_BlockRequest) isEvent_Type() {}

func (*Event_NewHead) isEvent_Type() {}

type BlockRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	HeadId  uint64             `protobuf:"varint,1,opt,name=head_id,json=headId,proto3" json:"head_id,omitempty"`
	Payload *payloadpb.Payload `protobuf:"bytes,2,opt,name=payload,proto3" json:"payload,omitempty"`
}

func (x *BlockRequest) Reset() {
	*x = BlockRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_protos_blockchain_minerpb_minerpb_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BlockRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BlockRequest) ProtoMessage() {}

func (x *BlockRequest) ProtoReflect() protoreflect.Message {
	mi := &file_protos_blockchain_minerpb_minerpb_proto_msgTypes[1]
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
	return file_protos_blockchain_minerpb_minerpb_proto_rawDescGZIP(), []int{1}
}

func (x *BlockRequest) GetHeadId() uint64 {
	if x != nil {
		return x.HeadId
	}
	return 0
}

func (x *BlockRequest) GetPayload() *payloadpb.Payload {
	if x != nil {
		return x.Payload
	}
	return nil
}

type NewHead struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	HeadId uint64 `protobuf:"varint,1,opt,name=head_id,json=headId,proto3" json:"head_id,omitempty"`
}

func (x *NewHead) Reset() {
	*x = NewHead{}
	if protoimpl.UnsafeEnabled {
		mi := &file_protos_blockchain_minerpb_minerpb_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *NewHead) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*NewHead) ProtoMessage() {}

func (x *NewHead) ProtoReflect() protoreflect.Message {
	mi := &file_protos_blockchain_minerpb_minerpb_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use NewHead.ProtoReflect.Descriptor instead.
func (*NewHead) Descriptor() ([]byte, []int) {
	return file_protos_blockchain_minerpb_minerpb_proto_rawDescGZIP(), []int{2}
}

func (x *NewHead) GetHeadId() uint64 {
	if x != nil {
		return x.HeadId
	}
	return 0
}

var File_protos_blockchain_minerpb_minerpb_proto protoreflect.FileDescriptor

var file_protos_blockchain_minerpb_minerpb_proto_rawDesc = []byte{
	0x0a, 0x27, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x73, 0x2d, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x63, 0x68,
	0x61, 0x69, 0x6e, 0x2f, 0x6d, 0x69, 0x6e, 0x65, 0x72, 0x70, 0x62, 0x2f, 0x6d, 0x69, 0x6e, 0x65,
	0x72, 0x70, 0x62, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x07, 0x6d, 0x69, 0x6e, 0x65, 0x72,
	0x70, 0x62, 0x1a, 0x23, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x73, 0x2f, 0x6d, 0x69, 0x72, 0x2f, 0x63,
	0x6f, 0x64, 0x65, 0x67, 0x65, 0x6e, 0x5f, 0x65, 0x78, 0x74, 0x65, 0x6e, 0x73, 0x69, 0x6f, 0x6e,
	0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x2b, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x73, 0x2d,
	0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x2f, 0x70, 0x61, 0x79, 0x6c, 0x6f,
	0x61, 0x64, 0x70, 0x62, 0x2f, 0x70, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x70, 0x62, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x22, 0x88, 0x01, 0x0a, 0x05, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x12, 0x3c,
	0x0a, 0x0d, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x5f, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x15, 0x2e, 0x6d, 0x69, 0x6e, 0x65, 0x72, 0x70, 0x62, 0x2e,
	0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x48, 0x00, 0x52, 0x0c,
	0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x2d, 0x0a, 0x08,
	0x6e, 0x65, 0x77, 0x5f, 0x68, 0x65, 0x61, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x10,
	0x2e, 0x6d, 0x69, 0x6e, 0x65, 0x72, 0x70, 0x62, 0x2e, 0x4e, 0x65, 0x77, 0x48, 0x65, 0x61, 0x64,
	0x48, 0x00, 0x52, 0x07, 0x6e, 0x65, 0x77, 0x48, 0x65, 0x61, 0x64, 0x3a, 0x04, 0x90, 0xa6, 0x1d,
	0x01, 0x42, 0x0c, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65, 0x12, 0x04, 0x80, 0xa6, 0x1d, 0x01, 0x22,
	0x5b, 0x0a, 0x0c, 0x42, 0x6c, 0x6f, 0x63, 0x6b, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12,
	0x17, 0x0a, 0x07, 0x68, 0x65, 0x61, 0x64, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x04,
	0x52, 0x06, 0x68, 0x65, 0x61, 0x64, 0x49, 0x64, 0x12, 0x2c, 0x0a, 0x07, 0x70, 0x61, 0x79, 0x6c,
	0x6f, 0x61, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x12, 0x2e, 0x70, 0x61, 0x79, 0x6c,
	0x6f, 0x61, 0x64, 0x70, 0x62, 0x2e, 0x50, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x52, 0x07, 0x70,
	0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x3a, 0x04, 0x98, 0xa6, 0x1d, 0x01, 0x22, 0x28, 0x0a, 0x07,
	0x4e, 0x65, 0x77, 0x48, 0x65, 0x61, 0x64, 0x12, 0x17, 0x0a, 0x07, 0x68, 0x65, 0x61, 0x64, 0x5f,
	0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x04, 0x52, 0x06, 0x68, 0x65, 0x61, 0x64, 0x49, 0x64,
	0x3a, 0x04, 0x98, 0xa6, 0x1d, 0x01, 0x42, 0x3b, 0x5a, 0x39, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62,
	0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x66, 0x69, 0x6c, 0x65, 0x63, 0x6f, 0x69, 0x6e, 0x2d, 0x70, 0x72,
	0x6f, 0x6a, 0x65, 0x63, 0x74, 0x2f, 0x6d, 0x69, 0x72, 0x2f, 0x70, 0x6b, 0x67, 0x2d, 0x62, 0x6c,
	0x6f, 0x63, 0x6b, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x2f, 0x70, 0x62, 0x2f, 0x6d, 0x69, 0x6e, 0x65,
	0x72, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_protos_blockchain_minerpb_minerpb_proto_rawDescOnce sync.Once
	file_protos_blockchain_minerpb_minerpb_proto_rawDescData = file_protos_blockchain_minerpb_minerpb_proto_rawDesc
)

func file_protos_blockchain_minerpb_minerpb_proto_rawDescGZIP() []byte {
	file_protos_blockchain_minerpb_minerpb_proto_rawDescOnce.Do(func() {
		file_protos_blockchain_minerpb_minerpb_proto_rawDescData = protoimpl.X.CompressGZIP(file_protos_blockchain_minerpb_minerpb_proto_rawDescData)
	})
	return file_protos_blockchain_minerpb_minerpb_proto_rawDescData
}

var file_protos_blockchain_minerpb_minerpb_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_protos_blockchain_minerpb_minerpb_proto_goTypes = []interface{}{
	(*Event)(nil),             // 0: minerpb.Event
	(*BlockRequest)(nil),      // 1: minerpb.BlockRequest
	(*NewHead)(nil),           // 2: minerpb.NewHead
	(*payloadpb.Payload)(nil), // 3: payloadpb.Payload
}
var file_protos_blockchain_minerpb_minerpb_proto_depIdxs = []int32{
	1, // 0: minerpb.Event.block_request:type_name -> minerpb.BlockRequest
	2, // 1: minerpb.Event.new_head:type_name -> minerpb.NewHead
	3, // 2: minerpb.BlockRequest.payload:type_name -> payloadpb.Payload
	3, // [3:3] is the sub-list for method output_type
	3, // [3:3] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_protos_blockchain_minerpb_minerpb_proto_init() }
func file_protos_blockchain_minerpb_minerpb_proto_init() {
	if File_protos_blockchain_minerpb_minerpb_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_protos_blockchain_minerpb_minerpb_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
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
		file_protos_blockchain_minerpb_minerpb_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
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
		file_protos_blockchain_minerpb_minerpb_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*NewHead); i {
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
	file_protos_blockchain_minerpb_minerpb_proto_msgTypes[0].OneofWrappers = []interface{}{
		(*Event_BlockRequest)(nil),
		(*Event_NewHead)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_protos_blockchain_minerpb_minerpb_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_protos_blockchain_minerpb_minerpb_proto_goTypes,
		DependencyIndexes: file_protos_blockchain_minerpb_minerpb_proto_depIdxs,
		MessageInfos:      file_protos_blockchain_minerpb_minerpb_proto_msgTypes,
	}.Build()
	File_protos_blockchain_minerpb_minerpb_proto = out.File
	file_protos_blockchain_minerpb_minerpb_proto_rawDesc = nil
	file_protos_blockchain_minerpb_minerpb_proto_goTypes = nil
	file_protos_blockchain_minerpb_minerpb_proto_depIdxs = nil
}
