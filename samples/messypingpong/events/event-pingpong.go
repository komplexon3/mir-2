package events

import (
	"fmt"
	"maps"
	"math/rand/v2"

	"github.com/filecoin-project/mir/stdtypes"
)

type PingPong struct {
	baseEvent
	InitiatorNode   stdtypes.NodeID
	InitiatorModule stdtypes.ModuleID
	SeqNr           uint64
}

func NewPingPong(srcNode stdtypes.NodeID, srcModule stdtypes.ModuleID, seqNr uint64) *PingPong {
	destModule := stdtypes.ModuleID("B")
	if rand.Float32() > 0.5 {
		destModule = stdtypes.ModuleID("A")
	}
	return &PingPong{
		baseEvent:       newBaseEvent(destModule),
		InitiatorNode:   srcNode,
		InitiatorModule: srcModule,
		SeqNr:           seqNr,
	}
}

func (e *PingPong) Next() *PingPong {
	return NewPingPong(e.InitiatorNode, e.InitiatorModule, e.SeqNr+1)
}

func (e *PingPong) NewSrc(newSrc stdtypes.ModuleID) stdtypes.Event {
	newEvent := *e
	newEvent.SrcModule = newSrc
	return &newEvent
}

func (e *PingPong) NewDest(newDest stdtypes.ModuleID) stdtypes.Event {
	newEvent := *e
	newEvent.DestModule = newDest
	return &newEvent
}

func (br *PingPong) ToBytes() ([]byte, error) {
	return serialize(br)
}

func (br *PingPong) ToString() string {
	data, err := br.ToBytes()
	if err != nil {
		return fmt.Sprintf("unmarshalableEvent(%+v)", br)
	}

	return string(data)
}

func (e *PingPong) SetMetadata(key string, value interface{}) (stdtypes.Event, error) {
	newE := *e
	metadata := maps.Clone(newE.Metadata)
	metadata[key] = value
	newE.Metadata = metadata
	return &newE, nil
}

func getDestModule() stdtypes.ModuleID {
	if rand.Float32() > 0.5 {
		return stdtypes.ModuleID("A")
	}
	return stdtypes.ModuleID("B")
}
