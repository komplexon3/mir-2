package events

import (
	"fmt"
	"maps"

	"github.com/filecoin-project/mir/stdtypes"
	"github.com/google/uuid"
)

type Deliver struct {
	baseEvent
	BroadcastID     uuid.UUID
	BroadcastSender stdtypes.NodeID
	Data            []byte
}

func NewDeliver(destModule stdtypes.ModuleID, data []byte, broadcastID uuid.UUID, broadcastSender stdtypes.NodeID) *Deliver {
	return &Deliver{
		baseEvent:       newBaseEvent(destModule),
		Data:            data,
		BroadcastID:     broadcastID,
		BroadcastSender: broadcastSender,
	}
}

func (e *Deliver) NewSrc(newSrc stdtypes.ModuleID) stdtypes.Event {
	newEvent := *e
	newEvent.SrcModule = newSrc
	return &newEvent
}

func (e *Deliver) NewDest(newDest stdtypes.ModuleID) stdtypes.Event {
	newEvent := *e
	newEvent.DestModule = newDest
	return &newEvent
}

func (br *Deliver) ToBytes() ([]byte, error) {
	return serialize(br)
}

func (br *Deliver) ToString() string {
	data, err := br.ToBytes()
	if err != nil {
		return fmt.Sprintf("unmarshalableEvent(%+v)", br)
	}

	return string(data)
}

func (e *Deliver) SetMetadata(key string, value interface{}) (stdtypes.Event, error) {
	newE := *e
	metadata := maps.Clone(newE.Metadata)
	metadata[key] = value
	newE.Metadata = metadata
	return &newE, nil
}
