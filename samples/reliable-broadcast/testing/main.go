package main

import (
	"context"
	"fmt"
	"math/rand/v2"
	"strings"
	"time"

	es "github.com/go-errors/errors"

	"github.com/filecoin-project/mir/fuzzer"
	"github.com/filecoin-project/mir/fuzzer/actions"
	"github.com/filecoin-project/mir/stdevents"

	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/trantor/types"
	broadcastevents "github.com/filecoin-project/mir/samples/reliable-broadcast/events"

	broadcastmessages "github.com/filecoin-project/mir/samples/reliable-broadcast/messages"
	"github.com/filecoin-project/mir/samples/reliable-broadcast/testing/nodeinstance"
	"github.com/filecoin-project/mir/samples/reliable-broadcast/testing/properties"
	"github.com/filecoin-project/mir/stdtypes"
)

const (
	MAX_RUN_DURATION = 1 * time.Second
	SEED1            = 123
	SEED2            = 321
)

var puppeteerEvents = []actions.DelayedEvents{
	{
		NodeID: stdtypes.NodeID("0"),
		Events: stdtypes.ListOf(broadcastevents.NewBroadcastRequest("broadcast", "hello")),
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
	actions.NewWeightedAction(func(e stdtypes.Event, sourceNode stdtypes.NodeID, byzantineNodes []stdtypes.NodeID) (string, map[stdtypes.NodeID]*stdtypes.EventList, []actions.DelayedEvents, error) {
		return fmt.Sprintf("event delayed (network) %v", e.ToString()),
			nil,
			[]actions.DelayedEvents{
				{
					NodeID: sourceNode,
					Events: stdtypes.ListOf(e),
				},
			},
			nil
	}, 1),
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
	actions.NewWeightedAction(func(e stdtypes.Event, sourceNode stdtypes.NodeID, byzantineNodes []stdtypes.NodeID) (string, map[stdtypes.NodeID]*stdtypes.EventList, []actions.DelayedEvents, error) {
		return fmt.Sprintf("dropped event %s", e.ToString()),
			nil,
			nil,
			nil
	}, 2),
	actions.NewWeightedAction(func(e stdtypes.Event, sourceNode stdtypes.NodeID, byzantineNodes []stdtypes.NodeID) (string, map[stdtypes.NodeID]*stdtypes.EventList, []actions.DelayedEvents, error) {
		e2, err := e.SetMetadata("duplicated", true)
		if err != nil {
			// TODO: should a failed action just be a "noop"
			return "", nil, nil, err
		}
		return fmt.Sprintf("duplicated event %v", e.ToString()),
			map[stdtypes.NodeID]*stdtypes.EventList{
				sourceNode: stdtypes.ListOf(e, e2),
			},
			nil,
			nil
	}, 1),
	actions.NewWeightedAction(func(e stdtypes.Event, sourceNode stdtypes.NodeID, byzantineNodes []stdtypes.NodeID) (string, map[stdtypes.NodeID]*stdtypes.EventList, []actions.DelayedEvents, error) {
		// swap message content on sends, if it is a message
		sendEvt, ok := e.(*stdevents.SendMessage)
		if !ok {
			return "noop",
				map[stdtypes.NodeID]*stdtypes.EventList{
					sourceNode: stdtypes.ListOf(e),
				},
				nil,
				nil
		}

		switch msg := sendEvt.Payload.(type) {
		case *broadcastmessages.StartMessage:
			msg.Data = msg.Data + "-changed"
		case *broadcastmessages.EchoMessage:
			msg.Data = msg.Data + "-changed"
		case *broadcastmessages.ReadyMessage:
			msg.Data = msg.Data + "-changed"
		default:
			return "",
				nil,
				nil,
				fmt.Errorf("unknown message type: %t", msg)
		}
		return fmt.Sprintf("swapped message data %v", e.ToString()),
			map[stdtypes.NodeID]*stdtypes.EventList{
				sourceNode: stdtypes.ListOf(e),
			},
			nil,
			nil
	}, 1),
}

func isInterestingEvent(event stdtypes.Event) bool {
	switch event.(type) {
	// case *broadcastevents.BroadcastRequest:
	// case *broadcastevents.Deliver:
	case *stdevents.SendMessage:
	case *stdevents.MessageReceived:
	default:
		return false
	}
	return true
}

func fuzz(
	name string,
	nodes []stdtypes.NodeID,
	byzantineNodes []stdtypes.NodeID,
	sender stdtypes.NodeID,
	byzantineActions []actions.WeightedAction,
	networkActions []actions.WeightedAction,
	rounds int,
	logLevel logging.LogLevel,
) (int, error) {
	// check that shit makes sense
	// rename to lower case

	nodeWeights := make(map[stdtypes.NodeID]types.VoteWeight, len(nodes))
	for i := range nodes {
		id := stdtypes.NewNodeIDFromInt(i)
		nodeWeights[id] = "1"
	}
	instanceUID := []byte("fuzzing instance")
	nodeConfigs := make(map[stdtypes.NodeID]nodeinstance.ReliableBroadcastNodeInstanceConfig, len(nodes))

	// TODO: the rounds should actually be part of fuzzer, and not handeled here...
	config := nodeinstance.ReliableBroadcastNodeInstanceConfig{InstanceUID: instanceUID, NumberOfNodes: len(nodes), Leader: sender}
	for _, nodeID := range nodes {
		nodeConfigs[nodeID] = config
	}

	checkerParams := properties.SystemConfig{
		AllNodes:       nodes,
		Sender:         sender,
		ByzantineNodes: byzantineNodes,
	}

	fuzzer, err := fuzzer.NewFuzzer(
		nodeinstance.CreateBroadcastNodeInstance,
		nodeConfigs,
		byzantineNodes,
		puppeteerEvents,
		isInterestingEvent,
		byzantineActions,
		networkActions,
		properties.CreateReliableBroadcastChecker,
		checkerParams,
		fmt.Sprintf("./report_%s_%s", time.Now().Format("2006-01-02_15-04-05"), strings.Join(strings.Split(name, " "), "_")),
	)
	if err != nil {
		return 0, es.Errorf("failed to create fuzzer: %v", err)
	}

	// TODO: properly deal with context
	ctx := context.Background()
	hits, err := fuzzer.Run(ctx, name, rounds, MAX_RUN_DURATION, rand.New(rand.NewPCG(SEED1, SEED2)), logLevel)
	if err != nil {
		return 0, es.Errorf("fuzzer encountered an issue: %v", err)
	}

	return hits, nil
}

func main() {
	rounds := 10000
	logLevel := logging.LevelTrace
	startTime := time.Now()
	hits, err := fuzz("test-reliable", []stdtypes.NodeID{"0", "1", "2", "3"}, []stdtypes.NodeID{"1", "2"}, stdtypes.NodeID("0"), weightedActionsForByzantineNodes, weightedActionsForNetwork, rounds, logLevel)
	if err != nil {
		fmt.Println(err)
	}
	duration := time.Since(startTime)
	fmt.Printf("=================================================================\nInteresting cases %d (out of %d - %.3f%%)\n", hits, rounds, float32(hits)/float32(rounds)*100)
	fmt.Printf("=================================================================\nExecution time: %s - for a total of %d rounds (avg per round: %s)\n", duration, rounds, time.Duration(int64(duration)/int64(rounds)))
}
