package centraladversay

import (
	"context"
	"fmt"
	"reflect"
	"slices"
	"sync"
	"time"

	"github.com/filecoin-project/mir"
	"github.com/filecoin-project/mir/checker"
	"github.com/filecoin-project/mir/fuzzer2/actions"
	"github.com/filecoin-project/mir/fuzzer2/centraladversary/cortexcreeper"
	"github.com/filecoin-project/mir/fuzzer2/heartbeat"
	"github.com/filecoin-project/mir/fuzzer2/nodeinstance"
	"github.com/filecoin-project/mir/fuzzer2/puppeteer"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/util/maputil"
	broadcastevents "github.com/filecoin-project/mir/samples/broadcast/events"
	"github.com/filecoin-project/mir/stdevents"
	"github.com/filecoin-project/mir/stdtypes"

	es "github.com/go-errors/errors"
)

var (
	MaxEventsOrHeartbeatShutdown = es.Errorf("Shutting down because max events or max inactive heartbeats has been exceeded.")
)

type ActionTraceEntry struct {
	node                stdtypes.NodeID
	afterByzantineNodes []stdtypes.NodeID
	actionLog           string
}

type Adversary struct {
	nodeIds          []stdtypes.NodeID
	nodeInstances    map[stdtypes.NodeID]nodeinstance.NodeInstance
	cortexCreepers   map[stdtypes.NodeID]*cortexcreeper.CortexCreeper
	actionSelector   actions.Actions
	eventsOfInterest map[reflect.Type]bool
	byzantineNodes   []stdtypes.NodeID
	actionTrace      []ActionTraceEntry
}

func NewAdversary[T interface{}](
	createNodeInstance nodeinstance.NodeInstanceCreationFunc[T],
	nodeConfigs nodeinstance.NodeConfigs[T],
	eventsOfInterest []stdtypes.Event,
	weightedActions []actions.WeightedAction,
	byzantineNodes []stdtypes.NodeID,
	logger logging.Logger,
) (*Adversary, error) {

	nodeIDs := maputil.GetKeys(nodeConfigs)
	for _, byzNodeID := range byzantineNodes {
		if !slices.Contains(nodeIDs, byzNodeID) {
			return nil, es.Errorf("Cannot use node %s as a byzantine node as there is no node config for this node.", byzNodeID)
		}
	}

	nodeInstances := make(map[stdtypes.NodeID]nodeinstance.NodeInstance)
	cortexCreepers := make(map[stdtypes.NodeID]*cortexcreeper.CortexCreeper, len(nodeConfigs))

	for nodeID, config := range nodeConfigs {
		nodeLogger := logging.Decorate(logger, string(nodeID)+" ")
		// nodeLogger := logger
		cortexCreeper := cortexcreeper.NewCortexCreeper()
		cortexCreepers[nodeID] = cortexCreeper
		nodeInstance, err := createNodeInstance(nodeID, config, cortexCreeper, nodeLogger)
		if err != nil {
			return nil, es.Errorf("Failed to create node instance with id %s: %v", nodeID, err)
		}
		nodeInstances[nodeID] = nodeInstance
	}

	actionSelector, err := actions.NewRandomActions(weightedActions)
	if err != nil {
		return nil, err
	}

	eventTypesOfInterest := make(map[reflect.Type]bool, len(eventsOfInterest))
	for _, e := range eventsOfInterest {
		eventTypesOfInterest[reflect.TypeOf(e)] = true
	}

	return &Adversary{
		maputil.GetKeys(nodeInstances),
		nodeInstances,
		cortexCreepers,
		actionSelector,
		eventTypesOfInterest,
		byzantineNodes,
		make([]ActionTraceEntry,
			0),
	}, nil
}

func (a *Adversary) RunNodes(ctx context.Context) error {
	// setup all nodes and run them
	wg := &sync.WaitGroup{}
	nodesContext, advCancel := context.WithCancel(ctx)
	defer advCancel()
	for nodeId, nodeInstance := range a.nodeInstances {
		go func() {
			errChan := make(chan error)
			wg.Add(1)
			defer wg.Done()
			defer nodeInstance.Stop()
			defer nodeInstance.Cleanup()
			nodeInstance.Setup()
			go func() {
				errChan <- nodeInstance.Run(nodesContext)
				fmt.Printf("node %s stopped\n", nodeId)
			}()

			select {
			case err := <-errChan:
				if err != nil && err != mir.ErrStopped {
					// node failed, kill all other nodes
					fmt.Printf("Node %s failed with error: %v", nodeId, err)
				}
			case <-nodesContext.Done():
			}
		}()
	}

	time.Sleep(time.Second)

	for _, nodeInstance := range a.nodeInstances {
		// inject heartbeat start
		nodeInstance.GetNode().InjectEvents(nodesContext,
			stdtypes.ListOf(
				stdevents.NewTimerRepeat("timer",
					time.Second,
					stdtypes.RetentionIndex(^uint64(0)),
					heartbeat.NewHeartbeat("null"),
				),
			),
		)
	}
	<-nodesContext.Done()

	wg.Wait()

	return nil
}

func (a *Adversary) RunCentralAdversary(maxEvents, maxHearbeatsInactive int, checker *checker.Checker, ctx context.Context) error {
	eventCount := 0
	heartbeatCount := 0

	// slice of cortex creepers ordered by nodeIds to easily reference them to the select cases
	ccsNodeIds := maputil.GetSortedKeys(a.cortexCreepers)
	ccs := maputil.GetValuesOf(a.cortexCreepers, ccsNodeIds)

	for {
		selectCases := make([]reflect.SelectCase, 0, len(ccs)+1)
		selectCases = append(selectCases, reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(ctx.Done()),
		})
		for _, cc := range ccs {
			selectCases = append(selectCases, reflect.SelectCase{
				Dir:  reflect.SelectRecv,
				Chan: reflect.ValueOf(cc.GetEventsIn()),
			})
		}

		ind, value, _ := reflect.Select(selectCases)
		if ind == 0 {
			// context was cancelled
			return nil
		}

		// process this event
		elIterator := value.Interface().(*stdtypes.EventList).Iterator()
		for event := elIterator.Next(); event != nil; event = elIterator.Next() {
			eventCount++
			switch event.(type) {
			case *heartbeat.Heartbeat:
				heartbeatCount++
			default:
				heartbeatCount = 0
			}

			if heartbeatCount > maxHearbeatsInactive || eventCount > maxEvents {
				return MaxEventsOrHeartbeatShutdown
			}

			sourceNodeID := ccsNodeIds[ind-1]

			// TODO: figure out how to actually do this filtering
			// hardcoding for now...

			// if not event of interest, just push and continue
			// eventType := reflect.TypeOf(event)
			// if _, ok := a.eventsOfInterest[eventType]; !ok {
			// 	fmt.Printf("Forwarding %v\n", event)
			// 	a.cortexCreepers[ind-1].PushEvents(stdtypes.ListOf(event))
			// 	continue
			// }

			// hardcoded tmp solution
			switch event.(type) {
			case *broadcastevents.BroadcastRequest:
				if !a.nodeIsPermittedToTakeByzantineAction(a.nodeIds[ind-1]) {
					a.pushEvents(sourceNodeID, stdtypes.ListOf(event), checker)
					continue
				}
			case *broadcastevents.Deliver:
				if !a.nodeIsPermittedToTakeByzantineAction(a.nodeIds[ind-1]) {
					a.pushEvents(sourceNodeID, stdtypes.ListOf(event), checker)
					continue
				}
			case *stdevents.SendMessage:
			case *stdevents.MessageReceived:
			default:
				a.pushEvents(sourceNodeID, stdtypes.ListOf(event), checker)
				continue
			}

			// otherwise pick an action to apply to this event
			action := a.actionSelector.SelectAction()
			actionLog, newEvents, err := action(event, sourceNodeID, a.byzantineNodes)
			if err != nil {
				return err
			}

			// ignore "" bc this signifies noop - not the best solution but works for now
			if actionLog != "" {
				a.actionTrace = append(a.actionTrace, ActionTraceEntry{
					node:                a.nodeIds[ind-1],
					afterByzantineNodes: slices.Clone(a.byzantineNodes),
					actionLog:           actionLog,
				})
			}

			// TODO: check for nil?
			for injectNodeID, injectEvents := range newEvents {
				fmt.Printf("%s (%s) - injecting v\n", injectNodeID, sourceNodeID)
				a.pushEvents(injectNodeID, injectEvents, checker)
			}
		}
	}
}

func (a *Adversary) pushEvents(nodeID stdtypes.NodeID, events *stdtypes.EventList, checker *checker.Checker) {
	if checker != nil {
		// if there's a checker, duplicate the events and have the checker look at them
		eIter := events.Iterator()
		for e := eIter.Next(); e != nil; e = eIter.Next() {
			// TODO: must be duplicated! Don't want checker to possibly affect the system
			checker.NextEvent(e)
		}
	}
	// TODO: we know that this cc exists, should I still check to make sure?
	cc := a.cortexCreepers[nodeID]
	cc.PushEvents(events)
}

func (a *Adversary) nodeIsPermittedToTakeByzantineAction(nodeId stdtypes.NodeID) bool {
	return slices.Contains(a.byzantineNodes, nodeId)
}

func (a *Adversary) RunExperiment(puppeteer puppeteer.Puppeteer, checker *checker.Checker, maxEvents, maxHeartbeatsInactive int) error {
	ctx, cancel := context.WithCancel(context.Background())
	wg := sync.WaitGroup{}
	defer cancel()
	if checker != nil {
		defer checker.Stop()
	}

	nodesErr := make(chan error)
	caErr := make(chan error)
	checkerErr := make(chan error)

	go func() {
		wg.Add(1)
		defer wg.Done()
		defer close(nodesErr)
		nodesErr <- a.RunNodes(ctx)
	}()

	go func() {
		wg.Add(1)
		defer wg.Done()
		defer close(caErr)
		// defer cancel() // cancel to stop nodes when adv had enough
		caErr <- a.RunCentralAdversary(maxEvents, maxHeartbeatsInactive, checker, ctx)
	}()

	if checker != nil {
		go func() {
			wg.Add(1)
			defer wg.Done()
			defer close(checkerErr)
			checkerErr <- checker.Start()
		}()
	}

	err := puppeteer.Run(a.nodeInstances)
	if err != nil {
		return err
	}

	select {
	case err = <-nodesErr:
		if err != nil {
			return es.Errorf("Nodes runtime (CA) error: %v", err)
		}
	case err = <-caErr:
		if err != nil {
			return es.Errorf("Central Adversary error: %v", err)
		}
	case err = <-caErr:
		if err != nil {
			return es.Errorf("Checker error: %v", err)
		}
	}

	wg.Wait()

	return nil
}

func (a *Adversary) GetByzantineNodes() []stdtypes.NodeID {
	return a.byzantineNodes
}

func (a *Adversary) GetActionLogString() string {
	logStr := ""
	for _, al := range a.actionTrace {
		logStr += fmt.Sprintf("Node %s took action: %v\nbyzNodes: %v\n\n", al.node, al.actionLog, al.afterByzantineNodes)
	}
	return logStr
}
