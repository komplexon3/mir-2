package messages

import (
	"github.com/filecoin-project/mir/samples/broadcast/types"
	"github.com/filecoin-project/mir/stdtypes"
	"github.com/google/uuid"
)

type StartMessage struct {
	types.BroadcastPayload
}

func (m *StartMessage) ToBytes() ([]byte, error) {
	return serialize(m)
}

func NewStartMessage(data []byte, broadcastSender stdtypes.NodeID, broadcastID uuid.UUID) *StartMessage {
	return &StartMessage{
		BroadcastPayload: types.BroadcastPayload{
			Data:            data,
			BroadcastSender: broadcastSender,
			BroadcastID:     broadcastID,
		},
	}
}

type EchoMessage struct {
	Signature   []byte
	BroadcastID uuid.UUID
}

func (m *EchoMessage) ToBytes() ([]byte, error) {
	return serialize(m)
}

func NewEchoMessage(signature []byte, broadcastID uuid.UUID) *EchoMessage {
	return &EchoMessage{
		BroadcastID: broadcastID,
		Signature:   signature,
	}
}

type FinalMessage struct {
	types.BroadcastPayload
	Signers    []stdtypes.NodeID
	Signatures [][]byte
}

func (m *FinalMessage) ToBytes() ([]byte, error) {
	return serialize(m)
}

func NewFinalMessage(data []byte, broadcastSender stdtypes.NodeID, broadcastID uuid.UUID, signers []stdtypes.NodeID, signatures [][]byte) *FinalMessage {
	return &FinalMessage{
		BroadcastPayload: types.BroadcastPayload{
			Data:            data,            // TODO: actually not needed with current instance concept
			BroadcastSender: broadcastSender, // will we actually need this?
			BroadcastID:     broadcastID,
		},
		Signers:    signers,
		Signatures: signatures,
	}
}
