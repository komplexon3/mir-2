package messages

import (
	"encoding/json"
	"fmt"

	es "github.com/go-errors/errors"

	"github.com/filecoin-project/mir/stdtypes"
)

// based on stdevents/serialization.go
type messageType string

const (
	StartMessageMsg = "StartMessage"
	EchoMessageMsg  = "EchoMessage"
	ReadyMessageMsg = "ReadyMessage"
)

type serializedMessage struct {
	MsgType messageType
	MsgData json.RawMessage
}

func serialize(e any) ([]byte, error) {

	var msgType messageType
	switch e.(type) {
	case *StartMessage:
		msgType = StartMessageMsg
	case *EchoMessage:
		msgType = EchoMessageMsg
	case *ReadyMessage:
		msgType = ReadyMessageMsg
	default:
		return nil, es.Errorf("unknown message type: %T", e)
	}

	msgData, err := json.Marshal(e)
	if err != nil {
		return nil, es.Errorf("could not marshal message: %w", err)
	}

	se := serializedMessage{
		MsgType: msgType,
		MsgData: msgData,
	}

	data, err := json.Marshal(se)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func Deserialize(data []byte) (stdtypes.Message, error) {
	var sm serializedMessage
	var e stdtypes.Message
	err := json.Unmarshal(data, &sm)
	if err != nil {
		return nil, err
	}
	switch sm.MsgType {
	case StartMessageMsg:
		e = &StartMessage{}
	case EchoMessageMsg:
		e = &EchoMessage{}
	case ReadyMessageMsg:
		e = &ReadyMessage{}
	default:
		return nil, fmt.Errorf("unknown message type: %s", sm.MsgType)
	}

	if err := json.Unmarshal(sm.MsgData, e); err != nil {
		return nil, err
	}
	return e, nil
}
