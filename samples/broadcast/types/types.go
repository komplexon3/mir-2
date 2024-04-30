package types

import (
	"github.com/filecoin-project/mir/stdtypes"
	"github.com/google/uuid"
)

type BroadcastPayload struct {
	Data            []byte
	BroadcastSender stdtypes.NodeID
	BroadcastID     uuid.UUID
}
