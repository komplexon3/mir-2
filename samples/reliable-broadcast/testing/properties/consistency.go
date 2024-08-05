package properties

import (
	"bytes"
	"slices"

	"github.com/filecoin-project/mir/samples/reliable-broadcast/events"
	"github.com/filecoin-project/mir/stdtypes"

	checkerevents "github.com/filecoin-project/mir/fuzzer/checker/events"
	"github.com/filecoin-project/mir/pkg/dsl"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/util/sliceutil"
)

// Consistency: If a correct party c-delivers m and another correct party c-delivers m′, then m = m′.

type Consistency struct {
	m                       dsl.Module
	logger                  logging.Logger
	broadcastDeliverTracker map[stdtypes.NodeID][]byte
	systemConfig            SystemConfig
}

func NewConsistency(sc SystemConfig, logger logging.Logger) dsl.Module {
	m := dsl.NewModule("consistency")

	v := Consistency{
		m:            m,
		systemConfig: sc,
		logger:       logger,

		broadcastDeliverTracker: make(map[stdtypes.NodeID][]byte, len(sc.AllNodes)),
	}

	dsl.UponEvent(m, v.handleDeliver)
	dsl.UponEvent(m, v.handleFinal)
	dsl.UponOtherEvent(m, func(_ stdtypes.Event) error { return nil })

	return m
}

func (c *Consistency) handleDeliver(e *events.Deliver) error {
	nodeID := getNodeIdFromMetadata(e)
	c.broadcastDeliverTracker[nodeID] = e.Data

	return nil
}

func (c *Consistency) handleFinal(e *checkerevents.FinalEvent) error {
	nonByzantineNodes := sliceutil.Filter(c.systemConfig.AllNodes, func(_ int, n stdtypes.NodeID) bool {
		return slices.Contains(c.systemConfig.ByzantineNodes, n)
	})

	// -> this would be a different problem...
	if len(nonByzantineNodes) < 2 {
		dsl.EmitEvent(c.m, checkerevents.NewSuccessEvent())
	}

	var ref []byte
	for n, d := range c.broadcastDeliverTracker {
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

	dsl.EmitEvent(c.m, checkerevents.NewSuccessEvent())

	return nil
}
