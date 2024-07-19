package centraladversay

import (
	"context"
	"fmt"
	"math/rand/v2"
	"reflect"
	"slices"
	"sync"

	"github.com/filecoin-project/mir"
	"github.com/filecoin-project/mir/checker"
	"github.com/filecoin-project/mir/fuzzer2/actions"
	"github.com/filecoin-project/mir/fuzzer2/centraladversary/cortexcreeper"
	"github.com/filecoin-project/mir/fuzzer2/heartbeat"
	"github.com/filecoin-project/mir/fuzzer2/nodeinstance"
	"github.com/filecoin-project/mir/fuzzer2/puppeteer"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/util/maputil"
	"github.com/filecoin-project/mir/pkg/util/sliceutil"
	"github.com/filecoin-project/mir/pkg/vcinterceptor/vectorclock"
	broadcastevents "github.com/filecoin-project/mir/samples/bcb-native/events"
	"github.com/filecoin-project/mir/stdevents"
	"github.com/filecoin-project/mir/stdtypes"

	es "github.com/go-errors/errors"
)

var (
	MaxEventsOrHeartbeatShutdown = es.Errorf("Shutting down because max events or max inactive heartbeats has been exceeded.")
	IdleWithoutDelayedEvents     = es.Errorf("Shutting down because all nodes are idle and there are no msgs to delay.")
)

type ActionTraceEntry struct {
	node      stdtypes.NodeID
	actionLog string
}

type Adversary struct {
	byzantineActionSelector actions.Actions
	networkActionSelector   actions.Actions
	nodeInstances           map[stdtypes.NodeID]nodeinstance.NodeInstance
	cortexCreepers          map[stdtypes.NodeID]*cortexcreeper.CortexCreeper
	idleNodesMonitor        *IdleNodesMonitor
	undeliveredMsgs         map[string]struct{}
	nodeIds                 []stdtypes.NodeID
	byzantineNodes          []stdtypes.NodeID
	actionTrace             []ActionTraceEntry
	delayedEvents           []actions.DelayedEvents
	logger                  logging.Logger
}

func NewAdversary[T any](
	createNodeInstance nodeinstance.NodeInstanceCreationFunc[T],
	nodeConfigs nodeinstance.NodeConfigs[T],
	byzantineWeightedActions []actions.WeightedAction,
	networkWeightedActions []actions.WeightedAction,
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
		nodeLogger := logging.Decorate(logger, string(nodeID)+" - ")
		cortexCreeper := cortexcreeper.NewCortexCreeper()
		cortexCreepers[nodeID] = cortexCreeper
		nodeInstance, err := createNodeInstance(nodeID, config, cortexCreeper, nodeLogger)
		if err != nil {
			return nil, es.Errorf("Failed to create node instance with id %s: %v", nodeID, err)
		}
		nodeInstances[nodeID] = nodeInstance
	}

	byzantineActionSelector, err := actions.NewRandomActions(append(byzantineWeightedActions, networkWeightedActions...))
	if err != nil {
		return nil, err
	}

	networkActionSelector, err := actions.NewRandomActions(byzantineWeightedActions)
	if err != nil {
		return nil, err
	}

	return &Adversary{
		byzantineActionSelector,
		networkActionSelector,
		nodeInstances,
		cortexCreepers,
		NewIdleNodesMonitor(),
		make(map[string]struct{}),
		byzantineNodes,
		maputil.GetKeys(nodeInstances),
		make([]ActionTraceEntry, 0),
		make([]actions.DelayedEvents, 0),
		logging.Decorate(logger, "CA - "),
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

	idleNotificationC := make(chan chan struct{})
	defer close(idleNotificationC)
	idleDetectionCs := sliceutil.Transform(ccs, func(_ int, cc *cortexcreeper.CortexCreeper) chan chan struct{} { return cc.IdleDetectionC })
	go a.idleNodesMonitor.Run(ctx, idleDetectionCs, idleNotificationC)

	for {
		// fan in events from different cortexCreepers + idle detection + error
		selectCases := make([]reflect.SelectCase, 0, len(ccs)+2)
		selectCases = append(selectCases, reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(ctx.Done()),
		})

		selectCases = append(selectCases, reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(idleNotificationC),
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
		} else if ind == 1 && a.noUndeliveredMsgs() {
			// idle notification
			a.logger.Log(logging.LevelDebug, "All nodes IDLE")
			select {
			case <-value.Interface().(chan struct{}):
				// just in case it was cancelled
				fmt.Println("All nodes IDLE - abort")
			default:
				// TODO: handle delayed to msgs or similar
				if len(a.delayedEvents) == 0 {
					return IdleWithoutDelayedEvents
				}
				delayedEvents := a.popRandomDelayedEvets()
				a.pushEvents(delayedEvents.NodeID, delayedEvents.Events, checker)
			}
			continue
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

			sourceNodeID := ccsNodeIds[ind-2]
			isByzantineSourceNode := sliceutil.Contains(a.nodeIds, sourceNodeID)
			isDeliverEvent := false

			a.logger.Log(logging.LevelDebug, "Info", "node", sourceNodeID, "byz", isByzantineSourceNode)

			// TODO: figure out how to actually do this filtering
			// hardcoding for now...

			// if not event of interest, just push and continue
			// eventType := reflect.TypeOf(event)
			// if _, ok := a.eventsOfInterest[eventType]; !ok {
			// 	fmt.Printf("Forwarding %v\n", event)
			// 	a.cortexCreepers[ind-1].PushEvents(stdtypes.ListOf(event))
			// 	continue
			// }

			a.logger.Log(logging.LevelDebug, "AAAAAAAAAAAAAAAAAAAAAAAAAAA")

			// hardcoded tmp solution
			switch evT := event.(type) {
			case *broadcastevents.BroadcastRequest:
				a.logger.Log(logging.LevelDebug, "BROADCAST BROADCAST BROADCAST")
			case *broadcastevents.Deliver:
				a.logger.Log(logging.LevelDebug, "DELIVER DELIVER DELIVER DELIVER")
			case *stdevents.SendMessage:
				a.logger.Log(logging.LevelDebug, "SEND SEND SEND SEND SEND")
				vc, err := evT.GetMetadata("vc")
				if err != nil {
					// TODO: how should I deal with errors in here?
					panic(err)
				}
				a.undeliveredMsgs[vc.(*vectorclock.VectorClock).String()] = struct{}{}
			case *stdevents.MessageReceived:
				a.logger.Log(logging.LevelDebug, "RECV RECV RECV RECV")
				isDeliverEvent = true
				vcSend, err := evT.GetMetadata("vcSend")
				if err != nil {
					// TODO: how should I deal with errors in here?
					panic(err)
				}
				delete(a.undeliveredMsgs, vcSend.(*vectorclock.VectorClock).String())
			default:
				a.logger.Log(logging.LevelDebug, "DEFAULT DEFAULT DEFAULT DEFAULT", "type", evT.ToString())
				a.pushEvents(sourceNodeID, stdtypes.ListOf(event), checker)
				continue
			}

			a.logger.Log(logging.LevelDebug, "BBBBBBBBBBBBBBBBBBBBBBBBBBB")

			// otherwise pick an action to apply to this event
			var action actions.Action
			if isByzantineSourceNode {
				action = a.byzantineActionSelector.SelectAction()
			} else if isDeliverEvent {
				action = a.networkActionSelector.SelectAction()
			} else {
				a.pushEvents(sourceNodeID, stdtypes.ListOf(event), checker)
				continue
			}

			actionLog, newEvents, delayedEvents, err := action(event, sourceNodeID, a.byzantineNodes)
			if err != nil {
				return err
			}
			a.logger.Log(logging.LevelDebug, "Action picked", "action", actionLog, "node", sourceNodeID, "isByz", isByzantineSourceNode, "isDeliver", isDeliverEvent)

			if delayedEvents != nil {
				a.delayedEvents = append(a.delayedEvents, delayedEvents...)
			}

			// ignore "" bc this signifies noop - not the best solution but works for now
			if actionLog != "" {
				a.actionTrace = append(a.actionTrace, ActionTraceEntry{
					node:      a.nodeIds[ind-2],
					actionLog: actionLog,
				})
			}

			// TODO: check for nil?
			for injectNodeID, injectEvents := range newEvents {
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
	puppeteerErr := make(chan error)
	//
	// TODO: use context for actual shutdown
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(caErr)
		nodesErr <- a.RunNodes(ctx)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(caErr)
		puppeteerErr <- puppeteer.Run(a.nodeInstances)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(caErr)
		// defer cancel() // cancel to stop nodes when adv had enough
		caErr <- a.RunCentralAdversary(maxEvents, maxHeartbeatsInactive, checker, ctx)
	}()

	if checker != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer close(checkerErr)
			checkerErr <- checker.Start()
		}()
	}

	var err error
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
		logStr += fmt.Sprintf("Node %s took action: %v\n\n", al.node, al.actionLog)
	}
	return logStr
}

func (a *Adversary) noUndeliveredMsgs() bool {
	return len(a.undeliveredMsgs) == 0
}

func (a *Adversary) popRandomDelayedEvets() actions.DelayedEvents {
	// TODO: HERE
	ind := rand.IntN(len(a.delayedEvents))
	de := a.delayedEvents[ind]
	// scrambeling order, ok bc we are picking random elements anyways
	a.delayedEvents[ind] = a.delayedEvents[len(a.delayedEvents)-1]
	a.delayedEvents = a.delayedEvents[:len(a.delayedEvents)-1]
	return de
}

type IdleNodesMonitor struct{}

func NewIdleNodesMonitor() *IdleNodesMonitor {
	return &IdleNodesMonitor{}
}

func (id *IdleNodesMonitor) Run(ctx context.Context, idleDetectionCs []chan chan struct{}, idleNotificationC chan<- chan struct{}) error {
	wg := sync.WaitGroup{}
	ctx, cancel := context.WithCancel(ctx)
	activeC := make(chan struct{})
	idleC := make(chan struct{})
	var noLongerInactiveC chan struct{}
	defer cancel()
	defer func() {
		for _, nc := range idleDetectionCs {
			close(nc)
		}
	}()

	wg.Add(len(idleDetectionCs))
	for _, idleDetectionC := range idleDetectionCs {
		go func() {
			defer wg.Done()
			var continueC chan struct{}
			for {
				select {
				case <-ctx.Done():
					return
				case <-continueC:
					activeC <- struct{}{}
					continueC = nil
				case cc := <-idleDetectionC:
					idleC <- struct{}{}
					continueC = cc
				}
			}
		}()
	}

	activeCount := len(idleDetectionCs)
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-activeC:
			if noLongerInactiveC != nil {
				close(noLongerInactiveC)
				noLongerInactiveC = nil
				fmt.Println("No longer idle")
			}
			activeCount++
		case <-idleC:
			activeCount--
			if activeCount == 0 {
				noLongerInactiveC = make(chan struct{})
				idleNotificationC <- noLongerInactiveC
			} else if activeCount < 0 {
				return es.Errorf("number of active nodes is negative")
			}
		}
	}
}
