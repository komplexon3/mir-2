package properties

import (
	"bytes"
	"slices"

	bcbevents "github.com/filecoin-project/mir/samples/bcb-native/events"
	"github.com/filecoin-project/mir/stdtypes"

	checkerevents "github.com/filecoin-project/mir/checker/events"
	"github.com/filecoin-project/mir/pkg/dsl"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/util/sliceutil"
)

type Validity struct {
	m            dsl.Module
	systemConfig SystemConfig
	logger       logging.Logger

	byzantineSender bool

	broadcastRequest        []byte
	broadcastDeliverTracker map[stdtypes.NodeID][]byte
}

func NewValidity(sc SystemConfig, logger logging.Logger) dsl.Module {
	m := dsl.NewModule("validity")

	// TODO: setup broadcast
	v := Validity{
		m:            m,
		systemConfig: sc,
		logger:       logger,

		byzantineSender: slices.Contains(sc.ByzantineNodes, sc.Sender),

		broadcastRequest:        nil,
		broadcastDeliverTracker: make(map[stdtypes.NodeID][]byte),
	}

	dsl.UponEvent(m, v.handleBroadcastRequest)
	dsl.UponEvent(m, v.handleDeliver)
	dsl.UponEvent(m, v.handleFinal)
	dsl.UponOtherEvent(m, func(_ stdtypes.Event) error { return nil })

	return m
}

func (v *Validity) handleBroadcastRequest(e *bcbevents.BroadcastRequest) error {
	nodeID := getNodeIdFromMetadata(e)
	if nodeID == v.systemConfig.Sender {
		v.broadcastRequest = e.Data
	}
	return nil
}

func (v *Validity) handleDeliver(e *bcbevents.Deliver) error {
	nodeID := getNodeIdFromMetadata(e)

	v.broadcastDeliverTracker[nodeID] = e.Data
	return nil
}

func (v *Validity) handleFinal(e *checkerevents.FinalEvent) error {
	// if sender byz or if honest sender didn't broadcast, success
	if v.byzantineSender || v.broadcastRequest == nil {
		dsl.EmitEvent(v.m, checkerevents.NewSuccessEvent())
		return nil
	}

	// checking that all nodes delivered the broadcasted value
	nonByzantineNodes := sliceutil.Filter(v.systemConfig.AllNodes, func(_ int, n stdtypes.NodeID) bool {
		return slices.Contains(v.systemConfig.ByzantineNodes, n)
	})

	// only checking non byzantine nodes
	for _, nonByzantineNode := range nonByzantineNodes {
		deliveredValue, ok := v.broadcastDeliverTracker[nonByzantineNode]
		if !ok || !bytes.Equal(v.broadcastRequest, deliveredValue) {
			dsl.EmitEvent(v.m, checkerevents.NewFailureEvent())
		}
	}

	dsl.EmitEvent(v.m, checkerevents.NewSuccessEvent())

	return nil
}
