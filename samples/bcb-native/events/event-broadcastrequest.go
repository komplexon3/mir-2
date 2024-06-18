package events

import (
	"fmt"

	"github.com/filecoin-project/mir/stdtypes"
)

type BroadcastRequest struct {
	baseEvent
	Data []byte
}

func NewBroadcastRequest(destModule stdtypes.ModuleID, data []byte) *BroadcastRequest {
	return &BroadcastRequest{
		baseEvent: newBaseEvent(destModule),
		Data:      data,
	}
}

func (e *BroadcastRequest) NewSrc(newSrc stdtypes.ModuleID) stdtypes.Event {
	newEvent := *e
	newEvent.SrcModule = newSrc
	return &newEvent
}

func (e *BroadcastRequest) NewDest(newDest stdtypes.ModuleID) stdtypes.Event {
	newEvent := *e
	newEvent.DestModule = newDest
	return &newEvent
}

func (br *BroadcastRequest) ToBytes() ([]byte, error) {
	return serialize(br)
}

func (br *BroadcastRequest) ToString() string {
	data, err := br.ToBytes()
	if err != nil {
		return fmt.Sprintf("unmarshalableEvent(%+v)", br)
	}

	return string(data)
}

func (e *BroadcastRequest) SetMetadata(key string, value interface{}) (stdtypes.Event, error) {
	newE := *e
	newE.Metadata[key] = value
	return &newE, nil
}
