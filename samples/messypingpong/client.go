package main

import (
	"context"
	"fmt"
	"reflect"
	"sync"

	es "github.com/go-errors/errors"

	"github.com/filecoin-project/mir/fuzzer/network"
	"github.com/filecoin-project/mir/stdtypes"

	"github.com/filecoin-project/mir"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/samples/messypingpong/pingpong"
)

func main() {
	nodeIds := []stdtypes.NodeID{"0", "1"}

	logger := logging.Synchronize(logging.ConsoleTraceLogger)

	wg := &sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())
	var err error
	errPrimary := make(chan error)
	errSecondary := make(chan error)
	inactiveChanNetwork := make(chan chan struct{})
	var continueChanNetwork chan struct{}
	defer close(errPrimary)
	defer close(errSecondary)
	defer close(inactiveChanNetwork)
	inactiveNodeChans := make(map[stdtypes.NodeID]chan chan struct{}, len(nodeIds))
	continueNodeChans := make(map[stdtypes.NodeID]chan struct{}, len(nodeIds))
	for _, nodeID := range nodeIds {
		inactiveNodeChans[nodeID] = make(chan chan struct{})
		continueNodeChans[nodeID] = nil
	}

	defer func() {
		for _, nc := range inactiveNodeChans {
			close(nc)
		}
	}()

	nLogger := logging.Decorate(logger, "N - ")
	ft := network.NewFuzzTransport(nodeIds, inactiveChanNetwork, nLogger, ctx)

	wg.Add(1)
	go func() {
		defer wg.Done()
		errPrimary <- run(true, ft, inactiveNodeChans[stdtypes.NodeID("0")], logger, ctx)
	}()

	wg.Add(1)
	go func() {
		wg.Done()
		errSecondary <- run(false, ft, inactiveNodeChans[stdtypes.NodeID("1")], logger, ctx)
	}()

	sLogger := logging.Decorate(logger, "S - ")
	wg.Add(1)
	go func() {
		defer wg.Done()
		activeCount := len(nodeIds) + 1 // nodes + network
		wasInactive := false
		stop := false
		for !stop {
			selectCases := make([]reflect.SelectCase, 0, 2*len(nodeIds)+3) // 2*nodes + network inactive and active + ctx done
			selectReactions := make([]func(recvVal reflect.Value), 0, 2*len(nodeIds)+3)

			selectCases = append(selectCases, reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(ctx.Done())})
			selectReactions = append(selectReactions, func(_ reflect.Value) { stop = true })

			// continue cases - must be first
			// network
			selectCases = append(selectCases, reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(continueChanNetwork)})
			selectReactions = append(selectReactions, func(_ reflect.Value) {
				sLogger.Log(logging.LevelTrace, "Network active")
				continueChanNetwork = nil
				activeCount++
			})

			// nodes
			for nodeID, continueChan := range continueNodeChans {
				selectCases = append(selectCases, reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(continueChan)})
				selectReactions = append(selectReactions, func(_ reflect.Value) {
					sLogger.Log(logging.LevelTrace, fmt.Sprintf("Node %s active", nodeID))
					continueNodeChans[nodeID] = nil
					activeCount++
				})
			}

			// inactive cases
			// network
			selectCases = append(selectCases, reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(inactiveChanNetwork)})
			selectReactions = append(selectReactions, func(recvVal reflect.Value) {
				sLogger.Log(logging.LevelTrace, "Network inactive")
				continueChanNetwork = recvVal.Interface().(chan struct{})
				activeCount--
			})

			// nodes
			for nodeID, inactiveChan := range inactiveNodeChans {
				selectCases = append(selectCases, reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(inactiveChan)})
				selectReactions = append(selectReactions, func(recvVal reflect.Value) {
					sLogger.Log(logging.LevelTrace, fmt.Sprintf("Node %s inactive", nodeID))
					continueNodeChans[nodeID] = recvVal.Interface().(chan struct{})
					activeCount--
				})
			}

			sLogger.Log(logging.LevelTrace, "~+~")
			chosenCase, recvValue, _ := reflect.Select(selectCases)
			selectReactions[chosenCase](recvValue)
			sLogger.Log(logging.LevelTrace, "~_~")

			if activeCount == 0 {
				sLogger.Log(logging.LevelTrace, "=== System inactive ===")
				wasInactive = true
			} else if wasInactive {
				sLogger.Log(logging.LevelTrace, "=== System active again ===")
				wasInactive = false
			}

		}
	}()

	select {
	case pErr := <-errPrimary:
		err = pErr
		cancel()
	case sErr := <-errSecondary:
		err = sErr
		cancel()
	}

	wg.Wait()
	if err != nil {
		fmt.Println(err)
	}
}

func run(isPrimary bool, fuzzTransport *network.FuzzTransport, inactiveChan chan chan struct{}, logger logging.Logger, ctx context.Context) error {
	selfNode := stdtypes.NodeID("0")
	otherNode := stdtypes.NodeID("1")
	if !isPrimary {
		selfNode = stdtypes.NodeID("1")
		otherNode = stdtypes.NodeID("0")
	}

	logger = logging.Decorate(logger, string(selfNode)+" - ")

	aModule := pingpong.NewModule(pingpong.ModuleConfig{
		Self:      stdtypes.ModuleID("A"),
		Other:     stdtypes.ModuleID("B"),
		Transport: "transport",
		SelfNode:  selfNode,
		OtherNode: otherNode,
	}, logger)
	bModule := pingpong.NewModule(pingpong.ModuleConfig{
		Self:      stdtypes.ModuleID("B"),
		Other:     stdtypes.ModuleID("A"),
		Transport: "transport",
		SelfNode:  selfNode,
		OtherNode: otherNode,
	}, logger)

	transport, err := fuzzTransport.Link(selfNode)
	if err != nil {
		return es.Errorf("failed to setup link: %v", err)
	}
	transport.Connect(nil) // fuzz transport doen't need membership

	m := map[stdtypes.ModuleID]modules.Module{
		"A":         aModule,
		"B":         bModule,
		"transport": transport,
	}

	// create a Mir node
	node, err := mir.NewNodeWithIdleDetection("internalpingpongtest", mir.DefaultNodeConfig().WithLogger(logger), m, nil, inactiveChan)
	// node, err := mir.NewNode("internalpingpongtest", mir.DefaultNodeConfig().WithLogger(logger), m, nil)
	if err != nil {
		return es.Errorf("error creating a Mir node: %w", err)
	}

	// run the node
	err = node.Run(ctx)
	if err != nil {
		return es.Errorf("error running node: %w", err)
	}

	return nil
}
