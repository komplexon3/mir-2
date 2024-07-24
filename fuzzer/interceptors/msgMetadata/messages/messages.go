package messages

import (
	"encoding/json"

	es "github.com/go-errors/errors"
	"google.golang.org/protobuf/proto"

	"github.com/filecoin-project/mir/pkg/pb/messagepb"
	"github.com/filecoin-project/mir/stdtypes"
)

type metadataMessageType string

const (
	MetadataNativeMessageType metadataMessageType = "MetadataNativeMessage"
	MetadataPbMessageType     metadataMessageType = "MetadataPbMessage"
)

type MetadataMessage interface {
	ToBytes() ([]byte, error)
	GetMetadata(key string) (interface{}, error)
}

type baseMetadataMessage struct {
	Metadata map[string]interface{}
}

type serializedMessage struct {
	Type     metadataMessageType
	Payload  []byte
	Metadata []byte
}

func (m baseMetadataMessage) GetMetadata(key string) (interface{}, error) {
	val, ok := m.Metadata[key]
	if !ok {
		return nil, es.Errorf("value for key %s not found in metadata", key)
	}

	return val, nil
}

type MetadataNativeMessage struct {
	baseMetadataMessage
	Payload stdtypes.Message
}

func (m MetadataNativeMessage) ToBytes() ([]byte, error) {
	serializedPayload, err := m.Payload.ToBytes()
	if err != nil {
		return nil, err
	}
	serializedMetadata, err := json.Marshal(m.Metadata)
	if err != nil {
		return nil, err
	}

	sm := serializedMessage{
		Type:     MetadataNativeMessageType,
		Payload:  serializedPayload,
		Metadata: serializedMetadata,
	}
	return json.Marshal(sm)
}

func (m MetadataNativeMessage) UnwrapMessage() (stdtypes.Message, map[string]interface{}) {
	return m.Payload, m.Metadata
}

func WrapNativeMessage(msg stdtypes.Message, metadata map[string]interface{}) MetadataMessage {
	return MetadataNativeMessage{
		baseMetadataMessage: baseMetadataMessage{
			Metadata: metadata,
		},
		Payload: msg,
	}
}

func (m MetadataNativeMessage) GetMessage() stdtypes.Message {
	return m.Payload
}

type MetadataPbMessage struct {
	baseMetadataMessage
	Payload *messagepb.Message
}

func (m MetadataPbMessage) ToBytes() ([]byte, error) {
	serializedPayload, err := proto.Marshal(m.Payload)
	if err != nil {
		return nil, err
	}

	serializedMetadata, err := json.Marshal(m.Metadata)
	if err != nil {
		return nil, err
	}

	sm := serializedMessage{
		Type:     MetadataPbMessageType,
		Payload:  serializedPayload,
		Metadata: serializedMetadata,
	}
	return json.Marshal(sm)
}

func (m MetadataPbMessage) UnwrapMessage() (*messagepb.Message, map[string]interface{}) {
	return m.Payload, m.Metadata
}

func WrapPbMessage(msg *messagepb.Message, metadata map[string]interface{}) MetadataMessage {
	return MetadataPbMessage{
		baseMetadataMessage: baseMetadataMessage{
			Metadata: metadata,
		},
		Payload: msg,
	}
}

func (m MetadataPbMessage) GetMessage() *messagepb.Message {
	return m.Payload
}

func Deserialize(data []byte) (MetadataMessage, error) {
	var sm serializedMessage
	var metadata map[string]interface{}

	err := json.Unmarshal(data, &sm)
	if err != nil {
		return nil, err
	}

	if sm.Metadata != nil {
		err = json.Unmarshal(sm.Metadata, &metadata)
		if err != nil {
			return nil, err
		}
	}

	switch sm.Type {
	case MetadataNativeMessageType:
		payload := stdtypes.RawMessage(sm.Payload)
		return &MetadataNativeMessage{
			baseMetadataMessage: baseMetadataMessage{
				Metadata: metadata,
			},
			Payload: payload,
		}, nil

	case MetadataPbMessageType:
		var pbPayload messagepb.Message
		err = proto.Unmarshal(sm.Payload, &pbPayload)
		if err != nil {
			return nil, err
		}
		return &MetadataPbMessage{
			baseMetadataMessage: baseMetadataMessage{
				Metadata: metadata,
			},
			Payload: &pbPayload,
		}, nil
	default:
		return nil, es.Errorf("message %v is not a metadata message", sm)
	}
}
