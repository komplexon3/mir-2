package centraladversay

import (
	"context"
	"fmt"
	"reflect"
	"slices"
	"sync"
	"time"

	"github.com/filecoin-project/mir"
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
	nodeIds           []stdtypes.NodeID
	nodeInstances     map[stdtypes.NodeID]nodeinstance.NodeInstance
	cortexCreepers    []*cortexcreeper.CortexCreeper
	actionSelector    actions.Actions
	eventsOfInterest  map[reflect.Type]bool
	byzantineNodes    []stdtypes.NodeID
	maxByzantineNodes int
	actionTrace       []ActionTraceEntry
}

func NewAdversary[T interface{}](
	createNodeInstance nodeinstance.NodeInstanceCreationFunc[T],
	nodeConfigs nodeinstance.NodeConfigs[T],
	eventsOfInterest []stdtypes.Event,
	weightedActions []actions.WeightedAction,
	maxByzantineNodes int,
	logger logging.Logger,
) (*Adversary, error) {

	if maxByzantineNodes > len(nodeConfigs) {
		return nil, es.Errorf("Cannot allow for more byzantine nodes than the number of nodes in the system.")
	}

	nodeInstances := make(map[stdtypes.NodeID]nodeinstance.NodeInstance)
	cortexCreepers := make([]*cortexcreeper.CortexCreeper, 0, len(nodeConfigs))

	for nodeID, config := range nodeConfigs {
		nodeLogger := logging.Decorate(logger, string(nodeID)+" ")
		// nodeLogger := logger
		cortexCreeper := cortexcreeper.NewCortexCreeper()
		cortexCreepers = append(cortexCreepers, cortexCreeper)
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

	return &Adversary{maputil.GetKeys(nodeInstances), nodeInstances, cortexCreepers, actionSelector, eventTypesOfInterest, make([]stdtypes.NodeID, 0), maxByzantineNodes, make([]ActionTraceEntry, 0)}, nil
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
	fmt.Println("node wg done")

	return nil
}

func (a *Adversary) RunCentralAdversary(maxEvents, maxHearbeatsInactive int, ctx context.Context) error {
	eventCount := 0
	heartbeatCount := 0

	for {
		selectCases := make([]reflect.SelectCase, 0, len(a.cortexCreepers)+1)
		selectCases = append(selectCases, reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(ctx.Done()),
		})
		for _, cc := range a.cortexCreepers {
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
					a.cortexCreepers[ind-1].PushEvents(stdtypes.ListOf(event))
					continue
				}
			case *broadcastevents.Deliver:
				if !a.nodeIsPermittedToTakeByzantineAction(a.nodeIds[ind-1]) {
					a.cortexCreepers[ind-1].PushEvents(stdtypes.ListOf(event))
					continue
				}
			case *stdevents.SendMessage:
			case *stdevents.MessageReceived:
			default:
				a.cortexCreepers[ind-1].PushEvents(stdtypes.ListOf(event))
				continue
			}
			// ^ check if can take byzantine action

			// otherwise pick an action to apply to this event
			action := a.actionSelector.SelectAction()
			newEvents, wasByzantine, actionLog, err := action(event)
			if err != nil {
				return err
			}
			if wasByzantine {
				a.registerNodeTookByzantineAction(a.nodeIds[ind-1])
			}

			// ignore "" bc this signifies noop - not the best solution but works for now
			if actionLog != "" {
				a.actionTrace = append(a.actionTrace, ActionTraceEntry{
					node:                a.nodeIds[ind-1],
					afterByzantineNodes: slices.Clone(a.byzantineNodes),
					actionLog:           actionLog,
				})
			}

			a.cortexCreepers[ind-1].PushEvents(newEvents)
		}
	}
}

func (a *Adversary) nodeIsPermittedToTakeByzantineAction(nodeId stdtypes.NodeID) bool {
	if len(a.byzantineNodes) < a.maxByzantineNodes {
		return true
	}

	return slices.Contains(a.byzantineNodes, nodeId)
}

func (a *Adversary) registerNodeTookByzantineAction(nodeId stdtypes.NodeID) {
	if !slices.Contains(a.byzantineNodes, nodeId) {
		a.byzantineNodes = append(a.byzantineNodes, nodeId)
	}
}

func (a *Adversary) RunExperiment(puppeteer puppeteer.Puppeteer, maxEvents, maxHeartbeatsInactive int) error {
	ctx, cancel := context.WithCancel(context.Background())
	wg := sync.WaitGroup{}
	defer cancel()

	nodesErr := make(chan error)
	go func() {
		wg.Add(1)
		defer wg.Done()
		nodesErr <- a.RunNodes(ctx)
	}()

	caErr := make(chan error)
	go func() {
		wg.Add(1)
		defer wg.Done()
		defer cancel() // cancel to stop nodes when adv had enough
		caErr <- a.RunCentralAdversary(maxEvents, maxHeartbeatsInactive, ctx)
	}()

	err := puppeteer.Run(a.nodeInstances)
	if err != nil {
		return err
	}

	select {
	case err = <-nodesErr:
		if err != nil {
			return err
		}
	case err = <-caErr:
		if err != nil {
			return err
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
