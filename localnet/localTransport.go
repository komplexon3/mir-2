package localnet

import (
	"context"
	"sync"

	"github.com/filecoin-project/mir/stdevents"
	"github.com/filecoin-project/mir/stdtypes"

	es "github.com/go-errors/errors"

	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	"github.com/filecoin-project/mir/pkg/pb/messagepb"
	transportpbevents "github.com/filecoin-project/mir/pkg/pb/transportpb/events"
	transportpbtypes "github.com/filecoin-project/mir/pkg/pb/transportpb/types"
	trantorpbtypes "github.com/filecoin-project/mir/pkg/pb/trantorpb/types"
)

type LocalTransport struct {
	LocalNetwork *LocalNetwork
	Source       stdtypes.NodeID
	DoneC        chan struct{}
	wg           sync.WaitGroup
}

func (fl *LocalTransport) ApplyEvents(
	ctx context.Context,
	eventList *stdtypes.EventList,
) error {
	iter := eventList.Iterator()
	for event := iter.Next(); event != nil; event = iter.Next() {
		switch evt := event.(type) {
		case *stdevents.Init:
			// no actions on init
		case *stdevents.SendMessage:
			for _, destID := range evt.DestNodes {
				if destID == fl.Source {
					// Send message to myself bypassing the network.
					// The sending must be done in its own goroutine in case writing to tr.incomingMessages blocks.
					// (Processing of input events must be non-blocking.)
					receiveEvent := stdevents.NewMessageReceived(evt.RemoteDestModule, fl.Source, evt.Payload)
					eventsOut := fl.LocalNetwork.NodeSinks[fl.Source]
					go func() {
						select {
						case eventsOut <- stdtypes.ListOf(receiveEvent):
						case <-ctx.Done():
						}
					}()
				} else {
					// Send message to another node.
					if err := fl.SendRawMessage(destID, evt.RemoteDestModule, evt.Payload); err != nil {
						fl.LocalNetwork.logger.Log(logging.LevelWarn, "failed to send a message", "err", err)
					}
				}
			}
		case *eventpb.Event:
			return fl.ApplyPbEvent(ctx, evt)
		default:
			return es.Errorf("GRPC transport only supports proto events and OutgoingMessage, received %T", event)
		}
	}

	return nil
}

func (fl *LocalTransport) ApplyPbEvent(ctx context.Context, evt *eventpb.Event) error {

	switch e := evt.Type.(type) {
	case *eventpb.Event_Transport:
		switch e := transportpbtypes.EventFromPb(e.Transport).Type.(type) {
		case *transportpbtypes.Event_SendMessage:
			for _, destID := range e.SendMessage.Destinations {
				if destID == fl.Source {
					// Send message to myself bypassing the network.

					receivedEvent := transportpbevents.MessageReceived(
						e.SendMessage.Msg.DestModule,
						fl.Source,
						e.SendMessage.Msg,
					)
					eventsOut := fl.LocalNetwork.NodeSinks[fl.Source]
					go func() {
						select {
						case eventsOut <- stdtypes.ListOf(receivedEvent.Pb()):
						case <-ctx.Done():
						}
					}()
				} else {
					// Send message to another node.
					if err := fl.Send(destID, e.SendMessage.Msg.Pb()); err != nil {
						fl.LocalNetwork.logger.Log(logging.LevelWarn, "failed to send a message", "err", err)
					}
				}
			}
		default:
			return es.Errorf("unexpected transport event type: %T", e)
		}
	default:
		return es.Errorf("unexpected type of Net event: %T", evt.Type)
	}
	return nil
}

// The ImplementsModule method only serves the purpose of indicating that this is a Module and must not be called.
func (fl *LocalTransport) ImplementsModule() {}

func (fl *LocalTransport) Send(dest stdtypes.NodeID, msg *messagepb.Message) error {
	fl.LocalNetwork.Send(fl.Source, dest, msg)
	return nil
}

func (fl *LocalTransport) SendRawMessage(destNode stdtypes.NodeID, destModule stdtypes.ModuleID, message stdtypes.Message) error {
	fl.LocalNetwork.SendRawMessage(fl.Source, destNode, destModule, message)
	return nil
}

func (fl *LocalTransport) EventsOut() <-chan *stdtypes.EventList {
	return fl.LocalNetwork.NodeSinks[fl.Source]
}

func (ft *LocalTransport) Close() {}

func (fl *LocalTransport) CloseOldConnections(_ *trantorpbtypes.Membership) {}

func (fl *LocalTransport) Start() error {
	return nil
}

func (fl *LocalTransport) Connect(_ *trantorpbtypes.Membership) {
	sourceBuffers := fl.LocalNetwork.Buffers[fl.Source]

	fl.wg.Add(len(sourceBuffers))

	for destID, buffer := range sourceBuffers {
		if fl.Source == destID {
			fl.wg.Done()
			continue
		}
		go func(destID stdtypes.NodeID, buffer chan *stdtypes.EventList) {
			defer fl.wg.Done()
			for {
				select {
				case msg := <-buffer:
					select {
					case fl.LocalNetwork.NodeSinks[destID] <- msg:
					case <-fl.DoneC:
						return
					}
				case <-fl.DoneC:
					return
				}
			}
		}(destID, buffer)
	}
}

// WaitFor returns immediately.
// It does not need to wait for anything, since the Connect() function already waits for all the connections.
func (fl *LocalTransport) WaitFor(_ int) error {
	return nil
}

func (fl *LocalTransport) Stop() {
	close(fl.DoneC)
	fl.wg.Wait()
}
