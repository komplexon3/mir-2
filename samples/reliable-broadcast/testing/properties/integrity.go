package properties

import (
	"bytes"
	"fmt"
	"slices"

	bcbevents "github.com/filecoin-project/mir/samples/reliable-broadcast/events"
	"github.com/filecoin-project/mir/stdtypes"

	checkerevents "github.com/filecoin-project/mir/fuzzer/checker/events"
	"github.com/filecoin-project/mir/pkg/dsl"
	"github.com/filecoin-project/mir/pkg/logging"
)

// Integrity: Every correct party c-delivers at most one request. Moreover, if the sender Ps is correct, then the request was previously c-broadcast by Ps.

type Integrity struct {
	m                       dsl.Module
	logger                  logging.Logger
	broadcastDeliverTracker map[stdtypes.NodeID]bool
	systemConfig            SystemConfig
	broadcastRequest        []byte
	byzantineSender         bool
}

func NewIntegrity(sc SystemConfig, logger logging.Logger) dsl.Module {
	m := dsl.NewModule("integrity")

	byzSender := slices.Contains(sc.ByzantineNodes, sc.Sender)

	// TODO: setup broadcast
	v := Integrity{
		m:            m,
		systemConfig: sc,
		logger:       logger,

		byzantineSender: byzSender,

		broadcastRequest:        nil,
		broadcastDeliverTracker: make(map[stdtypes.NodeID]bool, len(sc.AllNodes)),
	}

	v.broadcastDeliverTracker = make(map[stdtypes.NodeID]bool)

	dsl.UponEvent(m, v.handleBroadcastRequest)
	dsl.UponEvent(m, v.handleDeliver)
	dsl.UponEvent(m, v.handleFinal)
	dsl.UponOtherEvent(m, func(_ stdtypes.Event) error { return nil })

	return m
}

func (i *Integrity) handleBroadcastRequest(e *bcbevents.BroadcastRequest) error {
	nodeID := getNodeIdFromMetadata(e)
	if nodeID == i.systemConfig.Sender {
		i.broadcastRequest = e.Data
	}
	return nil
}

func (i *Integrity) handleDeliver(e *bcbevents.Deliver) error {
	nodeID := getNodeIdFromMetadata(e)

	if slices.Contains(i.systemConfig.ByzantineNodes, nodeID) {
		// byzantine node, ignore
	} else if _, ok := i.broadcastDeliverTracker[nodeID]; ok {
		// fmt.Printf("Node %s delivered a second value\n", nodeID)
		// fail
		dsl.EmitEvent(i.m, checkerevents.NewFailureEvent())
	} else if i.byzantineSender {
		// must not be delivered before so all ok, just note that we delivered something
		i.broadcastDeliverTracker[nodeID] = true
	} else if i.broadcastRequest == nil {
		fmt.Printf("Node %s delivered a value before anything was broadcasted", nodeID)
		// fail
		dsl.EmitEvent(i.m, checkerevents.NewFailureEvent())
	} else if !bytes.Equal(i.broadcastRequest, e.Data) {
		fmt.Printf("Node %s delivered a value different from what the honest sender broadcasted", nodeID)
		// fail
		dsl.EmitEvent(i.m, checkerevents.NewFailureEvent())
	} else {
		// all ok, just note that we delivered something
		i.broadcastDeliverTracker[nodeID] = true
	}

	return nil
}

func (i *Integrity) handleFinal(e *checkerevents.FinalEvent) error {
	// if not failed before, we're ok
	dsl.EmitEvent(i.m, checkerevents.NewSuccessEvent())
	return nil
}
