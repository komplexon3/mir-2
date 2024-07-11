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
				if destID == fl.Source {
					// Send message to myself bypassing the network.
					// The sending must be done in its own goroutine in case writing to tr.incomingMessages blocks.
					// (Processing of input events must be non-blocking.)
					receiveEvent := stdevents.NewMessageReceived(evt.RemoteDestModule, fl.Source, evt.Payload)
					eventsOut := fl.FuzzTransport.NodeSinks[fl.Source]
					go func() {
						select {
						case eventsOut <- stdtypes.ListOf(receiveEvent):
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
	logger     logging.Logger
	pauseChans map[stdtypes.NodeID]chan chan struct{}
	pause      bool
	pauseLock  *sync.RWMutex
	pauseCond  *sync.Cond
}

func NewFuzzTransport(nodeIDs []stdtypes.NodeID) *FuzzTransport {
	buffers := make(map[stdtypes.NodeID]map[stdtypes.NodeID]chan *stdtypes.EventList, len(nodeIDs))
	nodeSinks := make(map[stdtypes.NodeID]chan *stdtypes.EventList, len(nodeIDs))
	pauseChans := make(map[stdtypes.NodeID]chan chan struct{})
	for _, sourceID := range nodeIDs {
		pauseChans[sourceID] = make(chan chan struct{})
		buffers[sourceID] = make(map[stdtypes.NodeID]chan *stdtypes.EventList, len(nodeIDs)-1)
		for _, destID := range nodeIDs {
			if sourceID == destID {
				continue
			}
			buffers[sourceID][destID] = make(chan *stdtypes.EventList, 10000)
		}
		nodeSinks[sourceID] = make(chan *stdtypes.EventList)
	}

	pauseLock := sync.RWMutex{}

	return &FuzzTransport{
		Buffers:    buffers,
		NodeSinks:  nodeSinks,
		logger:     logging.ConsoleErrorLogger,
		pauseChans: pauseChans,
		pause:      false,
		pauseLock:  &pauseLock,
		pauseCond:  sync.NewCond(pauseLock.RLocker()),
	}
}

// TODO: should probably have a context
func (ft *FuzzTransport) Pause() (func(), chan struct{}) {
	ft.pauseLock.Lock()
	ft.pause = true
	ft.pauseLock.Unlock()

	wgPaused := &sync.WaitGroup{}
	wgResumed := &sync.WaitGroup{}
	doneC := make(chan struct{})
	pausedNotificationChannel := make(chan struct{})

	// pause all comm consume goroutines
	for _, pc := range ft.pauseChans {
		wgPaused.Add(1)
		wgResumed.Add(1)
		go func(pc chan chan struct{}, dc chan struct{}, wgP *sync.WaitGroup, wgR *sync.WaitGroup) {
			continueChan := make(chan struct{})
			defer close(continueChan)
			select {
			case <-dc:
				wgP.Done()
				wgR.Done()
				return
			case pc <- continueChan:
				wgP.Done()
				<-dc
				continueChan <- struct{}{}
				wgR.Done()
			}
		}(pc, doneC, wgPaused, wgResumed)
	}

	// wait until all consume gorourines are paused

	go func(pnC chan struct{}, wgP *sync.WaitGroup) {
		wgP.Wait()
		close(pnC)
	}(pausedNotificationChannel, wgPaused)

	// release all comm channels
	// resume/abort pause
	return func() {
		ft.pauseLock.Lock()
		ft.pause = false
		ft.pauseCond.Broadcast()
		ft.pauseLock.Unlock()

		close(doneC)

		// wait until all consumer routines have resumed
		wgResumed.Wait()
	}, pausedNotificationChannel
}

func (ft *FuzzTransport) CountMessagesInTransit() (int, error) {
	ft.pauseLock.RLock()
	defer ft.pauseLock.RUnlock()
	if ft.pause {
		return 0, es.Errorf("cannot count messages in transit while the network is running - you must pause it first ")
	}

	count := 0

	for _, bufferGroup := range ft.Buffers {
		for _, buffer := range bufferGroup {
			count += len(buffer)
		}
	}

	return count, nil
}

func (ft *FuzzTransport) SendRawMessage(sourceNode, destNode stdtypes.NodeID, destModule stdtypes.ModuleID, message stdtypes.Message) error {
	// block incase the system is paused
	ft.pauseLock.RLock()
	if ft.pause {
		ft.pauseCond.Wait()
	}
	ft.pauseLock.RUnlock()
	select {
	case ft.Buffers[sourceNode][destNode] <- stdtypes.ListOf(
		stdevents.NewMessageReceived(destModule, sourceNode, message),
	):
	default:
		fmt.Printf("Warning: Dropping message %v from %s to %s\n", message, sourceNode, destNode)
	}

	return nil
}

func (ft *FuzzTransport) Send(sourceNode, destNode stdtypes.NodeID, msg *messagepb.Message) {
	// block incase the system is paused
	ft.pauseLock.RLock()
	if ft.pause {
		ft.pauseCond.Wait()
	}
	ft.pauseLock.RUnlock()
	select {
	case ft.Buffers[sourceNode][destNode] <- stdtypes.ListOf(
		transportpbevents.MessageReceived(stdtypes.ModuleID(msg.DestModule), sourceNode, messagepbtypes.MessageFromPb(msg)).Pb(),
	):
	default:
		fmt.Printf("Warning: Dropping message %T from %s to %s\n", msg.Type, sourceNode, destNode)
	}
}

func (ft *FuzzTransport) Link(source stdtypes.NodeID) (net.Transport, error) {
	return &FuzzLink{
		Source:        source,
		FuzzTransport: ft,
		DoneC:         make(chan struct{}),
	}, nil
}

func (ft *FuzzTransport) Close() {}

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
				case cc := <-fl.FuzzTransport.pauseChans[destID]:
					// block waiting on continue signal
					<-cc
				case msg := <-buffer:
					// if paused but already here, it will still push it onto node sink, which is the event out for the node
					fl.FuzzTransport.logger.Log(logging.LevelTrace, "Popped msg - trying to push onto evtsOut", "event", msg)
					select {
					case fl.FuzzTransport.NodeSinks[destID] <- msg:
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
