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

type Validity struct {
	m            dsl.Module
	systemConfig SystemConfig
	logger       logging.Logger

	broadcastRequests       map[uuid.UUID]broadcastRequest
	broadcastDeliverTracker map[uuid.UUID]map[stdtypes.NodeID][]byte
}

func NewValidity(sc SystemConfig, logger logging.Logger) dsl.Module {
	m := dsl.NewModule("validity")

	// TODO: setup broadcast
	v := Validity{
		m:            m,
		systemConfig: sc,
		logger:       logger,

		broadcastRequests:       make(map[uuid.UUID]broadcastRequest),
		broadcastDeliverTracker: make(map[uuid.UUID]map[stdtypes.NodeID][]byte),
	}

	dsl.UponEvent(m, v.handleBroadcastRequest)
	dsl.UponEvent(m, v.handleDeliver)
	dsl.UponEvent(m, v.handleFinal)
	dsl.UponOtherEvent(m, func(_ stdtypes.Event) error { return nil })

	return m
}

func (v *Validity) handleBroadcastRequest(e *broadcastevents.BroadcastRequest) error {
	nodeID := getNodeIdFromMetadata(e)
	v.broadcastRequests[e.BroadcastID] = broadcastRequest{data: e.Data, node: nodeID}
	return nil
}

func (v *Validity) handleDeliver(e *broadcastevents.Deliver) error {
	nodeID := getNodeIdFromMetadata(e)

	if _, ok := v.broadcastDeliverTracker[e.BroadcastID]; !ok {
		v.broadcastDeliverTracker[e.BroadcastID] = make(map[stdtypes.NodeID][]byte)
	}
	v.broadcastDeliverTracker[e.BroadcastID][nodeID] = e.Data
	return nil
}

func (v *Validity) handleFinal(e *checkerevents.FinalEvent) error {
	// checking that all nodes delivered the broadcasted value
	nonByzantineNodes := slices.DeleteFunc(v.systemConfig.AllNodes, func(n stdtypes.NodeID) bool {
		return slices.Contains(v.systemConfig.ByzantineNodes, n)
	})

	for rbi, rbbr := range v.broadcastRequests {
		// ignore byzantine nodes
		if slices.Contains(v.systemConfig.ByzantineNodes, rbbr.node) {
			continue
		}

		bd, ok := v.broadcastDeliverTracker[rbi]
		if !ok {
			dsl.EmitEvent(v.m, checkerevents.NewFailureEvent())
		}

		// only checking non byzantine nodes
		for _, nonByzantineNode := range nonByzantineNodes {
			deliveredValue, ok := bd[nonByzantineNode]
			if !ok || !bytes.Equal(rbbr.data, deliveredValue) {
				dsl.EmitEvent(v.m, checkerevents.NewFailureEvent())
			}
		}
	}

	dsl.EmitEvent(v.m, checkerevents.NewSuccessEvent())

	return nil
}
