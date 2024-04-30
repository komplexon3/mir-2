package testmodules

import (
	"bytes"
	"fmt"
	"slices"

	broadcastevents "github.com/filecoin-project/mir/samples/broadcast/events"
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

	broadcastRequests       map[uuid.UUID]broadcastRequest
	broadcastDeliverTracker map[stdtypes.NodeID]map[uuid.UUID]bool
}

func NewIntegrity(sc SystemConfig, logger logging.Logger) dsl.Module {
	m := dsl.NewModule("validity")

	// TODO: setup broadcast
	v := Integrity{
		m:            m,
		systemConfig: sc,
		logger:       logger,

		broadcastRequests:       make(map[uuid.UUID]broadcastRequest),
		broadcastDeliverTracker: make(map[stdtypes.NodeID]map[uuid.UUID]bool, len(sc.AllNodes)),
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

func (i *Integrity) handleBroadcastRequest(e *broadcastevents.BroadcastRequest) error {
	nodeID := getNodeIdFromMetadata(e)
	i.broadcastRequests[e.BroadcastID] = broadcastRequest{e.Data, nodeID}
	return nil
}

func (i *Integrity) handleDeliver(e *broadcastevents.Deliver) error {
	nodeID := getNodeIdFromMetadata(e)

	if _, ok := i.broadcastDeliverTracker[nodeID][e.BroadcastID]; ok {
		fmt.Printf("Node %s delivered a second value\n", nodeID)
		// fail
		dsl.EmitEvent(i.m, checkerevents.NewFailureEvent())
	} else if !slices.Contains(i.systemConfig.ByzantineNodes, e.BroadcastSender) { // broadcastsender can be spoofed!
		br, ok := i.broadcastRequests[e.BroadcastID]
		if !ok {
			fmt.Println("Delivered without previous request")
			// fail
			dsl.EmitEvent(i.m, checkerevents.NewFailureEvent())
		} else if !bytes.Equal(br.data, e.Data) {
			fmt.Printf("Node %s delivered a different value than was requested\n", nodeID)
			// fail
			dsl.EmitEvent(i.m, checkerevents.NewFailureEvent())
		}
	}
	i.broadcastDeliverTracker[nodeID][e.BroadcastID] = true

	return nil
}

func (i *Integrity) handleFinal(e *checkerevents.FinalEvent) error {
	// if not failed before, we're ok
	dsl.EmitEvent(i.m, checkerevents.NewSuccessEvent())
	return nil
}
