package heartbeat

import (
	"encoding/json"
	"fmt"
	"maps"

	"github.com/filecoin-project/mir/stdtypes"
	es "github.com/go-errors/errors"
)

const (
	HeartbeatEvent = "Heartbeat"
)

type serializedEvent struct {
	EvType string
	EvData json.RawMessage
}

type Heartbeat struct {
	SrcModule  stdtypes.ModuleID
	DestModule stdtypes.ModuleID
	Metadata   map[string]interface{}
}

func NewHeartbeat(destModule stdtypes.ModuleID) *Heartbeat {
	return &Heartbeat{
		DestModule: destModule,
		Metadata:   make(map[string]interface{}),
	}
}

func (hb *Heartbeat) Src() stdtypes.ModuleID {
	return hb.SrcModule
}

func (hb *Heartbeat) Dest() stdtypes.ModuleID {
	return hb.DestModule
}

func (hb *Heartbeat) GetMetadata(key string) (interface{}, error) {
	val, ok := hb.Metadata[key]
	if !ok {
		return nil, es.Errorf("no metadata for key %s, Metadata: %v", key, hb.Metadata)
	}

	return val, nil
}

func (hb *Heartbeat) NewSrc(newSrc stdtypes.ModuleID) stdtypes.Event {
	newEvent := *hb
	newEvent.SrcModule = newSrc
	return &newEvent
}

func (hb *Heartbeat) NewDest(newDest stdtypes.ModuleID) stdtypes.Event {
	newEvent := *hb
	newEvent.DestModule = newDest
	return &newEvent
}

func (hb *Heartbeat) ToBytes() ([]byte, error) {
	evtData, err := json.Marshal(hb)
	if err != nil {
		return nil, es.Errorf("could not marshal message: %w", err)
	}

	se := serializedEvent{
		EvType: HeartbeatEvent,
		EvData: evtData,
	}

	data, err := json.Marshal(se)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (hb *Heartbeat) ToString() string {
	data, err := hb.ToBytes()
	if err != nil {
		return fmt.Sprintf("unmarshalableEvent(%+v)", hb)
	}

	return string(data)
}

func (hb *Heartbeat) SetMetadata(key string, value interface{}) (stdtypes.Event, error) {
	newE := *hb
	metadata := maps.Clone(newE.Metadata)
	metadata[key] = value
	newE.Metadata = metadata
	return &newE, nil
}

func Deserialize(data []byte) (stdtypes.Event, error) {
	var se serializedEvent
	err := json.Unmarshal(data, &se)
	if err != nil {
		return nil, err
	}

	if se.EvType != HeartbeatEvent {
		return nil, es.Errorf("unknown event type: %s", se.EvType)
	}

	hb := &Heartbeat{Metadata: make(map[string]interface{})}
	if err := json.Unmarshal(se.EvData, hb); err != nil {
		return nil, err
	}

	return hb, nil
}
