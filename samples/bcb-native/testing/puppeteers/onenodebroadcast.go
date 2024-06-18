package puppeteers

import (
	"context"
	"fmt"
	"time"

	"github.com/filecoin-project/mir/adversary"
	bcbevents "github.com/filecoin-project/mir/samples/bcb-native/events"
	"github.com/filecoin-project/mir/stdtypes"
)

type OneNodeBroadcast struct {
	node stdtypes.NodeID
}

func NewOneNodeBroadcast(nodeId stdtypes.NodeID) *OneNodeBroadcast {
	return &OneNodeBroadcast{
		nodeId,
	}
}

func (onb *OneNodeBroadcast) Run(nodeInstances map[stdtypes.NodeID]adversary.NodeInstance) error {
	ctx := context.Background()
	time.Sleep(time.Second)
	nodeInstance := nodeInstances[onb.node]
	node := nodeInstance.GetNode()
	msg := fmt.Sprintf("node %s injecting", node.ID)
	node.InjectEvents(ctx, stdtypes.ListOf(bcbevents.NewBroadcastRequest("broadcast", []byte(msg))))

	return nil
}
