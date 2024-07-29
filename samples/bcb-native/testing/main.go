package main

import (
	"context"
	"fmt"

	es "github.com/go-errors/errors"

	"github.com/filecoin-project/mir/fuzzer"
	"github.com/filecoin-project/mir/fuzzer/actions"

	"github.com/filecoin-project/mir/fuzzer/checker"
	"github.com/filecoin-project/mir/pkg/deploytest"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/trantor/types"
	broadcastevents "github.com/filecoin-project/mir/samples/bcb-native/events"
	"github.com/filecoin-project/mir/samples/bcb-native/testing/nodeinstance"
	"github.com/filecoin-project/mir/samples/bcb-native/testing/properties"
	"github.com/filecoin-project/mir/stdtypes"
)

const (
	MAX_EVENTS = 200
)

var puppeteerEvents = []actions.DelayedEvents{
	{
		NodeID: stdtypes.NodeID("0"),
		Events: stdtypes.ListOf(broadcastevents.NewBroadcastRequest("bcb", []byte("hello"))),
	},
}

var weightedActionsForNetwork = []actions.WeightedAction{
	actions.NewWeightedAction(func(e stdtypes.Event, sourceNode stdtypes.NodeID, byzantineNodes []stdtypes.NodeID) (string, map[stdtypes.NodeID]*stdtypes.EventList, []actions.DelayedEvents, error) {
		return "noop (network)",
			map[stdtypes.NodeID]*stdtypes.EventList{
				sourceNode: stdtypes.ListOf(e),
			},
			nil,
			nil
	}, 10),
	// actions.NewWeightedAction(func(e stdtypes.Event, sourceNode stdtypes.NodeID, byzantineNodes []stdtypes.NodeID) (string, map[stdtypes.NodeID]*stdtypes.EventList, []actions.DelayedEvents, error) {
	// 	return "event delayed (network)",
	// 		nil,
	// 		[]actions.DelayedEvents{
	// 			{
	// 				NodeID: sourceNode,
	// 				Events: stdtypes.ListOf(e),
	// 			},
	// 		},
	// 		nil
	// }, 1),
}

var weightedActionsForByzantineNodes = []actions.WeightedAction{
	actions.NewWeightedAction(func(e stdtypes.Event, sourceNode stdtypes.NodeID, byzantineNodes []stdtypes.NodeID) (string, map[stdtypes.NodeID]*stdtypes.EventList, []actions.DelayedEvents, error) {
		return "noop",
			map[stdtypes.NodeID]*stdtypes.EventList{
				sourceNode: stdtypes.ListOf(e),
			},
			nil,
			nil
	}, 10),
	// actions.NewWeightedAction(func(e stdtypes.Event, sourceNode stdtypes.NodeID, byzantineNodes []stdtypes.NodeID) (string, map[stdtypes.NodeID]*stdtypes.EventList, []actions.DelayedEvents, error) {
	// 	return fmt.Sprintf("dropped event %s", e.ToString()),
	// 		nil,
	// 		nil,
	// 		nil
	// }, 1),
	// actions.NewWeightedAction(func(e stdtypes.Event, sourceNode stdtypes.NodeID, byzantineNodes []stdtypes.NodeID) (string, map[stdtypes.NodeID]*stdtypes.EventList, []actions.DelayedEvents, error) {
	// 	e2, err := e.SetMetadata("duplicated", true)
	// 	if err != nil {
	// 		// TODO: should a failed action just be a "noop"
	// 		return "", nil, nil, err
	// 	}
	// 	return fmt.Sprintf("duplicated event %v", e.ToString()),
	// 		map[stdtypes.NodeID]*stdtypes.EventList{
	// 			sourceNode: stdtypes.ListOf(e, e2),
	// 		},
	// 		nil,
	// 		nil
	// }, 1),
}

func fuzzBCB(
	name string,
	nodes []stdtypes.NodeID,
	byzantineNodes []stdtypes.NodeID,
	sender stdtypes.NodeID,
	byzantineActions []actions.WeightedAction,
	networkActions []actions.WeightedAction,
	rounds int,
	logger logging.Logger,
) error {
	// check that shit makes sense
	// rename to lower case

	nodeWeights := make(map[stdtypes.NodeID]types.VoteWeight, len(nodes))
	for i := range nodes {
		id := stdtypes.NewNodeIDFromInt(i)
		nodeWeights[id] = "1"
	}
	instanceUID := []byte("fuzzing instance")
	nodeConfigs := make(map[stdtypes.NodeID]nodeinstance.BcbNodeInstanceConfig, len(nodes))

	// TODO: the rounds should actually be part of fuzzer, and not handeled here...
	config := nodeinstance.BcbNodeInstanceConfig{InstanceUID: instanceUID, NumberOfNodes: len(nodes), Leader: sender, FakeTransport: deploytest.NewFakeTransport(nodeWeights), ReportPath: "."}
	for _, nodeID := range nodes {
		fmt.Println(nodeID)
		nodeConfigs[nodeID] = config
	}

	systemConfig := &properties.SystemConfig{
		AllNodes:       nodes,
		Sender:         sender,
		ByzantineNodes: byzantineNodes,
	}

	checkerProperties := checker.Properties{
		"validity":    properties.NewValidity(*systemConfig, logger),
		"integrity":   properties.NewIntegrity(*systemConfig, logger),
		"consistency": properties.NewConsistency(*systemConfig, logger),
	}

	fuzzer, err := fuzzer.NewFuzzer(nodeinstance.CreateBcbNodeInstance, nodeConfigs, byzantineNodes, puppeteerEvents, byzantineActions, networkActions, "", checkerProperties)
	if err != nil {
		return es.Errorf("failed to create fuzzer: %v", err)
	}

	// TODO: properly deal with context
	ctx := context.Background()
	err = fuzzer.Run(ctx, name, rounds, MAX_EVENTS, logger)
	if err != nil {
		return es.Errorf("fuzzer encountered an issue: %v", err)
	}

	return nil
}

func main() {
	logger := logging.ConsoleDebugLogger
	logger = logging.Synchronize(logger)
	fuzzBCB("test", []stdtypes.NodeID{"0", "1", "2", "3"}, []stdtypes.NodeID{}, stdtypes.NodeID("0"), weightedActionsForByzantineNodes, weightedActionsForNetwork, 1, logger)
}
