package centraladversay

import (
	"context"
	"reflect"
	"slices"
	"sync"

	"github.com/filecoin-project/mir/checker"
	"github.com/filecoin-project/mir/fuzzer/actions"
	"github.com/filecoin-project/mir/fuzzer/cortexcreeper"
	"github.com/filecoin-project/mir/fuzzer/utils"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/util/maputil"
	"github.com/filecoin-project/mir/pkg/util/sliceutil"
	broadcastevents "github.com/filecoin-project/mir/samples/bcb-native/events"
	"github.com/filecoin-project/mir/stdevents"
	"github.com/filecoin-project/mir/stdtypes"

	es "github.com/go-errors/errors"
)

var (
	MaxEventsShutdown        = es.Errorf("Shutting down because max events reached")
	IdleWithoutDelayedEvents = es.Errorf("Shutting down because all nodes are idle and there are no msgs to delay.")
)

type Adversary struct {
	byzantineActionSelector actions.Actions
	networkActionSelector   actions.Actions
	logger                  logging.Logger
	cortexCreepers          cortexcreeper.CortexCreepers
	idleNodesMonitor        *IdleNodesMonitor
	undeliveredMsgs         map[string]struct{}
	actionTrace             *actionTrace
	delayedEventsManager    *delayedEventsManager
	nodeIDs                 []stdtypes.NodeID
	byzantineNodes          []stdtypes.NodeID
}

func NewAdversary(
	nodeIDs []stdtypes.NodeID,
	cortexCreepers cortexcreeper.CortexCreepers,
	byzantineWeightedActions []actions.WeightedAction,
	networkWeightedActions []actions.WeightedAction,
	byzantineNodes []stdtypes.NodeID,
	logger logging.Logger,
) (*Adversary, error) {
	byzantineActionSelector, err := actions.NewRandomActions(append(byzantineWeightedActions, networkWeightedActions...))
	if err != nil {
		return nil, err
	}

	networkActionSelector, err := actions.NewRandomActions(networkWeightedActions)
	if err != nil {
		return nil, err
	}

	return &Adversary{
		byzantineActionSelector: byzantineActionSelector,
		networkActionSelector:   networkActionSelector,
		cortexCreepers:          cortexCreepers,
		idleNodesMonitor:        NewIdleNodesMonitor(),
		undeliveredMsgs:         make(map[string]struct{}),
		byzantineNodes:          byzantineNodes,
		nodeIDs:                 nodeIDs,
		actionTrace:             NewActionTrace(),
		delayedEventsManager:    NewDelayedEventsManager(),
		logger:                  logging.Decorate(logger, "CA - "),
	}, nil
}

func (a *Adversary) RunCentralAdversary(maxEvents int, checker *checker.Checker, ctx context.Context) error {
	// slice of cortex creepers ordered by nodeIds to easily reference them to the select cases
	ccsNodeIds := maputil.GetSortedKeys(a.cortexCreepers)
	ccs := maputil.GetValuesOf(a.cortexCreepers, ccsNodeIds)

	idleDetectionCs := sliceutil.Transform(ccs, func(_ int, cc *cortexcreeper.CortexCreeper) chan chan struct{} { return cc.IdleDetectionC })
	go a.idleNodesMonitor.Run(ctx, idleDetectionCs)
	defer a.idleNodesMonitor.Stop()
	eventCount := 0

	for {
		// fan in events from different cortexCreepers + idle detection + error
		selectCases := make([]reflect.SelectCase, 0, len(ccs)+2)
		selectCases = append(selectCases, reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(ctx.Done()),
		})

		selectCases = append(selectCases, reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(a.idleNodesMonitor.IdleNotificationC()),
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
		} else if ind == 1 {
			if a.noUndeliveredMsgs() {
				// idle notification
				a.logger.Log(logging.LevelDebug, "All nodes IDLE")
				select {
				case <-value.Interface().(chan struct{}):
					// just in case it was cancelled
					a.logger.Log(logging.LevelDebug, "All nodes IDLE - abort")
					return nil
				default:
					// TODO: handle delayed to msgs or similar
					if a.delayedEventsManager.Empty() {
						a.logger.Log(logging.LevelDebug, "Should shut down...")
						return IdleWithoutDelayedEvents
					}
					delayedEvents := a.delayedEventsManager.PopRandomDelayedEvents()
					a.pushEvents(delayedEvents.NodeID, delayedEvents.Events, checker)
				}
			}
			continue
		}

		// process this event
		elIterator := value.Interface().(*stdtypes.EventList).Iterator()
		for event := elIterator.Next(); event != nil; event = elIterator.Next() {
			var err error
			eventCount++

			if eventCount > maxEvents {
				return MaxEventsShutdown
			}

			sourceNodeID := ccsNodeIds[ind-2]
			isByzantineSourceNode := sliceutil.Contains(a.nodeIDs, sourceNodeID)
			isDeliverEvent := false

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
			switch evT := event.(type) {
			case *broadcastevents.BroadcastRequest:
			case *broadcastevents.Deliver:
			case *stdevents.SendMessage:
				msgID := utils.RandomString(10)
				event, err = evT.SetMetadata("msgID", msgID)
				if err != nil {
					return err
				}
				a.undeliveredMsgs[msgID] = struct{}{}
			case *stdevents.MessageReceived:
				isDeliverEvent = true
				msgID, err := evT.GetMetadata("msgID")
				if err != nil {
					return err
				}
				delete(a.undeliveredMsgs, msgID.(string))
			default:
				a.pushEvents(sourceNodeID, stdtypes.ListOf(event), checker)
				continue
			}

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

			if delayedEvents != nil {
				a.delayedEventsManager.Push(delayedEvents...)
			}

			// ignore "" bc this signifies noop - not the best solution but works for now
			if actionLog != "noop" && actionLog != "noop (network)" {
				a.actionTrace.Push(a.nodeIDs[ind-2], actionLog)
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
	cc := a.cortexCreepers[nodeID]
	cc.PushEvents(events)
}

func (a *Adversary) nodeIsPermittedToTakeByzantineAction(nodeId stdtypes.NodeID) bool {
	return slices.Contains(a.byzantineNodes, nodeId)
}

// TODO: should take ctx
func (a *Adversary) RunExperiment(puppeteerSchedule []actions.DelayedEvents, checker *checker.Checker, maxEvents int) error {
	ctx, cancel := context.WithCancel(context.Background())
	wg := sync.WaitGroup{}
	defer cancel()

	// NOTE: the whole structure of this thing is a complete mess

	a.delayedEventsManager.Push(puppeteerSchedule...)

	caErr := make(chan error)
	checkerErr := make(chan error)

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(caErr)
		select {
		case caErr <- a.RunCentralAdversary(maxEvents, checker, ctx):
		default:
		}
	}()

	if checker != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer close(checkerErr)
			select {
			case checkerErr <- checker.Start():
			default:
			}
		}()
	}

	var err error
	select {
	case err = <-caErr:
		if checker != nil {
			checker.Stop()
		}
		if err != nil {
			err = es.Errorf("Central Adversary error: %v", err)
		}
	case err = <-checkerErr:
		if err != nil {
			err = es.Errorf("Checker error: %v", err)
		}
	}

	cancel()
	wg.Wait()

	return err
}

func (a *Adversary) GetByzantineNodes() []stdtypes.NodeID {
	return a.byzantineNodes
}

func (a *Adversary) GetActionTrace() *actionTrace {
	return a.actionTrace
}

func (a *Adversary) noUndeliveredMsgs() bool {
	return len(a.undeliveredMsgs) == 0
}
