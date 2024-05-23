package puppeteers

import (
	"context"
	"fmt"
	"time"

	"github.com/filecoin-project/mir/adversary"
	broadcastevents "github.com/filecoin-project/mir/samples/broadcast/events"
	"github.com/filecoin-project/mir/stdtypes"
	"github.com/google/uuid"
)

type OneNodeBroadcast struct {
	node stdtypes.NodeID
}

func NewOneNodeBroadcast(nodeId stdtypes.NodeID) (*OneNodeBroadcast, error) {
	return &OneNodeBroadcast{
		nodeId,
	}, nil
}

func (onb *OneNodeBroadcast) Run(nodeInstances map[stdtypes.NodeID]adversary.NodeInstance) error {
	ctx := context.Background()
  time.Sleep(time.Second)
	nodeInstance := nodeInstances[onb.node]
	node := nodeInstance.GetNode()
	msg := fmt.Sprintf("node %s injecting", node.ID)
	node.InjectEvents(ctx, stdtypes.ListOf(broadcastevents.NewBroadcastRequest("broadcast", []byte(msg), uuid.New())))

	return nil
}
