package cortexcreeper

import (
	"context"
	"fmt"

	"github.com/filecoin-project/mir"
	"github.com/filecoin-project/mir/stdtypes"
	es "github.com/go-errors/errors"
)

var ErrorShutdown = fmt.Errorf("shutdown signal received")

type CortexCreepers map[stdtypes.NodeID]*CortexCreeper

type CortexCreeper struct {
	node           *mir.Node
	eventsIn       chan *stdtypes.EventList
	eventsOut      chan *stdtypes.EventList
	IdleDetectionC chan chan struct{}
	doneC          chan struct{}
}

func NewCortexCreeper() *CortexCreeper {
	return &CortexCreeper{
		node:           nil,
		eventsIn:       make(chan *stdtypes.EventList),
		eventsOut:      make(chan *stdtypes.EventList),
		IdleDetectionC: make(chan chan struct{}),
		doneC:          make(chan struct{}),
	}
}

func (cc *CortexCreeper) Setup(node *mir.Node) error {
	cc.node = node
	return nil
}

func (cc *CortexCreeper) Run(ctx context.Context) error {
	if cc.node == nil {
		return es.Errorf("Initialize cortex creeper with node before running it (Setup).")
	}

	for {
		select {
		case evts := <-cc.eventsOut:
			// hard coded only one event
			// TODO: handle multiple events
			markedEvts := stdtypes.EmptyList()
			evtsIter := evts.Iterator()
			for e := evtsIter.Next(); e != nil; e = evtsIter.Next() {
				markedE, _ := e.SetMetadata("injected", true)
				markedEvts.PushBack(markedE)
			}
			cc.node.InjectEvents(ctx, markedEvts)
		case <-ctx.Done():
			close(cc.doneC)
			return nil
		case <-cc.doneC:
			return ErrorShutdown
		}
	}
}

func (cc *CortexCreeper) StopInjector() {
	close(cc.doneC)
}

func (cc *CortexCreeper) Intercept(evts *stdtypes.EventList) (*stdtypes.EventList, error) {
	// forward directly if injected, if one element is injected then all in the list are so we must only check one
	if evts.Len() >= 1 {
		if _, err := evts.Slice()[0].GetMetadata("injected"); err == nil {
			return evts, nil
		}
	}

	// events to adversary
	select {
	case cc.eventsIn <- evts:
	case <-cc.doneC:
		return nil, ErrorShutdown

		// TODO deal with closing the channels
	}
	return stdtypes.EmptyList(), nil
}

func (cc *CortexCreeper) PopEvents(ctx context.Context) *stdtypes.EventList {
	select {
	case <-ctx.Done():
		return nil
	case evts := <-cc.eventsIn:
		return evts
	}
}

func (cc *CortexCreeper) GetEventsIn() chan *stdtypes.EventList {
	return cc.eventsIn
}

func (cc *CortexCreeper) PushEvents(evts *stdtypes.EventList) {
	cc.eventsOut <- evts
}
