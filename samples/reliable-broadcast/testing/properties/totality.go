package properties

import (
	"bytes"
	"slices"

	events "github.com/filecoin-project/mir/samples/bcb-native/events"
	"github.com/filecoin-project/mir/stdtypes"

	checkerevents "github.com/filecoin-project/mir/fuzzer/checker/events"
	"github.com/filecoin-project/mir/pkg/dsl"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/util/sliceutil"
)

// Totality: If some correct party r-delivers a request, then all correct parties eventually r-deliver a request.

type Totality struct {
	m                       dsl.Module
	logger                  logging.Logger
	broadcastDeliverTracker map[stdtypes.NodeID]bool
	systemConfig            SystemConfig
	firstDeliveredValue     []byte
}

func NewTotality(sc SystemConfig, logger logging.Logger) dsl.Module {
	m := dsl.NewModule("totality")

	v := Totality{
		m:            m,
		systemConfig: sc,
		logger:       logger,

		firstDeliveredValue:     nil,
		broadcastDeliverTracker: make(map[stdtypes.NodeID]bool), // used as a set
	}

	dsl.UponEvent(m, v.handleBroadcastRequest)
	dsl.UponEvent(m, v.handleDeliver)
	dsl.UponEvent(m, v.handleFinal)
	dsl.UponOtherEvent(m, func(_ stdtypes.Event) error { return nil })

	return m
}

func (v *Totality) handleBroadcastRequest(e *events.BroadcastRequest) error {
	return nil
}

func (v *Totality) handleDeliver(e *events.Deliver) error {
	nodeID := getNodeIdFromMetadata(e)

	if v.firstDeliveredValue == nil {
		v.firstDeliveredValue = e.Data
		v.broadcastDeliverTracker[nodeID] = true
	} else if bytes.Equal(v.firstDeliveredValue, e.Data) {
		v.broadcastDeliverTracker[nodeID] = true
	}

	return nil
}

func (v *Totality) handleFinal(e *checkerevents.FinalEvent) error {
	// if no value was delivered, all ok
	if v.firstDeliveredValue == nil {
		dsl.EmitEvent(v.m, checkerevents.NewSuccessEvent())
		return nil
	}

	// checking that all correct nodes delivered the broadcasted value
	nonByzantineNodes := sliceutil.Filter(v.systemConfig.AllNodes, func(_ int, n stdtypes.NodeID) bool {
		return slices.Contains(v.systemConfig.ByzantineNodes, n)
	})

	// only checking non byzantine nodes
	for _, nonByzantineNode := range nonByzantineNodes {
		_, ok := v.broadcastDeliverTracker[nonByzantineNode]
		if !ok {
			dsl.EmitEvent(v.m, checkerevents.NewFailureEvent())
		}
	}

	dsl.EmitEvent(v.m, checkerevents.NewSuccessEvent())

	return nil
}
