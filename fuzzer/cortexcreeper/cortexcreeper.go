package cortexcreeper

import (
	"context"
	"fmt"

	"github.com/filecoin-project/mir"
	"github.com/filecoin-project/mir/pkg/eventlog"
	"github.com/filecoin-project/mir/stdtypes"
	es "github.com/go-errors/errors"
)

type CortexCreeper interface {
	Setup(node *mir.Node) error
	RunInjector(ctx context.Context) error
	StopInjector()
	eventlog.Interceptor
}

type ByzantineCortexCreeper struct {
	cortexCreeperLink CortexCreeperLink
	node              *mir.Node
	cancel            context.CancelFunc
}

func NewByzantineCortexCreeper(link CortexCreeperLink) CortexCreeper {
	return &ByzantineCortexCreeper{
		cortexCreeperLink: link,
		node:              nil,
	}
}

func (cc *ByzantineCortexCreeper) Setup(node *mir.Node) error {
	cc.node = node
	return nil
}

func (cc *ByzantineCortexCreeper) RunInjector(ctx context.Context) error {
	if cc.node == nil {
		return es.Errorf("Initialize cortex creeper with node before running it (Setup).")
	}

	ctx, cancel := context.WithCancel(ctx)
	cc.cancel = cancel

	for {
		select {
		case newEvts := <-cc.cortexCreeperLink.ToNode:
			// hard coded only one event
			// TODO: handle multiple events
			ev, _ := newEvts.Slice()[0].SetMetadata("injected", true)
			cc.node.InjectEvents(ctx, stdtypes.ListOf(ev))
		case <-ctx.Done():
			return nil
		}
	}
}

func (cc *ByzantineCortexCreeper) StopInjector() {
	if cc.cancel != nil {
		cc.cancel()
	}
}

func (cc *ByzantineCortexCreeper) Intercept(evts *stdtypes.EventList) (*stdtypes.EventList, error) {
	// if it is from adv, forward directly - for now, we always know that these lists only contain one element
	// TODO: handle multi element

	if evts.Len() == 1 {
		if _, err := evts.Slice()[0].GetMetadata("injected"); err == nil {
			fmt.Printf("Injected - not sending to adv (%v)\n", evts.Slice()[0])
			return evts, nil
		}
	}

	// events to adversary
	cc.cortexCreeperLink.ToAdversary <- evts
	return stdtypes.EmptyList(), nil
}

type CortexCreeperLink struct {
	ToAdversary chan *stdtypes.EventList
	ToNode      chan *stdtypes.EventList
}

func NewCortexCreeperLink() CortexCreeperLink {
	return CortexCreeperLink{
		make(chan *stdtypes.EventList),
		make(chan *stdtypes.EventList),
	}
}
