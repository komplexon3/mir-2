package events

import (
	"encoding/json"

	es "github.com/go-errors/errors"

	"github.com/filecoin-project/mir/stdtypes"
)

// based on stdevents/serialization.go
type eventType string

const (
	PingPongEvent = "PingPong"
)

type serializedEvent struct {
	EvType eventType
	EvData json.RawMessage
}

func serialize(e any) ([]byte, error) {

	var evType eventType
	switch e.(type) {
	case *PingPong:
		evType = PingPongEvent
	default:
		return nil, es.Errorf("unknown message type: %T", e)
	}

	evtData, err := json.Marshal(e)
	if err != nil {
		return nil, es.Errorf("could not marshal message: %w", err)
	}

	se := serializedEvent{
		EvType: evType,
		EvData: evtData,
	}

	data, err := json.Marshal(se)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func Deserialize(data []byte) (stdtypes.Event, error) {
	var se serializedEvent
	var e stdtypes.Event
	err := json.Unmarshal(data, &se)
	if err != nil {
		return nil, err
	}

	switch se.EvType {
	case PingPongEvent:
		e = &PingPong{baseEvent: baseEvent{Metadata: make(map[string]interface{})}}
	default:
		return nil, es.Errorf("unknown event type: %s", se.EvType)
	}

	if err := json.Unmarshal(se.EvData, e); err != nil {
		return nil, err
	}

	return e, nil
}
