// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        v4.24.4
// source: apppb/apppb.proto

package apppb

import (
	checkpointpb "github.com/filecoin-project/mir/pkg/pb/checkpointpb"
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
	//	*Event_SnapshotRequest
	//	*Event_Snapshot
	//	*Event_RestoreState
	//	*Event_NewEpoch
	Type isEvent_Type `protobuf_oneof:"type"`
}

func (x *Event) Reset() {
	*x = Event{}
	if protoimpl.UnsafeEnabled {
		mi := &file_apppb_apppb_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Event) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Event) ProtoMessage() {}

func (x *Event) ProtoReflect() protoreflect.Message {
	mi := &file_apppb_apppb_proto_msgTypes[0]
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
	return file_apppb_apppb_proto_rawDescGZIP(), []int{0}
}

func (m *Event) GetType() isEvent_Type {
	if m != nil {
		return m.Type
	}
	return nil
}

func (x *Event) GetSnapshotRequest() *SnapshotRequest {
	if x, ok := x.GetType().(*Event_SnapshotRequest); ok {
		return x.SnapshotRequest
	}
	return nil
}

func (x *Event) GetSnapshot() *Snapshot {
	if x, ok := x.GetType().(*Event_Snapshot); ok {
		return x.Snapshot
	}
	return nil
}

func (x *Event) GetRestoreState() *RestoreState {
	if x, ok := x.GetType().(*Event_RestoreState); ok {
		return x.RestoreState
	}
	return nil
}

func (x *Event) GetNewEpoch() *NewEpoch {
	if x, ok := x.GetType().(*Event_NewEpoch); ok {
		return x.NewEpoch
	}
	return nil
}

type isEvent_Type interface {
	isEvent_Type()
}

type Event_SnapshotRequest struct {
	SnapshotRequest *SnapshotRequest `protobuf:"bytes,1,opt,name=snapshot_request,json=snapshotRequest,proto3,oneof"`
}

type Event_Snapshot struct {
	Snapshot *Snapshot `protobuf:"bytes,2,opt,name=snapshot,proto3,oneof"`
}

type Event_RestoreState struct {
	RestoreState *RestoreState `protobuf:"bytes,3,opt,name=restore_state,json=restoreState,proto3,oneof"`
}

type Event_NewEpoch struct {
	NewEpoch *NewEpoch `protobuf:"bytes,4,opt,name=new_epoch,json=newEpoch,proto3,oneof"`
}

func (*Event_SnapshotRequest) isEvent_Type() {}

func (*Event_Snapshot) isEvent_Type() {}

func (*Event_RestoreState) isEvent_Type() {}

func (*Event_NewEpoch) isEvent_Type() {}

type SnapshotRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ReplyTo string `protobuf:"bytes,1,opt,name=reply_to,json=replyTo,proto3" json:"reply_to,omitempty"`
}

func (x *SnapshotRequest) Reset() {
	*x = SnapshotRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_apppb_apppb_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SnapshotRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SnapshotRequest) ProtoMessage() {}

func (x *SnapshotRequest) ProtoReflect() protoreflect.Message {
	mi := &file_apppb_apppb_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SnapshotRequest.ProtoReflect.Descriptor instead.
func (*SnapshotRequest) Descriptor() ([]byte, []int) {
	return file_apppb_apppb_proto_rawDescGZIP(), []int{1}
}

func (x *SnapshotRequest) GetReplyTo() string {
	if x != nil {
		return x.ReplyTo
	}
	return ""
}

type Snapshot struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	AppData []byte `protobuf:"bytes,1,opt,name=app_data,json=appData,proto3" json:"app_data,omitempty"`
}

func (x *Snapshot) Reset() {
	*x = Snapshot{}
	if protoimpl.UnsafeEnabled {
		mi := &file_apppb_apppb_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Snapshot) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Snapshot) ProtoMessage() {}

func (x *Snapshot) ProtoReflect() protoreflect.Message {
	mi := &file_apppb_apppb_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Snapshot.ProtoReflect.Descriptor instead.
func (*Snapshot) Descriptor() ([]byte, []int) {
	return file_apppb_apppb_proto_rawDescGZIP(), []int{2}
}

func (x *Snapshot) GetAppData() []byte {
	if x != nil {
		return x.AppData
	}
	return nil
}

type RestoreState struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Checkpoint *checkpointpb.StableCheckpoint `protobuf:"bytes,1,opt,name=checkpoint,proto3" json:"checkpoint,omitempty"`
}

func (x *RestoreState) Reset() {
	*x = RestoreState{}
	if protoimpl.UnsafeEnabled {
		mi := &file_apppb_apppb_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RestoreState) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RestoreState) ProtoMessage() {}

func (x *RestoreState) ProtoReflect() protoreflect.Message {
	mi := &file_apppb_apppb_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RestoreState.ProtoReflect.Descriptor instead.
func (*RestoreState) Descriptor() ([]byte, []int) {
	return file_apppb_apppb_proto_rawDescGZIP(), []int{3}
}

func (x *RestoreState) GetCheckpoint() *checkpointpb.StableCheckpoint {
	if x != nil {
		return x.Checkpoint
	}
	return nil
}

type NewEpoch struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	EpochNr        uint64 `protobuf:"varint,1,opt,name=epoch_nr,json=epochNr,proto3" json:"epoch_nr,omitempty"`
	ProtocolModule string `protobuf:"bytes,2,opt,name=protocol_module,json=protocolModule,proto3" json:"protocol_module,omitempty"`
}

func (x *NewEpoch) Reset() {
	*x = NewEpoch{}
	if protoimpl.UnsafeEnabled {
		mi := &file_apppb_apppb_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *NewEpoch) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*NewEpoch) ProtoMessage() {}

func (x *NewEpoch) ProtoReflect() protoreflect.Message {
	mi := &file_apppb_apppb_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use NewEpoch.ProtoReflect.Descriptor instead.
func (*NewEpoch) Descriptor() ([]byte, []int) {
	return file_apppb_apppb_proto_rawDescGZIP(), []int{4}
}

func (x *NewEpoch) GetEpochNr() uint64 {
	if x != nil {
		return x.EpochNr
	}
	return 0
}

func (x *NewEpoch) GetProtocolModule() string {
	if x != nil {
		return x.ProtocolModule
	}
	return ""
}

var File_apppb_apppb_proto protoreflect.FileDescriptor

var file_apppb_apppb_proto_rawDesc = []byte{
	0x0a, 0x11, 0x61, 0x70, 0x70, 0x70, 0x62, 0x2f, 0x61, 0x70, 0x70, 0x70, 0x62, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x12, 0x05, 0x61, 0x70, 0x70, 0x70, 0x62, 0x1a, 0x1f, 0x63, 0x68, 0x65, 0x63,
	0x6b, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x70, 0x62, 0x2f, 0x63, 0x68, 0x65, 0x63, 0x6b, 0x70, 0x6f,
	0x69, 0x6e, 0x74, 0x70, 0x62, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1c, 0x6d, 0x69, 0x72,
	0x2f, 0x63, 0x6f, 0x64, 0x65, 0x67, 0x65, 0x6e, 0x5f, 0x65, 0x78, 0x74, 0x65, 0x6e, 0x73, 0x69,
	0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xfb, 0x01, 0x0a, 0x05, 0x45, 0x76,
	0x65, 0x6e, 0x74, 0x12, 0x43, 0x0a, 0x10, 0x73, 0x6e, 0x61, 0x70, 0x73, 0x68, 0x6f, 0x74, 0x5f,
	0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x16, 0x2e,
	0x61, 0x70, 0x70, 0x70, 0x62, 0x2e, 0x53, 0x6e, 0x61, 0x70, 0x73, 0x68, 0x6f, 0x74, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x48, 0x00, 0x52, 0x0f, 0x73, 0x6e, 0x61, 0x70, 0x73, 0x68, 0x6f,
	0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x2d, 0x0a, 0x08, 0x73, 0x6e, 0x61, 0x70,
	0x73, 0x68, 0x6f, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0f, 0x2e, 0x61, 0x70, 0x70,
	0x70, 0x62, 0x2e, 0x53, 0x6e, 0x61, 0x70, 0x73, 0x68, 0x6f, 0x74, 0x48, 0x00, 0x52, 0x08, 0x73,
	0x6e, 0x61, 0x70, 0x73, 0x68, 0x6f, 0x74, 0x12, 0x3a, 0x0a, 0x0d, 0x72, 0x65, 0x73, 0x74, 0x6f,
	0x72, 0x65, 0x5f, 0x73, 0x74, 0x61, 0x74, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x13,
	0x2e, 0x61, 0x70, 0x70, 0x70, 0x62, 0x2e, 0x52, 0x65, 0x73, 0x74, 0x6f, 0x72, 0x65, 0x53, 0x74,
	0x61, 0x74, 0x65, 0x48, 0x00, 0x52, 0x0c, 0x72, 0x65, 0x73, 0x74, 0x6f, 0x72, 0x65, 0x53, 0x74,
	0x61, 0x74, 0x65, 0x12, 0x2e, 0x0a, 0x09, 0x6e, 0x65, 0x77, 0x5f, 0x65, 0x70, 0x6f, 0x63, 0x68,
	0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0f, 0x2e, 0x61, 0x70, 0x70, 0x70, 0x62, 0x2e, 0x4e,
	0x65, 0x77, 0x45, 0x70, 0x6f, 0x63, 0x68, 0x48, 0x00, 0x52, 0x08, 0x6e, 0x65, 0x77, 0x45, 0x70,
	0x6f, 0x63, 0x68, 0x3a, 0x04, 0x90, 0xa6, 0x1d, 0x01, 0x42, 0x0c, 0x0a, 0x04, 0x74, 0x79, 0x70,
	0x65, 0x12, 0x04, 0x80, 0xa6, 0x1d, 0x01, 0x22, 0x69, 0x0a, 0x0f, 0x53, 0x6e, 0x61, 0x70, 0x73,
	0x68, 0x6f, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x50, 0x0a, 0x08, 0x72, 0x65,
	0x70, 0x6c, 0x79, 0x5f, 0x74, 0x6f, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x42, 0x35, 0x82, 0xa6,
	0x1d, 0x31, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x66, 0x69, 0x6c,
	0x65, 0x63, 0x6f, 0x69, 0x6e, 0x2d, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x2f, 0x6d, 0x69,
	0x72, 0x2f, 0x73, 0x74, 0x64, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2e, 0x4d, 0x6f, 0x64, 0x75, 0x6c,
	0x65, 0x49, 0x44, 0x52, 0x07, 0x72, 0x65, 0x70, 0x6c, 0x79, 0x54, 0x6f, 0x3a, 0x04, 0x98, 0xa6,
	0x1d, 0x01, 0x22, 0x2b, 0x0a, 0x08, 0x53, 0x6e, 0x61, 0x70, 0x73, 0x68, 0x6f, 0x74, 0x12, 0x19,
	0x0a, 0x08, 0x61, 0x70, 0x70, 0x5f, 0x64, 0x61, 0x74, 0x61, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c,
	0x52, 0x07, 0x61, 0x70, 0x70, 0x44, 0x61, 0x74, 0x61, 0x3a, 0x04, 0x98, 0xa6, 0x1d, 0x01, 0x22,
	0x54, 0x0a, 0x0c, 0x52, 0x65, 0x73, 0x74, 0x6f, 0x72, 0x65, 0x53, 0x74, 0x61, 0x74, 0x65, 0x12,
	0x3e, 0x0a, 0x0a, 0x63, 0x68, 0x65, 0x63, 0x6b, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x1e, 0x2e, 0x63, 0x68, 0x65, 0x63, 0x6b, 0x70, 0x6f, 0x69, 0x6e, 0x74,
	0x70, 0x62, 0x2e, 0x53, 0x74, 0x61, 0x62, 0x6c, 0x65, 0x43, 0x68, 0x65, 0x63, 0x6b, 0x70, 0x6f,
	0x69, 0x6e, 0x74, 0x52, 0x0a, 0x63, 0x68, 0x65, 0x63, 0x6b, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x3a,
	0x04, 0x98, 0xa6, 0x1d, 0x01, 0x22, 0xca, 0x01, 0x0a, 0x08, 0x4e, 0x65, 0x77, 0x45, 0x70, 0x6f,
	0x63, 0x68, 0x12, 0x58, 0x0a, 0x08, 0x65, 0x70, 0x6f, 0x63, 0x68, 0x5f, 0x6e, 0x72, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x04, 0x42, 0x3d, 0x82, 0xa6, 0x1d, 0x39, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62,
	0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x66, 0x69, 0x6c, 0x65, 0x63, 0x6f, 0x69, 0x6e, 0x2d, 0x70, 0x72,
	0x6f, 0x6a, 0x65, 0x63, 0x74, 0x2f, 0x6d, 0x69, 0x72, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x74, 0x72,
	0x61, 0x6e, 0x74, 0x6f, 0x72, 0x2f, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2e, 0x45, 0x70, 0x6f, 0x63,
	0x68, 0x4e, 0x72, 0x52, 0x07, 0x65, 0x70, 0x6f, 0x63, 0x68, 0x4e, 0x72, 0x12, 0x5e, 0x0a, 0x0f,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x5f, 0x6d, 0x6f, 0x64, 0x75, 0x6c, 0x65, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x09, 0x42, 0x35, 0x82, 0xa6, 0x1d, 0x31, 0x67, 0x69, 0x74, 0x68, 0x75,
	0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x66, 0x69, 0x6c, 0x65, 0x63, 0x6f, 0x69, 0x6e, 0x2d, 0x70,
	0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x2f, 0x6d, 0x69, 0x72, 0x2f, 0x73, 0x74, 0x64, 0x74, 0x79,
	0x70, 0x65, 0x73, 0x2e, 0x4d, 0x6f, 0x64, 0x75, 0x6c, 0x65, 0x49, 0x44, 0x52, 0x0e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x4d, 0x6f, 0x64, 0x75, 0x6c, 0x65, 0x3a, 0x04, 0x98, 0xa6,
	0x1d, 0x01, 0x42, 0x2e, 0x5a, 0x2c, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d,
	0x2f, 0x66, 0x69, 0x6c, 0x65, 0x63, 0x6f, 0x69, 0x6e, 0x2d, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63,
	0x74, 0x2f, 0x6d, 0x69, 0x72, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x70, 0x62, 0x2f, 0x61, 0x70, 0x70,
	0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_apppb_apppb_proto_rawDescOnce sync.Once
	file_apppb_apppb_proto_rawDescData = file_apppb_apppb_proto_rawDesc
)

func file_apppb_apppb_proto_rawDescGZIP() []byte {
	file_apppb_apppb_proto_rawDescOnce.Do(func() {
		file_apppb_apppb_proto_rawDescData = protoimpl.X.CompressGZIP(file_apppb_apppb_proto_rawDescData)
	})
	return file_apppb_apppb_proto_rawDescData
}

var file_apppb_apppb_proto_msgTypes = make([]protoimpl.MessageInfo, 5)
var file_apppb_apppb_proto_goTypes = []interface{}{
	(*Event)(nil),                         // 0: apppb.Event
	(*SnapshotRequest)(nil),               // 1: apppb.SnapshotRequest
	(*Snapshot)(nil),                      // 2: apppb.Snapshot
	(*RestoreState)(nil),                  // 3: apppb.RestoreState
	(*NewEpoch)(nil),                      // 4: apppb.NewEpoch
	(*checkpointpb.StableCheckpoint)(nil), // 5: checkpointpb.StableCheckpoint
}
var file_apppb_apppb_proto_depIdxs = []int32{
	1, // 0: apppb.Event.snapshot_request:type_name -> apppb.SnapshotRequest
	2, // 1: apppb.Event.snapshot:type_name -> apppb.Snapshot
	3, // 2: apppb.Event.restore_state:type_name -> apppb.RestoreState
	4, // 3: apppb.Event.new_epoch:type_name -> apppb.NewEpoch
	5, // 4: apppb.RestoreState.checkpoint:type_name -> checkpointpb.StableCheckpoint
	5, // [5:5] is the sub-list for method output_type
	5, // [5:5] is the sub-list for method input_type
	5, // [5:5] is the sub-list for extension type_name
	5, // [5:5] is the sub-list for extension extendee
	0, // [0:5] is the sub-list for field type_name
}

func init() { file_apppb_apppb_proto_init() }
func file_apppb_apppb_proto_init() {
	if File_apppb_apppb_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_apppb_apppb_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
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
		file_apppb_apppb_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SnapshotRequest); i {
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
		file_apppb_apppb_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Snapshot); i {
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
		file_apppb_apppb_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RestoreState); i {
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
		file_apppb_apppb_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*NewEpoch); i {
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
	file_apppb_apppb_proto_msgTypes[0].OneofWrappers = []interface{}{
		(*Event_SnapshotRequest)(nil),
		(*Event_Snapshot)(nil),
		(*Event_RestoreState)(nil),
		(*Event_NewEpoch)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_apppb_apppb_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   5,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_apppb_apppb_proto_goTypes,
		DependencyIndexes: file_apppb_apppb_proto_depIdxs,
		MessageInfos:      file_apppb_apppb_proto_msgTypes,
	}.Build()
	File_apppb_apppb_proto = out.File
	file_apppb_apppb_proto_rawDesc = nil
	file_apppb_apppb_proto_goTypes = nil
	file_apppb_apppb_proto_depIdxs = nil
}
