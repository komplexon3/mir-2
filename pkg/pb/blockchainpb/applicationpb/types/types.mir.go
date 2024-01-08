// Code generated by Mir codegen. DO NOT EDIT.

package applicationpbtypes

import (
	mirreflect "github.com/filecoin-project/mir/codegen/mirreflect"
	blockchainpb "github.com/filecoin-project/mir/pkg/pb/blockchainpb"
	applicationpb "github.com/filecoin-project/mir/pkg/pb/blockchainpb/applicationpb"
	payloadpb "github.com/filecoin-project/mir/pkg/pb/blockchainpb/payloadpb"
	statepb "github.com/filecoin-project/mir/pkg/pb/blockchainpb/statepb"
	reflectutil "github.com/filecoin-project/mir/pkg/util/reflectutil"
)

type Event struct {
	Type Event_Type
}

type Event_Type interface {
	mirreflect.GeneratedType
	isEvent_Type()
	Pb() applicationpb.Event_Type
}

type Event_TypeWrapper[T any] interface {
	Event_Type
	Unwrap() *T
}

func Event_TypeFromPb(pb applicationpb.Event_Type) Event_Type {
	if pb == nil {
		return nil
	}
	switch pb := pb.(type) {
	case *applicationpb.Event_NewHead:
		return &Event_NewHead{NewHead: NewHeadFromPb(pb.NewHead)}
	case *applicationpb.Event_VerifyBlockRequest:
		return &Event_VerifyBlockRequest{VerifyBlockRequest: VerifyBlockRequestFromPb(pb.VerifyBlockRequest)}
	case *applicationpb.Event_VerifyBlockResponse:
		return &Event_VerifyBlockResponse{VerifyBlockResponse: VerifyBlockResponseFromPb(pb.VerifyBlockResponse)}
	case *applicationpb.Event_PayloadRequest:
		return &Event_PayloadRequest{PayloadRequest: PayloadRequestFromPb(pb.PayloadRequest)}
	case *applicationpb.Event_PayloadResponse:
		return &Event_PayloadResponse{PayloadResponse: PayloadResponseFromPb(pb.PayloadResponse)}
	case *applicationpb.Event_ForkUpdate:
		return &Event_ForkUpdate{ForkUpdate: ForkUpdateFromPb(pb.ForkUpdate)}
	case *applicationpb.Event_MessageInput:
		return &Event_MessageInput{MessageInput: MessageInputFromPb(pb.MessageInput)}
	}
	return nil
}

type Event_NewHead struct {
	NewHead *NewHead
}

func (*Event_NewHead) isEvent_Type() {}

func (w *Event_NewHead) Unwrap() *NewHead {
	return w.NewHead
}

func (w *Event_NewHead) Pb() applicationpb.Event_Type {
	if w == nil {
		return nil
	}
	if w.NewHead == nil {
		return &applicationpb.Event_NewHead{}
	}
	return &applicationpb.Event_NewHead{NewHead: (w.NewHead).Pb()}
}

func (*Event_NewHead) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*applicationpb.Event_NewHead]()}
}

type Event_VerifyBlockRequest struct {
	VerifyBlockRequest *VerifyBlockRequest
}

func (*Event_VerifyBlockRequest) isEvent_Type() {}

func (w *Event_VerifyBlockRequest) Unwrap() *VerifyBlockRequest {
	return w.VerifyBlockRequest
}

func (w *Event_VerifyBlockRequest) Pb() applicationpb.Event_Type {
	if w == nil {
		return nil
	}
	if w.VerifyBlockRequest == nil {
		return &applicationpb.Event_VerifyBlockRequest{}
	}
	return &applicationpb.Event_VerifyBlockRequest{VerifyBlockRequest: (w.VerifyBlockRequest).Pb()}
}

func (*Event_VerifyBlockRequest) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*applicationpb.Event_VerifyBlockRequest]()}
}

type Event_VerifyBlockResponse struct {
	VerifyBlockResponse *VerifyBlockResponse
}

func (*Event_VerifyBlockResponse) isEvent_Type() {}

func (w *Event_VerifyBlockResponse) Unwrap() *VerifyBlockResponse {
	return w.VerifyBlockResponse
}

func (w *Event_VerifyBlockResponse) Pb() applicationpb.Event_Type {
	if w == nil {
		return nil
	}
	if w.VerifyBlockResponse == nil {
		return &applicationpb.Event_VerifyBlockResponse{}
	}
	return &applicationpb.Event_VerifyBlockResponse{VerifyBlockResponse: (w.VerifyBlockResponse).Pb()}
}

func (*Event_VerifyBlockResponse) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*applicationpb.Event_VerifyBlockResponse]()}
}

type Event_PayloadRequest struct {
	PayloadRequest *PayloadRequest
}

func (*Event_PayloadRequest) isEvent_Type() {}

func (w *Event_PayloadRequest) Unwrap() *PayloadRequest {
	return w.PayloadRequest
}

func (w *Event_PayloadRequest) Pb() applicationpb.Event_Type {
	if w == nil {
		return nil
	}
	if w.PayloadRequest == nil {
		return &applicationpb.Event_PayloadRequest{}
	}
	return &applicationpb.Event_PayloadRequest{PayloadRequest: (w.PayloadRequest).Pb()}
}

func (*Event_PayloadRequest) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*applicationpb.Event_PayloadRequest]()}
}

type Event_PayloadResponse struct {
	PayloadResponse *PayloadResponse
}

func (*Event_PayloadResponse) isEvent_Type() {}

func (w *Event_PayloadResponse) Unwrap() *PayloadResponse {
	return w.PayloadResponse
}

func (w *Event_PayloadResponse) Pb() applicationpb.Event_Type {
	if w == nil {
		return nil
	}
	if w.PayloadResponse == nil {
		return &applicationpb.Event_PayloadResponse{}
	}
	return &applicationpb.Event_PayloadResponse{PayloadResponse: (w.PayloadResponse).Pb()}
}

func (*Event_PayloadResponse) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*applicationpb.Event_PayloadResponse]()}
}

type Event_ForkUpdate struct {
	ForkUpdate *ForkUpdate
}

func (*Event_ForkUpdate) isEvent_Type() {}

func (w *Event_ForkUpdate) Unwrap() *ForkUpdate {
	return w.ForkUpdate
}

func (w *Event_ForkUpdate) Pb() applicationpb.Event_Type {
	if w == nil {
		return nil
	}
	if w.ForkUpdate == nil {
		return &applicationpb.Event_ForkUpdate{}
	}
	return &applicationpb.Event_ForkUpdate{ForkUpdate: (w.ForkUpdate).Pb()}
}

func (*Event_ForkUpdate) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*applicationpb.Event_ForkUpdate]()}
}

type Event_MessageInput struct {
	MessageInput *MessageInput
}

func (*Event_MessageInput) isEvent_Type() {}

func (w *Event_MessageInput) Unwrap() *MessageInput {
	return w.MessageInput
}

func (w *Event_MessageInput) Pb() applicationpb.Event_Type {
	if w == nil {
		return nil
	}
	if w.MessageInput == nil {
		return &applicationpb.Event_MessageInput{}
	}
	return &applicationpb.Event_MessageInput{MessageInput: (w.MessageInput).Pb()}
}

func (*Event_MessageInput) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*applicationpb.Event_MessageInput]()}
}

func EventFromPb(pb *applicationpb.Event) *Event {
	if pb == nil {
		return nil
	}
	return &Event{
		Type: Event_TypeFromPb(pb.Type),
	}
}

func (m *Event) Pb() *applicationpb.Event {
	if m == nil {
		return nil
	}
	pbMessage := &applicationpb.Event{}
	{
		if m.Type != nil {
			pbMessage.Type = (m.Type).Pb()
		}
	}

	return pbMessage
}

func (*Event) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*applicationpb.Event]()}
}

type NewHead struct {
	HeadId uint64
}

func NewHeadFromPb(pb *applicationpb.NewHead) *NewHead {
	if pb == nil {
		return nil
	}
	return &NewHead{
		HeadId: pb.HeadId,
	}
}

func (m *NewHead) Pb() *applicationpb.NewHead {
	if m == nil {
		return nil
	}
	pbMessage := &applicationpb.NewHead{}
	{
		pbMessage.HeadId = m.HeadId
	}

	return pbMessage
}

func (*NewHead) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*applicationpb.NewHead]()}
}

type VerifyBlockRequest struct {
	RequestId uint64
	Block     *blockchainpb.Block
}

func VerifyBlockRequestFromPb(pb *applicationpb.VerifyBlockRequest) *VerifyBlockRequest {
	if pb == nil {
		return nil
	}
	return &VerifyBlockRequest{
		RequestId: pb.RequestId,
		Block:     pb.Block,
	}
}

func (m *VerifyBlockRequest) Pb() *applicationpb.VerifyBlockRequest {
	if m == nil {
		return nil
	}
	pbMessage := &applicationpb.VerifyBlockRequest{}
	{
		pbMessage.RequestId = m.RequestId
		if m.Block != nil {
			pbMessage.Block = m.Block
		}
	}

	return pbMessage
}

func (*VerifyBlockRequest) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*applicationpb.VerifyBlockRequest]()}
}

type VerifyBlockResponse struct {
	RequestId uint64
	Ok        bool
}

func VerifyBlockResponseFromPb(pb *applicationpb.VerifyBlockResponse) *VerifyBlockResponse {
	if pb == nil {
		return nil
	}
	return &VerifyBlockResponse{
		RequestId: pb.RequestId,
		Ok:        pb.Ok,
	}
}

func (m *VerifyBlockResponse) Pb() *applicationpb.VerifyBlockResponse {
	if m == nil {
		return nil
	}
	pbMessage := &applicationpb.VerifyBlockResponse{}
	{
		pbMessage.RequestId = m.RequestId
		pbMessage.Ok = m.Ok
	}

	return pbMessage
}

func (*VerifyBlockResponse) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*applicationpb.VerifyBlockResponse]()}
}

type ForkUpdate struct {
	RemovedChain *blockchainpb.Blockchain
	AddedChain   *blockchainpb.Blockchain
	ForkState    *statepb.State
}

func ForkUpdateFromPb(pb *applicationpb.ForkUpdate) *ForkUpdate {
	if pb == nil {
		return nil
	}
	return &ForkUpdate{
		RemovedChain: pb.RemovedChain,
		AddedChain:   pb.AddedChain,
		ForkState:    pb.ForkState,
	}
}

func (m *ForkUpdate) Pb() *applicationpb.ForkUpdate {
	if m == nil {
		return nil
	}
	pbMessage := &applicationpb.ForkUpdate{}
	{
		if m.RemovedChain != nil {
			pbMessage.RemovedChain = m.RemovedChain
		}
		if m.AddedChain != nil {
			pbMessage.AddedChain = m.AddedChain
		}
		if m.ForkState != nil {
			pbMessage.ForkState = m.ForkState
		}
	}

	return pbMessage
}

func (*ForkUpdate) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*applicationpb.ForkUpdate]()}
}

type PayloadRequest struct {
	HeadId uint64
}

func PayloadRequestFromPb(pb *applicationpb.PayloadRequest) *PayloadRequest {
	if pb == nil {
		return nil
	}
	return &PayloadRequest{
		HeadId: pb.HeadId,
	}
}

func (m *PayloadRequest) Pb() *applicationpb.PayloadRequest {
	if m == nil {
		return nil
	}
	pbMessage := &applicationpb.PayloadRequest{}
	{
		pbMessage.HeadId = m.HeadId
	}

	return pbMessage
}

func (*PayloadRequest) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*applicationpb.PayloadRequest]()}
}

type PayloadResponse struct {
	HeadId  uint64
	Payload *payloadpb.Payload
}

func PayloadResponseFromPb(pb *applicationpb.PayloadResponse) *PayloadResponse {
	if pb == nil {
		return nil
	}
	return &PayloadResponse{
		HeadId:  pb.HeadId,
		Payload: pb.Payload,
	}
}

func (m *PayloadResponse) Pb() *applicationpb.PayloadResponse {
	if m == nil {
		return nil
	}
	pbMessage := &applicationpb.PayloadResponse{}
	{
		pbMessage.HeadId = m.HeadId
		if m.Payload != nil {
			pbMessage.Payload = m.Payload
		}
	}

	return pbMessage
}

func (*PayloadResponse) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*applicationpb.PayloadResponse]()}
}

type MessageInput struct {
	Text string
}

func MessageInputFromPb(pb *applicationpb.MessageInput) *MessageInput {
	if pb == nil {
		return nil
	}
	return &MessageInput{
		Text: pb.Text,
	}
}

func (m *MessageInput) Pb() *applicationpb.MessageInput {
	if m == nil {
		return nil
	}
	pbMessage := &applicationpb.MessageInput{}
	{
		pbMessage.Text = m.Text
	}

	return pbMessage
}

func (*MessageInput) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*applicationpb.MessageInput]()}
}
