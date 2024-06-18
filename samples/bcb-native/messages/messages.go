package messages

import (
	"github.com/filecoin-project/mir/stdtypes"
)

type StartMessage struct {
	Data []byte
}

func (m *StartMessage) ToBytes() ([]byte, error) {
	return serialize(m)
}

func NewStartMessage(data []byte) *StartMessage {
	return &StartMessage{
		Data: data,
	}
}

type EchoMessage struct {
	Signature []byte
}

func (m *EchoMessage) ToBytes() ([]byte, error) {
	return serialize(m)
}

func NewEchoMessage(signature []byte) *EchoMessage {
	return &EchoMessage{
		Signature: signature,
	}
}

type FinalMessage struct {
	Data       []byte
	Signers    []stdtypes.NodeID
	Signatures [][]byte
}

func (m *FinalMessage) ToBytes() ([]byte, error) {
	return serialize(m)
}

func NewFinalMessage(data []byte, signers []stdtypes.NodeID, signatures [][]byte) *FinalMessage {
	return &FinalMessage{
		Data:       data,
		Signers:    signers,
		Signatures: signatures,
	}
}
