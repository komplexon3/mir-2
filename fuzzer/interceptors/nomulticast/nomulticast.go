package nomulticast

import (
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	"github.com/filecoin-project/mir/stdevents"
	"github.com/filecoin-project/mir/stdtypes"
)

type NoMulticast struct{}

func (nm *NoMulticast) Intercept(events *stdtypes.EventList) (*stdtypes.EventList, error) {
	newEvts := stdtypes.EmptyList()
	evtIterator := events.Iterator()
	for e := evtIterator.Next(); e != nil; e = evtIterator.Next() {
		switch eT := e.(type) {
		case *stdevents.SendMessage:
			if len(eT.DestNodes) == 1 {
				newEvts.PushBack(e)
			} else {
				for _, destNode := range eT.DestNodes {
					singleDestSendEvent := stdevents.NewSendMessage(eT.DestModule, eT.RemoteDestModule, eT.Payload, destNode)
					for key, val := range eT.Metadata {
						singleDestSendEvent.Metadata[key] = val
					}
					newEvts.PushBack(singleDestSendEvent)
				}
			}

		case *eventpb.Event:
			switch eT.Type.(type) {
			case *eventpb.Event_Transport:
				panic("No multicast interceptor can only be used with non-proto transport (send/receive) events")
			default:
				newEvts.PushBack(e)
			}
		default:
			newEvts.PushBack(e)
		}
	}
	return newEvts, nil
}
