package cortexcreeper

import (
	"context"

	"github.com/filecoin-project/mir"
	"github.com/filecoin-project/mir/stdtypes"
	es "github.com/go-errors/errors"
)

type CortexCreeper struct {
	node           *mir.Node
	cancel         context.CancelFunc
	eventsIn       chan *stdtypes.EventList
	eventsOut      chan *stdtypes.EventList
	abort          chan struct{}
	IdleDetectionC chan chan struct{}
}

func NewCortexCreeper() *CortexCreeper {
	return &CortexCreeper{
		node:           nil,
		eventsIn:       make(chan *stdtypes.EventList),
		eventsOut:      make(chan *stdtypes.EventList),
		abort:          make(chan struct{}),
		IdleDetectionC: make(chan chan struct{}),
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

	ctx, cancel := context.WithCancel(ctx)
	cc.cancel = cancel

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
			defer close(cc.abort)
			cc.abort <- struct{}{}
			return nil
		}
	}
}

func (cc *CortexCreeper) StopInjector() {
	if cc.cancel != nil {
		cc.cancel()
	}
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
	case <-cc.abort:
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
