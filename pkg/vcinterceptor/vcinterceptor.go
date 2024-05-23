package vcinterceptor

import (
	"sync"

	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	"github.com/filecoin-project/mir/pkg/pb/messagepb"
	messagepbtypes "github.com/filecoin-project/mir/pkg/pb/messagepb/types"
	transportpbevents "github.com/filecoin-project/mir/pkg/pb/transportpb/events"
	transportpbtypes "github.com/filecoin-project/mir/pkg/pb/transportpb/types"
	"github.com/filecoin-project/mir/pkg/vcinterceptor/messages"
	"github.com/filecoin-project/mir/pkg/vcinterceptor/vectorclock"
	"github.com/filecoin-project/mir/stdevents"
	"github.com/filecoin-project/mir/stdtypes"
	es "github.com/go-errors/errors"
)

// TODO: instrument each event with current (incremented) vc
// for every incoming message, unwrap and combine with current vc (and increment?)
// for outgoing message, wrap and add current vc (incremented?)

const vcKey = "vc"

type VectorClockInterceptor struct {
	nodeID  stdtypes.NodeID
	vc      vectorclock.VectorClock
	vcMutex sync.RWMutex
}

func New(nodeId stdtypes.NodeID) *VectorClockInterceptor {
	return &VectorClockInterceptor{
		nodeID:  nodeId,
		vc:      *vectorclock.NewVectorClock(),
		vcMutex: sync.RWMutex{},
	}
}

func (vci *VectorClockInterceptor) Intercept(events *stdtypes.EventList) (*stdtypes.EventList, error) {
	newEvents := stdtypes.EmptyList()
	iter := events.Iterator()
	for ev := iter.Next(); ev != nil; ev = iter.Next() {
		var nEv stdtypes.Event
		var err error

		// TODO: proto msgs?
		switch evT := ev.(type) {
		case *stdevents.SendMessage:
			vc := func() *vectorclock.VectorClock {
				vci.vcMutex.Lock()
				defer vci.vcMutex.Unlock()
				vci.vc.Increment(vci.nodeID)
				return vci.vc.Clone()
			}()
			payload := evT.Payload
			wrappedPayload := messages.WrapNativeMessage(payload, *vc)
			nEv = stdevents.NewSendMessageWithSrc(evT.SrcModule, evT.DestModule, evT.RemoteDestModule, wrappedPayload, evT.DestNodes...)
			nEv, err = nEv.SetMetadata(vcKey, vc)
			if err != nil {
				return nil, err
			}
		case *stdevents.MessageReceived:
			var vcMsg vectorclock.VectorClock
			rawWrappedPayload, err := evT.Payload.ToBytes()
			if err != nil {
				return nil, err
			}
			vcm, err := messages.Deserialize(rawWrappedPayload)
			if err != nil {
				return nil, err
			}

			switch vcmT := vcm.(type) {
			case *messages.VCNativeMessage:
				var payload stdtypes.Message
				payload, vcMsg = vcmT.UnwrapMessage()
				nEv = stdevents.NewMessageReceivedWithSrc(evT.SrcModule, evT.DestModule, evT.Sender, payload)
			case *messages.VCPbMessage:
				var payload *messagepb.Message
				payload, vcMsg = vcmT.UnwrapMessage()
				nEv = transportpbevents.MessageReceived(evT.DestModule, evT.Sender, messagepbtypes.MessageFromPb(payload)).Pb()
			default:
				panic("we shouldn't get here")
			}

			vc := func() *vectorclock.VectorClock {
				vci.vcMutex.Lock()
				defer vci.vcMutex.Unlock()
				vci.vc.CombineAndIncrement(&vcMsg, vci.nodeID)
				return vci.vc.Clone()
			}()
			nEv, err = nEv.SetMetadata(vcKey, vc)
		case *eventpb.Event:
			switch pbEvT := evT.Type.(type) {
			case *eventpb.Event_Transport:
				switch pbTEvT := transportpbtypes.EventFromPb(pbEvT.Transport).Type.(type) {
				case *transportpbtypes.Event_SendMessage:
					vc := func() *vectorclock.VectorClock {
						vci.vcMutex.Lock()
						defer vci.vcMutex.Unlock()
						vci.vc.Increment(vci.nodeID)
						return vci.vc.Clone()
					}()
					payload := pbTEvT.SendMessage.Msg.Pb()
					wrappedPayload := messages.WrapPbMessage(payload, *vc)
					nEv = stdevents.NewSendMessageWithSrc(evT.Src(), evT.Dest(), pbTEvT.SendMessage.Msg.DestModule, wrappedPayload, pbTEvT.SendMessage.Destinations...)
					nEv, err = nEv.SetMetadata(vcKey, vc)
					if err != nil {
						return nil, err
					}
				case *transportpbtypes.Event_MessageReceived:
					return nil, es.Errorf("unexpected message received event, all pb send events should be wrapped vector message: %T", pbTEvT)
				default:
					return nil, es.Errorf("unexpected type of transport event: %T", pbTEvT)
				}
			default:
				vc := func() *vectorclock.VectorClock {
					vci.vcMutex.Lock()
					defer vci.vcMutex.Unlock()
					vci.vc.Increment(vci.nodeID)
					return vci.vc.Clone()
				}()
				nEv, err = ev.SetMetadata(vcKey, vc)
				if err != nil {
					return nil, err
				}
			}
		default:
			vc := func() *vectorclock.VectorClock {
				vci.vcMutex.Lock()
				defer vci.vcMutex.Unlock()
				vci.vc.Increment(vci.nodeID)
				return vci.vc.Clone()
			}()
			nEv, err = ev.SetMetadata(vcKey, vc)
			if err != nil {
				return nil, err
			}
		}

		newEvents.PushBack(nEv)
	}
	return newEvents, nil
}
