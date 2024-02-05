// Code generated by Mir codegen. DO NOT EDIT.

package eventpbtypes

import (
	mirreflect "github.com/filecoin-project/mir/codegen/mirreflect"
	types11 "github.com/filecoin-project/mir/pkg/pb/apppb/types"
	types4 "github.com/filecoin-project/mir/pkg/pb/availabilitypb/batchdbpb/types"
	types3 "github.com/filecoin-project/mir/pkg/pb/availabilitypb/types"
	types5 "github.com/filecoin-project/mir/pkg/pb/batchfetcherpb/types"
	types1 "github.com/filecoin-project/mir/pkg/pb/bcbpb/types"
	types21 "github.com/filecoin-project/mir/pkg/pb/blockchainpb/applicationpb/types"
	types17 "github.com/filecoin-project/mir/pkg/pb/blockchainpb/bcmpb/types"
	types19 "github.com/filecoin-project/mir/pkg/pb/blockchainpb/broadcastpb/types"
	types22 "github.com/filecoin-project/mir/pkg/pb/blockchainpb/interceptorpb/types"
	types18 "github.com/filecoin-project/mir/pkg/pb/blockchainpb/minerpb/types"
	types20 "github.com/filecoin-project/mir/pkg/pb/blockchainpb/synchronizerpb/types"
	types13 "github.com/filecoin-project/mir/pkg/pb/checkpointpb/chkpvalidatorpb/types"
	types7 "github.com/filecoin-project/mir/pkg/pb/checkpointpb/types"
	types10 "github.com/filecoin-project/mir/pkg/pb/cryptopb/types"
	eventpb "github.com/filecoin-project/mir/pkg/pb/eventpb"
	types "github.com/filecoin-project/mir/pkg/pb/hasherpb/types"
	types8 "github.com/filecoin-project/mir/pkg/pb/isspb/types"
	types2 "github.com/filecoin-project/mir/pkg/pb/mempoolpb/types"
	types14 "github.com/filecoin-project/mir/pkg/pb/ordererpb/pprepvalidatorpb/types"
	types9 "github.com/filecoin-project/mir/pkg/pb/ordererpb/types"
	types15 "github.com/filecoin-project/mir/pkg/pb/pingpongpb/types"
	types16 "github.com/filecoin-project/mir/pkg/pb/testerpb/types"
	types6 "github.com/filecoin-project/mir/pkg/pb/threshcryptopb/types"
	types12 "github.com/filecoin-project/mir/pkg/pb/transportpb/types"
	reflectutil "github.com/filecoin-project/mir/pkg/util/reflectutil"
	stdtypes "github.com/filecoin-project/mir/stdtypes"
	wrapperspb "google.golang.org/protobuf/types/known/wrapperspb"
)

type Event struct {
	DestModule stdtypes.ModuleID
	Type       Event_Type
}

type Event_Type interface {
	mirreflect.GeneratedType
	isEvent_Type()
	Pb() eventpb.Event_Type
}

type Event_TypeWrapper[T any] interface {
	Event_Type
	Unwrap() *T
}

func Event_TypeFromPb(pb eventpb.Event_Type) Event_Type {
	if pb == nil {
		return nil
	}
	switch pb := pb.(type) {
	case *eventpb.Event_Hasher:
		return &Event_Hasher{Hasher: types.EventFromPb(pb.Hasher)}
	case *eventpb.Event_Bcb:
		return &Event_Bcb{Bcb: types1.EventFromPb(pb.Bcb)}
	case *eventpb.Event_Mempool:
		return &Event_Mempool{Mempool: types2.EventFromPb(pb.Mempool)}
	case *eventpb.Event_Availability:
		return &Event_Availability{Availability: types3.EventFromPb(pb.Availability)}
	case *eventpb.Event_BatchDb:
		return &Event_BatchDb{BatchDb: types4.EventFromPb(pb.BatchDb)}
	case *eventpb.Event_BatchFetcher:
		return &Event_BatchFetcher{BatchFetcher: types5.EventFromPb(pb.BatchFetcher)}
	case *eventpb.Event_ThreshCrypto:
		return &Event_ThreshCrypto{ThreshCrypto: types6.EventFromPb(pb.ThreshCrypto)}
	case *eventpb.Event_Checkpoint:
		return &Event_Checkpoint{Checkpoint: types7.EventFromPb(pb.Checkpoint)}
	case *eventpb.Event_Iss:
		return &Event_Iss{Iss: types8.EventFromPb(pb.Iss)}
	case *eventpb.Event_Orderer:
		return &Event_Orderer{Orderer: types9.EventFromPb(pb.Orderer)}
	case *eventpb.Event_Crypto:
		return &Event_Crypto{Crypto: types10.EventFromPb(pb.Crypto)}
	case *eventpb.Event_App:
		return &Event_App{App: types11.EventFromPb(pb.App)}
	case *eventpb.Event_Transport:
		return &Event_Transport{Transport: types12.EventFromPb(pb.Transport)}
	case *eventpb.Event_ChkpValidator:
		return &Event_ChkpValidator{ChkpValidator: types13.EventFromPb(pb.ChkpValidator)}
	case *eventpb.Event_PprepValiadtor:
		return &Event_PprepValiadtor{PprepValiadtor: types14.EventFromPb(pb.PprepValiadtor)}
	case *eventpb.Event_Serialized:
		return &Event_Serialized{Serialized: pb.Serialized}
	case *eventpb.Event_PingPong:
		return &Event_PingPong{PingPong: types15.EventFromPb(pb.PingPong)}
	case *eventpb.Event_TestingString:
		return &Event_TestingString{TestingString: pb.TestingString}
	case *eventpb.Event_TestingUint:
		return &Event_TestingUint{TestingUint: pb.TestingUint}
	case *eventpb.Event_Tester:
		return &Event_Tester{Tester: types16.TesterFromPb(pb.Tester)}
	case *eventpb.Event_Bcm:
		return &Event_Bcm{Bcm: types17.EventFromPb(pb.Bcm)}
	case *eventpb.Event_Miner:
		return &Event_Miner{Miner: types18.EventFromPb(pb.Miner)}
	case *eventpb.Event_Broadcast:
		return &Event_Broadcast{Broadcast: types19.EventFromPb(pb.Broadcast)}
	case *eventpb.Event_Synchronizer:
		return &Event_Synchronizer{Synchronizer: types20.EventFromPb(pb.Synchronizer)}
	case *eventpb.Event_Application:
		return &Event_Application{Application: types21.EventFromPb(pb.Application)}
	case *eventpb.Event_Bcinterceptor:
		return &Event_Bcinterceptor{Bcinterceptor: types22.EventFromPb(pb.Bcinterceptor)}
	}
	return nil
}

type Event_Hasher struct {
	Hasher *types.Event
}

func (*Event_Hasher) isEvent_Type() {}

func (w *Event_Hasher) Unwrap() *types.Event {
	return w.Hasher
}

func (w *Event_Hasher) Pb() eventpb.Event_Type {
	if w == nil {
		return nil
	}
	if w.Hasher == nil {
		return &eventpb.Event_Hasher{}
	}
	return &eventpb.Event_Hasher{Hasher: (w.Hasher).Pb()}
}

func (*Event_Hasher) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.Event_Hasher]()}
}

type Event_Bcb struct {
	Bcb *types1.Event
}

func (*Event_Bcb) isEvent_Type() {}

func (w *Event_Bcb) Unwrap() *types1.Event {
	return w.Bcb
}

func (w *Event_Bcb) Pb() eventpb.Event_Type {
	if w == nil {
		return nil
	}
	if w.Bcb == nil {
		return &eventpb.Event_Bcb{}
	}
	return &eventpb.Event_Bcb{Bcb: (w.Bcb).Pb()}
}

func (*Event_Bcb) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.Event_Bcb]()}
}

type Event_Mempool struct {
	Mempool *types2.Event
}

func (*Event_Mempool) isEvent_Type() {}

func (w *Event_Mempool) Unwrap() *types2.Event {
	return w.Mempool
}

func (w *Event_Mempool) Pb() eventpb.Event_Type {
	if w == nil {
		return nil
	}
	if w.Mempool == nil {
		return &eventpb.Event_Mempool{}
	}
	return &eventpb.Event_Mempool{Mempool: (w.Mempool).Pb()}
}

func (*Event_Mempool) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.Event_Mempool]()}
}

type Event_Availability struct {
	Availability *types3.Event
}

func (*Event_Availability) isEvent_Type() {}

func (w *Event_Availability) Unwrap() *types3.Event {
	return w.Availability
}

func (w *Event_Availability) Pb() eventpb.Event_Type {
	if w == nil {
		return nil
	}
	if w.Availability == nil {
		return &eventpb.Event_Availability{}
	}
	return &eventpb.Event_Availability{Availability: (w.Availability).Pb()}
}

func (*Event_Availability) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.Event_Availability]()}
}

type Event_BatchDb struct {
	BatchDb *types4.Event
}

func (*Event_BatchDb) isEvent_Type() {}

func (w *Event_BatchDb) Unwrap() *types4.Event {
	return w.BatchDb
}

func (w *Event_BatchDb) Pb() eventpb.Event_Type {
	if w == nil {
		return nil
	}
	if w.BatchDb == nil {
		return &eventpb.Event_BatchDb{}
	}
	return &eventpb.Event_BatchDb{BatchDb: (w.BatchDb).Pb()}
}

func (*Event_BatchDb) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.Event_BatchDb]()}
}

type Event_BatchFetcher struct {
	BatchFetcher *types5.Event
}

func (*Event_BatchFetcher) isEvent_Type() {}

func (w *Event_BatchFetcher) Unwrap() *types5.Event {
	return w.BatchFetcher
}

func (w *Event_BatchFetcher) Pb() eventpb.Event_Type {
	if w == nil {
		return nil
	}
	if w.BatchFetcher == nil {
		return &eventpb.Event_BatchFetcher{}
	}
	return &eventpb.Event_BatchFetcher{BatchFetcher: (w.BatchFetcher).Pb()}
}

func (*Event_BatchFetcher) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.Event_BatchFetcher]()}
}

type Event_ThreshCrypto struct {
	ThreshCrypto *types6.Event
}

func (*Event_ThreshCrypto) isEvent_Type() {}

func (w *Event_ThreshCrypto) Unwrap() *types6.Event {
	return w.ThreshCrypto
}

func (w *Event_ThreshCrypto) Pb() eventpb.Event_Type {
	if w == nil {
		return nil
	}
	if w.ThreshCrypto == nil {
		return &eventpb.Event_ThreshCrypto{}
	}
	return &eventpb.Event_ThreshCrypto{ThreshCrypto: (w.ThreshCrypto).Pb()}
}

func (*Event_ThreshCrypto) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.Event_ThreshCrypto]()}
}

type Event_Checkpoint struct {
	Checkpoint *types7.Event
}

func (*Event_Checkpoint) isEvent_Type() {}

func (w *Event_Checkpoint) Unwrap() *types7.Event {
	return w.Checkpoint
}

func (w *Event_Checkpoint) Pb() eventpb.Event_Type {
	if w == nil {
		return nil
	}
	if w.Checkpoint == nil {
		return &eventpb.Event_Checkpoint{}
	}
	return &eventpb.Event_Checkpoint{Checkpoint: (w.Checkpoint).Pb()}
}

func (*Event_Checkpoint) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.Event_Checkpoint]()}
}

type Event_Iss struct {
	Iss *types8.Event
}

func (*Event_Iss) isEvent_Type() {}

func (w *Event_Iss) Unwrap() *types8.Event {
	return w.Iss
}

func (w *Event_Iss) Pb() eventpb.Event_Type {
	if w == nil {
		return nil
	}
	if w.Iss == nil {
		return &eventpb.Event_Iss{}
	}
	return &eventpb.Event_Iss{Iss: (w.Iss).Pb()}
}

func (*Event_Iss) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.Event_Iss]()}
}

type Event_Orderer struct {
	Orderer *types9.Event
}

func (*Event_Orderer) isEvent_Type() {}

func (w *Event_Orderer) Unwrap() *types9.Event {
	return w.Orderer
}

func (w *Event_Orderer) Pb() eventpb.Event_Type {
	if w == nil {
		return nil
	}
	if w.Orderer == nil {
		return &eventpb.Event_Orderer{}
	}
	return &eventpb.Event_Orderer{Orderer: (w.Orderer).Pb()}
}

func (*Event_Orderer) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.Event_Orderer]()}
}

type Event_Crypto struct {
	Crypto *types10.Event
}

func (*Event_Crypto) isEvent_Type() {}

func (w *Event_Crypto) Unwrap() *types10.Event {
	return w.Crypto
}

func (w *Event_Crypto) Pb() eventpb.Event_Type {
	if w == nil {
		return nil
	}
	if w.Crypto == nil {
		return &eventpb.Event_Crypto{}
	}
	return &eventpb.Event_Crypto{Crypto: (w.Crypto).Pb()}
}

func (*Event_Crypto) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.Event_Crypto]()}
}

type Event_App struct {
	App *types11.Event
}

func (*Event_App) isEvent_Type() {}

func (w *Event_App) Unwrap() *types11.Event {
	return w.App
}

func (w *Event_App) Pb() eventpb.Event_Type {
	if w == nil {
		return nil
	}
	if w.App == nil {
		return &eventpb.Event_App{}
	}
	return &eventpb.Event_App{App: (w.App).Pb()}
}

func (*Event_App) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.Event_App]()}
}

type Event_Transport struct {
	Transport *types12.Event
}

func (*Event_Transport) isEvent_Type() {}

func (w *Event_Transport) Unwrap() *types12.Event {
	return w.Transport
}

func (w *Event_Transport) Pb() eventpb.Event_Type {
	if w == nil {
		return nil
	}
	if w.Transport == nil {
		return &eventpb.Event_Transport{}
	}
	return &eventpb.Event_Transport{Transport: (w.Transport).Pb()}
}

func (*Event_Transport) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.Event_Transport]()}
}

type Event_ChkpValidator struct {
	ChkpValidator *types13.Event
}

func (*Event_ChkpValidator) isEvent_Type() {}

func (w *Event_ChkpValidator) Unwrap() *types13.Event {
	return w.ChkpValidator
}

func (w *Event_ChkpValidator) Pb() eventpb.Event_Type {
	if w == nil {
		return nil
	}
	if w.ChkpValidator == nil {
		return &eventpb.Event_ChkpValidator{}
	}
	return &eventpb.Event_ChkpValidator{ChkpValidator: (w.ChkpValidator).Pb()}
}

func (*Event_ChkpValidator) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.Event_ChkpValidator]()}
}

type Event_PprepValiadtor struct {
	PprepValiadtor *types14.Event
}

func (*Event_PprepValiadtor) isEvent_Type() {}

func (w *Event_PprepValiadtor) Unwrap() *types14.Event {
	return w.PprepValiadtor
}

func (w *Event_PprepValiadtor) Pb() eventpb.Event_Type {
	if w == nil {
		return nil
	}
	if w.PprepValiadtor == nil {
		return &eventpb.Event_PprepValiadtor{}
	}
	return &eventpb.Event_PprepValiadtor{PprepValiadtor: (w.PprepValiadtor).Pb()}
}

func (*Event_PprepValiadtor) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.Event_PprepValiadtor]()}
}

type Event_Serialized struct {
	Serialized *eventpb.SerializedEvent
}

func (*Event_Serialized) isEvent_Type() {}

func (w *Event_Serialized) Unwrap() *eventpb.SerializedEvent {
	return w.Serialized
}

func (w *Event_Serialized) Pb() eventpb.Event_Type {
	if w == nil {
		return nil
	}
	if w.Serialized == nil {
		return &eventpb.Event_Serialized{}
	}
	return &eventpb.Event_Serialized{Serialized: w.Serialized}
}

func (*Event_Serialized) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.Event_Serialized]()}
}

type Event_PingPong struct {
	PingPong *types15.Event
}

func (*Event_PingPong) isEvent_Type() {}

func (w *Event_PingPong) Unwrap() *types15.Event {
	return w.PingPong
}

func (w *Event_PingPong) Pb() eventpb.Event_Type {
	if w == nil {
		return nil
	}
	if w.PingPong == nil {
		return &eventpb.Event_PingPong{}
	}
	return &eventpb.Event_PingPong{PingPong: (w.PingPong).Pb()}
}

func (*Event_PingPong) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.Event_PingPong]()}
}

type Event_TestingString struct {
	TestingString *wrapperspb.StringValue
}

func (*Event_TestingString) isEvent_Type() {}

func (w *Event_TestingString) Unwrap() *wrapperspb.StringValue {
	return w.TestingString
}

func (w *Event_TestingString) Pb() eventpb.Event_Type {
	if w == nil {
		return nil
	}
	if w.TestingString == nil {
		return &eventpb.Event_TestingString{}
	}
	return &eventpb.Event_TestingString{TestingString: w.TestingString}
}

func (*Event_TestingString) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.Event_TestingString]()}
}

type Event_TestingUint struct {
	TestingUint *wrapperspb.UInt64Value
}

func (*Event_TestingUint) isEvent_Type() {}

func (w *Event_TestingUint) Unwrap() *wrapperspb.UInt64Value {
	return w.TestingUint
}

func (w *Event_TestingUint) Pb() eventpb.Event_Type {
	if w == nil {
		return nil
	}
	if w.TestingUint == nil {
		return &eventpb.Event_TestingUint{}
	}
	return &eventpb.Event_TestingUint{TestingUint: w.TestingUint}
}

func (*Event_TestingUint) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.Event_TestingUint]()}
}

type Event_Tester struct {
	Tester *types16.Tester
}

func (*Event_Tester) isEvent_Type() {}

func (w *Event_Tester) Unwrap() *types16.Tester {
	return w.Tester
}

func (w *Event_Tester) Pb() eventpb.Event_Type {
	if w == nil {
		return nil
	}
	if w.Tester == nil {
		return &eventpb.Event_Tester{}
	}
	return &eventpb.Event_Tester{Tester: (w.Tester).Pb()}
}

func (*Event_Tester) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.Event_Tester]()}
}

type Event_Bcm struct {
	Bcm *types17.Event
}

func (*Event_Bcm) isEvent_Type() {}

func (w *Event_Bcm) Unwrap() *types17.Event {
	return w.Bcm
}

func (w *Event_Bcm) Pb() eventpb.Event_Type {
	if w == nil {
		return nil
	}
	if w.Bcm == nil {
		return &eventpb.Event_Bcm{}
	}
	return &eventpb.Event_Bcm{Bcm: (w.Bcm).Pb()}
}

func (*Event_Bcm) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.Event_Bcm]()}
}

type Event_Miner struct {
	Miner *types18.Event
}

func (*Event_Miner) isEvent_Type() {}

func (w *Event_Miner) Unwrap() *types18.Event {
	return w.Miner
}

func (w *Event_Miner) Pb() eventpb.Event_Type {
	if w == nil {
		return nil
	}
	if w.Miner == nil {
		return &eventpb.Event_Miner{}
	}
	return &eventpb.Event_Miner{Miner: (w.Miner).Pb()}
}

func (*Event_Miner) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.Event_Miner]()}
}

type Event_Broadcast struct {
	Broadcast *types19.Event
}

func (*Event_Broadcast) isEvent_Type() {}

func (w *Event_Broadcast) Unwrap() *types19.Event {
	return w.Broadcast
}

func (w *Event_Broadcast) Pb() eventpb.Event_Type {
	if w == nil {
		return nil
	}
	if w.Broadcast == nil {
		return &eventpb.Event_Broadcast{}
	}
	return &eventpb.Event_Broadcast{Broadcast: (w.Broadcast).Pb()}
}

func (*Event_Broadcast) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.Event_Broadcast]()}
}

type Event_Synchronizer struct {
	Synchronizer *types20.Event
}

func (*Event_Synchronizer) isEvent_Type() {}

func (w *Event_Synchronizer) Unwrap() *types20.Event {
	return w.Synchronizer
}

func (w *Event_Synchronizer) Pb() eventpb.Event_Type {
	if w == nil {
		return nil
	}
	if w.Synchronizer == nil {
		return &eventpb.Event_Synchronizer{}
	}
	return &eventpb.Event_Synchronizer{Synchronizer: (w.Synchronizer).Pb()}
}

func (*Event_Synchronizer) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.Event_Synchronizer]()}
}

type Event_Application struct {
	Application *types21.Event
}

func (*Event_Application) isEvent_Type() {}

func (w *Event_Application) Unwrap() *types21.Event {
	return w.Application
}

func (w *Event_Application) Pb() eventpb.Event_Type {
	if w == nil {
		return nil
	}
	if w.Application == nil {
		return &eventpb.Event_Application{}
	}
	return &eventpb.Event_Application{Application: (w.Application).Pb()}
}

func (*Event_Application) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.Event_Application]()}
}

type Event_Bcinterceptor struct {
	Bcinterceptor *types22.Event
}

func (*Event_Bcinterceptor) isEvent_Type() {}

func (w *Event_Bcinterceptor) Unwrap() *types22.Event {
	return w.Bcinterceptor
}

func (w *Event_Bcinterceptor) Pb() eventpb.Event_Type {
	if w == nil {
		return nil
	}
	if w.Bcinterceptor == nil {
		return &eventpb.Event_Bcinterceptor{}
	}
	return &eventpb.Event_Bcinterceptor{Bcinterceptor: (w.Bcinterceptor).Pb()}
}

func (*Event_Bcinterceptor) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.Event_Bcinterceptor]()}
}

func EventFromPb(pb *eventpb.Event) *Event {
	if pb == nil {
		return nil
	}
	return &Event{
		DestModule: (stdtypes.ModuleID)(pb.DestModule),
		Type:       Event_TypeFromPb(pb.Type),
	}
}

func (m *Event) Pb() *eventpb.Event {
	if m == nil {
		return nil
	}
	pbMessage := &eventpb.Event{}
	{
		pbMessage.DestModule = (string)(m.DestModule)
		if m.Type != nil {
			pbMessage.Type = (m.Type).Pb()
		}
	}

	return pbMessage
}

func (*Event) MirReflect() mirreflect.Type {
	return mirreflect.TypeImpl{PbType_: reflectutil.TypeOf[*eventpb.Event]()}
}
