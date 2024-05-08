package adversary

import (
	"github.com/filecoin-project/mir/pkg/eventlog"
	"github.com/filecoin-project/mir/stdtypes"
)

// Rename to Headjack for the matrix reference?

type CortexCreeper = eventlog.Interceptor

type ByzantineCortexCreeper struct {
  cortexCreeperLink CortexCreeperLink
}

func NewByzantineCortexCreeper(link CortexCreeperLink) CortexCreeper {
	return &ByzantineCortexCreeper{
    cortexCreeperLink: link,
	}
}

func (cc *ByzantineCortexCreeper) Intercept(evts *stdtypes.EventList) (*stdtypes.EventList, error) {
	// events to adversary
	cc.cortexCreeperLink.toAdversary <- evts
	// and back
	newEvts := <-cc.cortexCreeperLink.toNode
	return newEvts, nil
}

type BenignCortexCreeper struct{}

// A begnign cortex creeper for non-byzantine nodes
func NewBenignCortexCreeper() CortexCreeper {
	return &BenignCortexCreeper{}
}

func (cc *BenignCortexCreeper) Intercept(evts *stdtypes.EventList) (*stdtypes.EventList, error) {
	return evts, nil
}

type CortexCreeperLink struct {
	toAdversary chan *stdtypes.EventList
	toNode      chan *stdtypes.EventList
}

func NewCortexCreeperLink() CortexCreeperLink {
  return CortexCreeperLink{
    make(chan *stdtypes.EventList),
    make(chan *stdtypes.EventList),
  }
}
