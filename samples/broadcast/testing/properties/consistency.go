package testmodules

import (
	"bytes"
	"slices"

	broadcastevents "github.com/filecoin-project/mir/samples/broadcast/events"
	"github.com/filecoin-project/mir/stdtypes"
	"github.com/google/uuid"

	checkerevents "github.com/filecoin-project/mir/checker/events"
	"github.com/filecoin-project/mir/pkg/dsl"
	"github.com/filecoin-project/mir/pkg/logging"
)

type Consistency struct {
	m            dsl.Module
	systemConfig SystemConfig
	logger       logging.Logger

	broadcastDeliverTracker map[uuid.UUID]map[stdtypes.NodeID][]byte
}

func NewConsistency(sc SystemConfig, logger logging.Logger) dsl.Module {
	m := dsl.NewModule("validity")

	// TODO: setup broadcast
	v := Consistency{
		m:            m,
		systemConfig: sc,
		logger:       logger,

		broadcastDeliverTracker: make(map[uuid.UUID]map[stdtypes.NodeID][]byte, len(sc.AllNodes)),
	}

	dsl.UponEvent(m, v.handleDeliver)
	dsl.UponEvent(m, v.handleFinal)
	dsl.UponOtherEvent(m, func(_ stdtypes.Event) error { return nil })

	return m
}

func (c *Consistency) handleDeliver(e *broadcastevents.Deliver) error {
	nodeID := getNodeIdFromMetadata(e)
	if _, ok := c.broadcastDeliverTracker[e.BroadcastID]; !ok {
		c.broadcastDeliverTracker[e.BroadcastID] = make(map[stdtypes.NodeID][]byte)
	}
	c.broadcastDeliverTracker[e.BroadcastID][nodeID] = e.Data

	return nil
}

func (c *Consistency) handleFinal(e *checkerevents.FinalEvent) error {
	// first non-byzantine node
	nonByzantineNodes := slices.DeleteFunc(c.systemConfig.AllNodes, func(n stdtypes.NodeID) bool {
		return slices.Contains(c.systemConfig.ByzantineNodes, n)
	})

	// -> this would be a different problem...
	if len(nonByzantineNodes) < 2 {
		dsl.EmitEvent(c.m, checkerevents.NewSuccessEvent())
	}

	for _, b := range c.broadcastDeliverTracker {
		var ref []byte
		for n, d := range b {
			if slices.Contains(c.systemConfig.ByzantineNodes, n) {
				continue // byzantine node, we don't care...
			}
			if ref == nil {
				ref = d
				continue
			}

			if !bytes.Equal(ref, d) {
				dsl.EmitEvent(c.m, checkerevents.NewFailureEvent())
				return nil
			}
		}
	}

	dsl.EmitEvent(c.m, checkerevents.NewSuccessEvent())

	return nil
}
