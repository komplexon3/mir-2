package cortexcreeper

import (
	"context"
	"fmt"

	"github.com/filecoin-project/mir"
	"github.com/filecoin-project/mir/stdtypes"
)

var ErrorShutdown = fmt.Errorf("shutdown signal received")

type CortexCreepers map[stdtypes.NodeID]*CortexCreeper

type EventsAck struct {
	Events *stdtypes.EventList
	Ack    chan struct{}
}

func newEventsAck(events *stdtypes.EventList) *EventsAck {
	return &EventsAck{
		Events: events,
		Ack:    make(chan struct{}),
	}
}

type CortexCreeper struct {
	node           *mir.Node
	eventsIn       chan *EventsAck
	eventsOut      chan *stdtypes.EventList
	interceptDoneC chan struct{}
}

func NewCortexCreeper() *CortexCreeper {
	return &CortexCreeper{
		node:           nil,
		eventsIn:       make(chan *EventsAck),
		eventsOut:      make(chan *stdtypes.EventList),
		interceptDoneC: make(chan struct{}),
	}
}

func (cc *CortexCreeper) Setup(node *mir.Node) error {
	cc.node = node
	return nil
}

func (cc *CortexCreeper) AbortIntercepts() {
	close(cc.interceptDoneC)
}

func (cc *CortexCreeper) Intercept(evts *stdtypes.EventList) (*stdtypes.EventList, error) {
	// forward directly if injected, if one element is injected then all in the list are so we must only check one
	if evts.Len() >= 1 {
		if _, err := evts.Slice()[0].GetMetadata("injected"); err == nil {
			return evts, nil
		}
	}

	evtsAck := newEventsAck(evts)

	// events to adversary
	select {
	case cc.eventsIn <- evtsAck:
	case <-cc.interceptDoneC:
	}
	select {
	case <-evtsAck.Ack:
	case <-cc.interceptDoneC:
	}
	return stdtypes.EmptyList(), nil
}

func (cc *CortexCreeper) GetEventsIn() chan *EventsAck {
	return cc.eventsIn
}

func (cc *CortexCreeper) PushEvents(ctx context.Context, evts *stdtypes.EventList) {
	// hard coded only one event
	// TODO: handle multiple events
	markedEvts := stdtypes.EmptyList()
	evtsIter := evts.Iterator()
	for e := evtsIter.Next(); e != nil; e = evtsIter.Next() {
		markedE, _ := e.SetMetadata("injected", true)
		markedEvts.PushBack(markedE)
	}
	cc.node.InjectEvents(ctx, markedEvts)
}
