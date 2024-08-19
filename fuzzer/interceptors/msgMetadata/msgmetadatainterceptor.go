package msgmetadata

import (
	"github.com/filecoin-project/mir/fuzzer/interceptors/msgMetadata/messages"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	messagepbtypes "github.com/filecoin-project/mir/pkg/pb/messagepb/types"
	transportpbevents "github.com/filecoin-project/mir/pkg/pb/transportpb/events"
	transportpbtypes "github.com/filecoin-project/mir/pkg/pb/transportpb/types"
	"github.com/filecoin-project/mir/stdevents"
	"github.com/filecoin-project/mir/stdtypes"
	es "github.com/go-errors/errors"
)

type msgMetadataInterceptor struct {
	logger logging.Logger
	keys   []string
}

func newMsgMetadataInterceptor(logger logging.Logger, keys []string) msgMetadataInterceptor {
	logger = logging.Decorate(logger, "Message Metadata Interceptor - ")
	return msgMetadataInterceptor{
		keys:   keys,
		logger: logger,
	}
}

func NewMsgMetadataInterceptorPair(logger logging.Logger, keys ...string) (*MsgMetadataInterceptorIn, *MsgMetadataInterceptorOut) {
	mi := newMsgMetadataInterceptor(logger, keys)
	return &MsgMetadataInterceptorIn{mi}, &MsgMetadataInterceptorOut{mi}
}

type MsgMetadataInterceptorIn struct {
	msgMetadataInterceptor
}

func (mi *MsgMetadataInterceptorIn) Intercept(events *stdtypes.EventList) (*stdtypes.EventList, error) {
	mi.logger.Log(logging.LevelTrace, "Intercepting IN", "events", events)
	newEvents := stdtypes.EmptyList()
	iter := events.Iterator()
	for ev := iter.Next(); ev != nil; ev = iter.Next() {
		var nEv stdtypes.Event
		switch evT := ev.(type) {
		case *stdevents.MessageReceived:
			rawWrappedPayload, err := evT.Payload.ToBytes()
			if err != nil {
				mi.logger.Log(logging.LevelError, "to bytes", "error", err)
				return nil, err
			}
			msg, err := messages.Deserialize(rawWrappedPayload)
			if err != nil {
				mi.logger.Log(logging.LevelError, "deserialize", "error", err)
				return nil, err
			}

			switch msgT := msg.(type) {
			case *messages.MetadataNativeMessage:
				payload := msgT.GetMessage()
				nEv = stdevents.NewMessageReceivedWithSrc(evT.SrcModule, evT.DestModule, evT.Sender, payload)
			case *messages.MetadataPbMessage:
				payload := msgT.GetMessage()
				nEv = transportpbevents.MessageReceived(evT.DestModule, evT.Sender, messagepbtypes.MessageFromPb(payload)).Pb()
			default:
				panic("we shouldn't get here")
			}

			for _, key := range mi.keys {
				msgMetadataVal, err := msg.GetMetadata(key)
				if err != nil {
					continue
					// TODO: add error to all GetMetadata - then use this handling
					// if err == stdtypes.KeyNotFoundErr {
					// 	continue
					// }
					// return nil, err
				}
				nEv, err = nEv.SetMetadata(key, msgMetadataVal)
				if err != nil {
					return nil, err
				}
			}
		case *eventpb.Event:
			switch pbEvT := evT.Type.(type) {
			case *eventpb.Event_Transport:
				switch pbTEvT := transportpbtypes.EventFromPb(pbEvT.Transport).Type.(type) {
				case *transportpbtypes.Event_MessageReceived:
					return nil, es.Errorf("unexpected message received event, all pb send events should be wrapped vector message: %T", pbTEvT)
				default:
					return nil, es.Errorf("unexpected type of transport event: %T", pbTEvT)
				}
			default:
				nEv = ev
			}
		default:
			nEv = ev
		}

		newEvents.PushBack(nEv)
	}

	mi.logger.Log(logging.LevelTrace, "Intercepting IN done", "events", newEvents)
	return newEvents, nil
}

type MsgMetadataInterceptorOut struct {
	msgMetadataInterceptor
}

func (mi *MsgMetadataInterceptorOut) Intercept(events *stdtypes.EventList) (*stdtypes.EventList, error) {
	mi.logger.Log(logging.LevelTrace, "Intercepting OUT", "events", events)
	newEvents := stdtypes.EmptyList()
	iter := events.Iterator()
	for ev := iter.Next(); ev != nil; ev = iter.Next() {
		var nEv stdtypes.Event
		msgMetadata := make(map[string]interface{})
		for _, key := range mi.keys {
			metadataVal, err := ev.GetMetadata(key)
			if err != nil {
				continue
				// TODO: add error to all GetMetadata - then use this handling
				// if err == stdtypes.KeyNotFoundErr {
				// 	continue
				// }
				// return nil, err
			}
			msgMetadata[key] = metadataVal
		}

		switch evT := ev.(type) {
		case *stdevents.SendMessage:
			wrappedPayload := messages.WrapNativeMessage(evT.Payload, msgMetadata)
			nEv = stdevents.NewSendMessageWithSrc(evT.SrcModule, evT.DestModule, evT.RemoteDestModule, wrappedPayload, evT.DestNodes...)
		case *eventpb.Event:
			switch pbEvT := evT.Type.(type) {
			case *eventpb.Event_Transport:
				switch pbTEvT := transportpbtypes.EventFromPb(pbEvT.Transport).Type.(type) {
				case *transportpbtypes.Event_SendMessage:
					wrappedPayload := messages.WrapPbMessage(pbTEvT.SendMessage.Msg.Pb(), msgMetadata)
					nEv = stdevents.NewSendMessageWithSrc(evT.Src(), evT.Dest(), pbTEvT.SendMessage.Msg.DestModule, wrappedPayload, pbTEvT.SendMessage.Destinations...)
				case *transportpbtypes.Event_MessageReceived:
				default:
					return nil, es.Errorf("unexpected type of transport event: %T", pbTEvT)
				}
			default:
				nEv = ev
			}
		default:
			nEv = ev
		}

		newEvents.PushBack(nEv)
	}
	mi.logger.Log(logging.LevelTrace, "Intercepting OUT done", "events", newEvents)
	return newEvents, nil
}
