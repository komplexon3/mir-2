package puppeteers

import (
	"context"
	"fmt"
	"time"

	"github.com/filecoin-project/mir/adversary"
	broadcastevents "github.com/filecoin-project/mir/samples/broadcast/events"
	"github.com/filecoin-project/mir/stdtypes"
	es "github.com/go-errors/errors"
	"github.com/google/uuid"
)

type RoundRobinPuppeteer struct {
	repetitions int
	delay       time.Duration
}

func NewRoundRobinPuppeteer(repetitions int, delay time.Duration) (*RoundRobinPuppeteer, error) {
	if repetitions < 1 {
		return nil, es.Errorf("repetitions must be >= 1")
	}
	return &RoundRobinPuppeteer{
		repetitions,
		delay,
	}, nil
}

func (rrp *RoundRobinPuppeteer) Run(nodeInstances []adversary.NodeInstance) error {
	ctx := context.Background()
	for round := 0; round < rrp.repetitions; round++ {
		for _, nodeInstance := range nodeInstances {
			node := nodeInstance.GetNode()
			msg := fmt.Sprintf("node %s - round %d", node.ID, round)
			node.InjectEvents(ctx, stdtypes.ListOf(broadcastevents.NewBroadcastRequest("broadcast", []byte(msg), uuid.New())))
			time.Sleep(rrp.delay)
		}
	}

	return nil
}
