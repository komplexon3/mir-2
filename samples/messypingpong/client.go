package main

import (
	"context"
	"fmt"
	"sync"

	es "github.com/go-errors/errors"

	"github.com/filecoin-project/mir/fuzzer2/network"
	"github.com/filecoin-project/mir/stdtypes"

	"github.com/filecoin-project/mir"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/samples/messypingpong/pingpong"
)

func main() {
	nodeIds := []stdtypes.NodeID{"0", "1"}

	// fkt := deploytest.NewFakeTransport(map[stdtypes.NodeID]types.VoteWeight{
	// 	stdtypes.NodeID("0"): "1",
	// 	stdtypes.NodeID("1"): "1",
	// })
	//
	logger := logging.Synchronize(logging.ConsoleTraceLogger)

	wg := &sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())
	var err error
	errPrimary := make(chan error)
	errSecondary := make(chan error)
	inactiveChanNode0 := make(chan chan struct{})
	inactiveChanNode1 := make(chan chan struct{})
	inactiveChanNodeNetwork := make(chan chan struct{})
	// var inactiveChanNode0 chan chan struct{}
	// var inactiveChanNode1 chan chan struct{}
	defer close(errPrimary)
	defer close(errSecondary)
	defer close(inactiveChanNode0)
	defer close(inactiveChanNode1)
	defer close(inactiveChanNodeNetwork)

	nLogger := logging.Decorate(logger, "N - ")
	ft := network.NewFuzzTransport(nodeIds, &inactiveChanNodeNetwork, nLogger, ctx)

	wg.Add(1)
	go func() {
		defer wg.Done()
		errPrimary <- run(true, ft, inactiveChanNode0, logger, ctx)
	}()

	wg.Add(1)
	go func() {
		wg.Done()
		errSecondary <- run(false, ft, inactiveChanNode1, logger, ctx)
	}()

	sLogger := logging.Decorate(logger, "S - ")
	wg.Add(1)
	go func(wg *sync.WaitGroup, ctx context.Context) {
		defer wg.Done()

		wg.Add(1)
		go func() {
			defer wg.Done()
			node0Inactive := false
			node1Inactive := false
			networkInactive := false
			wasInactive := false

			var continueChanNode0 chan struct{}
			var continueChanNode1 chan struct{}
			var continueChanNetwork chan struct{}

			for {
				select {
				case cc := <-inactiveChanNode0:
					sLogger.Log(logging.LevelTrace, "Node 0 inactive")
					node0Inactive = true
					continueChanNode0 = cc
				case cc := <-inactiveChanNode1:
					sLogger.Log(logging.LevelTrace, "Node 1 inactive")
					node1Inactive = true
					continueChanNode1 = cc
				case cc := <-inactiveChanNodeNetwork:
					sLogger.Log(logging.LevelTrace, "Network inactive")
					networkInactive = true
					continueChanNetwork = cc
				case <-continueChanNode0:
					sLogger.Log(logging.LevelTrace, "Node 0 active")
					node0Inactive = false
					continueChanNode0 = nil
				case <-continueChanNode1:
					sLogger.Log(logging.LevelTrace, "Node 1 active")
					node1Inactive = false
					continueChanNode1 = nil
				case <-continueChanNetwork:
					sLogger.Log(logging.LevelTrace, "Network active")
					networkInactive = false
					continueChanNetwork = nil
				case <-ctx.Done():
					return
				}

				sLogger.Log(logging.LevelTrace, fmt.Sprintf("n0 %t, n1 %t, net %t", node0Inactive, node1Inactive, networkInactive))

				if node0Inactive && node1Inactive && networkInactive {
					sLogger.Log(logging.LevelTrace, "=== System inactive ===")
					wasInactive = true
				} else if wasInactive {
					sLogger.Log(logging.LevelTrace, "=== System active again ===")
					wasInactive = false
				}
			}
		}()
	}(wg, ctx)

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
	node, err := mir.NewNode("internalpingpongtest", mir.DefaultNodeConfig().WithLogger(logger), m, nil)
	if err != nil {
		return es.Errorf("error creating a Mir node: %w", err)
	}

	node.SetInactiveNotificationChannel(inactiveChan)

	// run the node
	err = node.Run(ctx)
	if err != nil {
		return es.Errorf("error running node: %w", err)
	}

	return nil
}
