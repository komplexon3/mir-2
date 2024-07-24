/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// TODO: This is the original old code with very few modifications.
//       Go through all of it, comment what is to be kept and delete what is not needed.

package network

import (
	"context"
	"fmt"
	"sync"

	"github.com/filecoin-project/mir/stdevents"
	"github.com/filecoin-project/mir/stdtypes"

	es "github.com/go-errors/errors"

	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/net"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	"github.com/filecoin-project/mir/pkg/pb/messagepb"
	messagepbtypes "github.com/filecoin-project/mir/pkg/pb/messagepb/types"
	transportpbevents "github.com/filecoin-project/mir/pkg/pb/transportpb/events"
	transportpbtypes "github.com/filecoin-project/mir/pkg/pb/transportpb/types"
	trantorpbtypes "github.com/filecoin-project/mir/pkg/pb/trantorpb/types"
)

// Module Part

type FuzzLink struct {
	FuzzTransport *FuzzTransport
	Source        stdtypes.NodeID
	DoneC         chan struct{}
	wg            sync.WaitGroup
}

func (fl *FuzzLink) ApplyEvents(
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
				fl.FuzzTransport.ftObserver.RegisterSend()
				if destID == fl.Source {
					// Send message to myself bypassing the network.
					// The sending must be done in its own goroutine in case writing to tr.incomingMessages blocks.
					// (Processing of input events must be non-blocking.)
					receiveEvent := stdevents.NewMessageReceived(evt.RemoteDestModule, fl.Source, evt.Payload)
					eventsOut := fl.FuzzTransport.NodeSinks[fl.Source]
					go func() {
						select {
						case eventsOut <- stdtypes.ListOf(receiveEvent):
							fl.FuzzTransport.ftObserver.RegisterReceive()
							fl.FuzzTransport.logger.Log(logging.LevelTrace, "Pushed onto evtsOut (shortcut)", "event", receiveEvent)
						case <-ctx.Done():
						}
					}()
				} else {
					// Send message to another node.
					if err := fl.SendRawMessage(destID, evt.RemoteDestModule, evt.Payload); err != nil {
						fl.FuzzTransport.logger.Log(logging.LevelWarn, "failed to send a message", "err", err)
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

func (fl *FuzzLink) ApplyPbEvent(ctx context.Context, evt *eventpb.Event) error {
	switch e := evt.Type.(type) {
	case *eventpb.Event_Transport:
		switch e := transportpbtypes.EventFromPb(e.Transport).Type.(type) {
		case *transportpbtypes.Event_SendMessage:
			for _, destID := range e.SendMessage.Destinations {
				fl.FuzzTransport.ftObserver.RegisterSend()
				if destID == fl.Source {
					// Send message to myself bypassing the network.
					receivedEvent := transportpbevents.MessageReceived(
						e.SendMessage.Msg.DestModule,
						fl.Source,
						e.SendMessage.Msg,
					)
					eventsOut := fl.FuzzTransport.NodeSinks[fl.Source]
					go func() {
						select {
						case eventsOut <- stdtypes.ListOf(receivedEvent.Pb()):
							fl.FuzzTransport.ftObserver.RegisterReceive()
							fl.FuzzTransport.logger.Log(logging.LevelTrace, "Pushed onto evtsOut (shortcut)", "event", receivedEvent.Pb())
						case <-ctx.Done():
						}
					}()
				} else {
					// Send message to another node.
					if err := fl.Send(destID, e.SendMessage.Msg.Pb()); err != nil {
						fl.FuzzTransport.logger.Log(logging.LevelWarn, "failed to send a message", "err", err)
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
func (fl *FuzzLink) ImplementsModule() {}

func (fl *FuzzLink) Send(dest stdtypes.NodeID, msg *messagepb.Message) error {
	fl.FuzzTransport.Send(fl.Source, dest, msg)
	return nil
}

func (fl *FuzzLink) SendRawMessage(destNode stdtypes.NodeID, destModule stdtypes.ModuleID, message stdtypes.Message) error {
	fl.FuzzTransport.SendRawMessage(fl.Source, destNode, destModule, message)
	return nil
}

func (fl *FuzzLink) EventsOut() <-chan *stdtypes.EventList {
	return fl.FuzzTransport.NodeSinks[fl.Source]
}

// Underlying "network"

type FuzzTransport struct {
	// Buffers is source x dest
	Buffers    map[stdtypes.NodeID]map[stdtypes.NodeID]chan *stdtypes.EventList
	NodeSinks  map[stdtypes.NodeID]chan *stdtypes.EventList
	ftObserver *FuzzTransportObserver
	logger     logging.Logger
}

func NewFuzzTransport(nodeIDs []stdtypes.NodeID, inactiveNotificationC chan chan struct{}, logger logging.Logger, ctx context.Context) *FuzzTransport {
	buffers := make(map[stdtypes.NodeID]map[stdtypes.NodeID]chan *stdtypes.EventList, len(nodeIDs))
	nodeSinks := make(map[stdtypes.NodeID]chan *stdtypes.EventList, len(nodeIDs))
	for _, sourceID := range nodeIDs {
		buffers[sourceID] = make(map[stdtypes.NodeID]chan *stdtypes.EventList, len(nodeIDs)-1)
		for _, destID := range nodeIDs {
			if sourceID == destID {
				continue
			}
			buffers[sourceID][destID] = make(chan *stdtypes.EventList, 1000)
		}
		nodeSinks[sourceID] = make(chan *stdtypes.EventList)
	}

	ft := &FuzzTransport{
		Buffers:   buffers,
		NodeSinks: nodeSinks,
		logger:    logger,
	}

	if inactiveNotificationC != nil {
		ft.ftObserver = NewFuzzTransportObserver(ft, inactiveNotificationC)
		go func() {
			err := ft.ftObserver.Run(ctx)
			if err != nil {
				panic(es.Errorf("FuzzTransportObserver failed unexpectedly: %v", err))
			}
		}()
	}

	return ft
}

func (ft *FuzzTransport) SendRawMessage(sourceNode, destNode stdtypes.NodeID, destModule stdtypes.ModuleID, message stdtypes.Message) error {
	ft.Buffers[sourceNode][destNode] <- stdtypes.ListOf(
		stdevents.NewMessageReceived(destModule, sourceNode, message),
	)

	return nil
}

func (ft *FuzzTransport) Send(sourceNode, destNode stdtypes.NodeID, msg *messagepb.Message) {
	// block incase the system is paused
	ft.Buffers[sourceNode][destNode] <- stdtypes.ListOf(
		transportpbevents.MessageReceived(stdtypes.ModuleID(msg.DestModule), sourceNode, messagepbtypes.MessageFromPb(msg)).Pb(),
	)
}

func (ft *FuzzTransport) Link(source stdtypes.NodeID) (net.Transport, error) {
	return &FuzzLink{
		Source:        source,
		FuzzTransport: ft,
		DoneC:         make(chan struct{}),
	}, nil
}

func (ft *FuzzTransport) Close() {

}

func (fl *FuzzLink) CloseOldConnections(_ *trantorpbtypes.Membership) {}

func (fl *FuzzLink) Start() error {
	return nil
}

func (fl *FuzzLink) Connect(_ *trantorpbtypes.Membership) {
	sourceBuffers := fl.FuzzTransport.Buffers[fl.Source]

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
					fl.FuzzTransport.logger.Log(logging.LevelTrace, "Popped msg - trying to push onto evtsOut", "event", msg)
					select {
					case fl.FuzzTransport.NodeSinks[destID] <- msg:
						fl.FuzzTransport.ftObserver.RegisterReceive()
						fl.FuzzTransport.logger.Log(logging.LevelTrace, "Pushed onto evtsOut", "event", msg)
					case <-fl.DoneC:
						return
					}
				case <-fl.DoneC:
					return
				}
				// every iteration corresponds to 'receive', or done
			}
		}(destID, buffer)
	}
}

// WaitFor returns immediately.
// It does not need to wait for anything, since the Connect() function already waits for all the connections.
func (fl *FuzzLink) WaitFor(_ int) error {
	return nil
}

func (fl *FuzzLink) Stop() {
	close(fl.DoneC)
	fl.wg.Wait()
}

type FuzzTransportObserver struct {
	FuzzTransport         *FuzzTransport
	sendRevcTrackC        chan bool
	inactiveNotificationC chan chan struct{}
	doneC                 chan struct{}
	wg                    sync.WaitGroup
}

func NewFuzzTransportObserver(ft *FuzzTransport, inactiveNotificationChan chan chan struct{}) *FuzzTransportObserver {
	return &FuzzTransportObserver{
		FuzzTransport:         ft,
		sendRevcTrackC:        make(chan bool),
		inactiveNotificationC: inactiveNotificationChan,
		doneC:                 make(chan struct{}),
	}
}

func (fto *FuzzTransportObserver) RegisterSend() {
	if fto == nil {
		return
	}
	fto.FuzzTransport.logger.Log(logging.LevelTrace, "send +++")
	fto.sendRevcTrackC <- true
}

func (fto *FuzzTransportObserver) RegisterReceive() {
	if fto == nil {
		return
	}
	fto.FuzzTransport.logger.Log(logging.LevelTrace, "receive ---")
	fto.sendRevcTrackC <- false
}

func (fto *FuzzTransportObserver) Run(ctx context.Context) error {
	fto.wg.Add(1)
	defer fto.wg.Done()

	msgsInTransit := 0
	var activeAgainChan chan struct{}
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-fto.doneC:
			return nil
		case wasSend := <-fto.sendRevcTrackC:
			if wasSend {
				msgsInTransit++
				// if consumer currently thinks the network is inactive
				if activeAgainChan != nil {
					close(activeAgainChan)
					activeAgainChan = nil
				}
			} else {
				msgsInTransit--
			}
		}

		fto.FuzzTransport.logger.Log(logging.LevelTrace, fmt.Sprintf("%d msg in transit", msgsInTransit))

		if msgsInTransit == 0 && len(fto.sendRevcTrackC) == 0 {
			activeAgainChan = make(chan struct{})
			fto.inactiveNotificationC <- activeAgainChan
		} else if msgsInTransit < 0 {
			return es.Errorf("number of msgs in transit should never be negative! (current count: %d)", msgsInTransit)
		}
	}
}

func (fto *FuzzTransportObserver) Stop() {
	close(fto.doneC)
	fto.wg.Wait()
}
