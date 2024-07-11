package events

import (
	"fmt"
	"maps"

	"github.com/filecoin-project/mir/stdtypes"
)

type PingPong struct {
	baseEvent
	SeqNr uint64
}

func NewPingPong(destModule stdtypes.ModuleID, seqNr uint64) *PingPong {
	return &PingPong{
		baseEvent: newBaseEvent(destModule),
		SeqNr:     seqNr,
	}
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
