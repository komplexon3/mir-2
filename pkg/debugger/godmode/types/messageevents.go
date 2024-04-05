package types

import (
	"encoding/json"
	"time"

	"github.com/filecoin-project/mir/stdtypes"
)

type MessageEventType string

const (
	TypeMesssageBuffered  MessageEventType = "TypeNewMessage"
	TypeMessageDropped    MessageEventType = "TypeMessageDropped"
	TypeMessageDispatched MessageEventType = "TypeMessageDispatched"
)

type MessageEvent struct {
	Type      MessageEventType `json:"Type"`
	Timestamp int64            `json:"Timestamp"`
	Message   stdtypes.Message `json:"Message"`
}

func NewMessageEvent(m stdtypes.Message, updateType MessageEventType, timestamp time.Time) MessageEvent {
	return MessageEvent{
		Type:      updateType,
		Timestamp: timestamp.UnixMilli(), // TODO: check what unit is appropriate
		Message:   m,
	}
}

func (iu *MessageEvent) Marshal() ([]byte, error) {
	return json.Marshal(iu)
}
