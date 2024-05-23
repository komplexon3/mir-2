package messages

import (
	"encoding/json"

	es "github.com/go-errors/errors"
	"google.golang.org/protobuf/proto"

	"github.com/filecoin-project/mir/pkg/pb/messagepb"
	"github.com/filecoin-project/mir/pkg/vcinterceptor/vectorclock"
	"github.com/filecoin-project/mir/stdtypes"
)

type vcMessageType string

const (
	VCNativeMessageType vcMessageType = "VCNativeMessage"
	VCPbMessageType     vcMessageType = "VCPbMessage"
)

type VCM interface {
	ToBytes() ([]byte, error)
}

type serializedMessage struct {
	Type        vcMessageType
	Payload     []byte
	VectorClock []byte
}

type VCNativeMessage struct {
	Payload     stdtypes.Message
	VectorClock vectorclock.VectorClock
}

func (m VCNativeMessage) ToBytes() ([]byte, error) {
	serializedPayload, err := m.Payload.ToBytes()
	if err != nil {
		return nil, err
	}
	serializedVC, err := json.Marshal(m.VectorClock)
	if err != nil {
		return nil, err
	}

	sm := serializedMessage{
		Type:        VCNativeMessageType,
		Payload:     serializedPayload,
		VectorClock: serializedVC,
	}
	return json.Marshal(sm)
}

func (m VCNativeMessage) UnwrapMessage() (stdtypes.Message, vectorclock.VectorClock) {
	return m.Payload, m.VectorClock
}

func WrapNativeMessage(msg stdtypes.Message, vc vectorclock.VectorClock) VCM {
	return VCNativeMessage{
		Payload:     msg,
		VectorClock: vc,
	}
}

type VCPbMessage struct {
	Payload     *messagepb.Message
	VectorClock vectorclock.VectorClock
}

func (m VCPbMessage) ToBytes() ([]byte, error) {
	serializedPayload, err := proto.Marshal(m.Payload)
	if err != nil {
		return nil, err
	}
	serializedVC, err := json.Marshal(m.VectorClock)
	if err != nil {
		return nil, err
	}

	sm := serializedMessage{
		Type:        VCPbMessageType,
		Payload:     serializedPayload,
		VectorClock: serializedVC,
	}
	return json.Marshal(sm)
}

func (m VCPbMessage) UnwrapMessage() (*messagepb.Message, vectorclock.VectorClock) {
	return m.Payload, m.VectorClock
}

func WrapPbMessage(msg *messagepb.Message, vc vectorclock.VectorClock) VCM {
	return VCPbMessage{
		Payload:     msg,
		VectorClock: vc,
	}
}

func Deserialize(data []byte) (VCM, error) {
	var sm serializedMessage
	var vc vectorclock.VectorClock

	err := json.Unmarshal(data, &sm)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(sm.VectorClock, &vc)
	if err != nil {
		return nil, err
	}

	switch sm.Type {
	case VCNativeMessageType:
    payload := stdtypes.RawMessage(sm.Payload)
		if err != nil {
			return nil, err
		}
		return &VCNativeMessage{
			Payload:     payload,
			VectorClock: vc,
		}, nil

	case VCPbMessageType:
		var pbPayload messagepb.Message
		err = proto.Unmarshal(sm.Payload, &pbPayload)
		if err != nil {
			return nil, err
		}
		return &VCPbMessage{
			Payload:     &pbPayload,
			VectorClock: vc,
		}, nil
	}
	return nil, es.Errorf("message %v is not a vector clock message", sm)
}
