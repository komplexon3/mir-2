package cortexcreeper

import (
	"context"
	"fmt"

	"github.com/filecoin-project/mir"
	"github.com/filecoin-project/mir/pkg/eventlog"
	"github.com/filecoin-project/mir/stdtypes"
	es "github.com/go-errors/errors"
)

type CortexCreeper struct {
	node      *mir.Node
	cancel    context.CancelFunc
	eventsIn  chan *stdtypes.EventList
	eventsOut chan *stdtypes.EventList
	abort     chan struct{}
	eventlog.Interceptor
}

func NewCortexCreeper() *CortexCreeper {
	return &CortexCreeper{
		node:      nil,
		eventsIn:  make(chan *stdtypes.EventList),
		eventsOut: make(chan *stdtypes.EventList),
		abort:     make(chan struct{}),
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
		case newEvts := <-cc.eventsOut:
			// hard coded only one event
			// TODO: handle multiple events
			ev, _ := newEvts.Slice()[0].SetMetadata("injected", true)
			cc.node.InjectEvents(ctx, stdtypes.ListOf(ev))
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
	// if it is from adv, forward directly - for now, we always know that these lists only contain one element
	// TODO: handle multi element
	if evts.Len() == 1 {
		fmt.Printf("%v\n", evts.Slice()[0])
		if _, err := evts.Slice()[0].GetMetadata("injected"); err == nil {
			fmt.Printf("Injected - not sending to adv (%v)\n", evts.Slice()[0])
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
