package testmodules

import (
	"bytes"
	"fmt"
	"slices"

	bcbevents "github.com/filecoin-project/mir/samples/bcb-native/events"
	"github.com/filecoin-project/mir/stdtypes"
	"github.com/google/uuid"

	checkerevents "github.com/filecoin-project/mir/checker/events"
	"github.com/filecoin-project/mir/pkg/dsl"
	"github.com/filecoin-project/mir/pkg/logging"
)

type Integrity struct {
	m            dsl.Module
	systemConfig SystemConfig
	logger       logging.Logger

	byzantineSender bool

	broadcastRequest        []byte
	broadcastDeliverTracker map[stdtypes.NodeID]bool
}

func NewIntegrity(sc SystemConfig, logger logging.Logger) dsl.Module {
	m := dsl.NewModule("validity")

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

	for _, nodeId := range sc.AllNodes {
		v.broadcastDeliverTracker[nodeId] = make(map[uuid.UUID]bool)
	}

	dsl.UponEvent(m, v.handleBroadcastRequest)
	dsl.UponEvent(m, v.handleDeliver)
	dsl.UponEvent(m, v.handleFinal)
	dsl.UponOtherEvent(m, func(_ stdtypes.Event) error { return nil })

	return m
}

func (i *Integrity) handleBroadcastRequest(e *bcbevents.BroadcastRequest) error {
	nodeID := getNodeIdFromMetadata(e)
	if nodeID == i.systemConfig.Sender {
		broadcastRequest = e.Data
	}
	return nil
}

func (i *Integrity) handleDeliver(e *broadcastevents.Deliver) error {
	nodeID := getNodeIdFromMetadata(e)

	if _, ok := i.broadcastDeliverTracker[nodeID]; ok {
		fmt.Printf("Node %s delivered a second value\n", nodeID)
		// fail
		dsl.EmitEvent(i.m, checkerevents.NewFailureEvent())
	} else if i.byzantineSender {
		// must not be delivered before so all ok, just note that we delivered something
		i.boradcastDeliverTracker[nodeId] = true
	} else if i.broadcastRequest != nil {
		fmt.Printf("Node %s delivered a value before anything was broadcasted", nodeID)
		// fail
		dsl.EmitEvent(i.m, checkerevents.NewFailureEvent())
	} else if !bytes.Equal(broadcastRequest, e.Data) {
		fmt.Printf("Node %s delivered a value different from what the honest sender broadcasted", nodeID)
		// fail
		dsl.EmitEvent(i.m, checkerevents.NewFailureEvent())
	} else {
		// all ok, just note that we delivered something
		i.boradcastDeliverTracker[nodeId] = true
	}

	return nil
}

func (i *Integrity) handleFinal(e *checkerevents.FinalEvent) error {
	// if not failed before, we're ok
	dsl.EmitEvent(i.m, checkerevents.NewSuccessEvent())
	return nil
}
