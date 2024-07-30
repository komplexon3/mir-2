package centraladversay

import (
	"context"
	"reflect"

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
	globalEventsStreamOut   chan stdtypes.Event
	nodeIDs                 []stdtypes.NodeID
	byzantineNodes          []stdtypes.NodeID
}

func NewAdversary(
	nodeIDs []stdtypes.NodeID,
	cortexCreepers cortexcreeper.CortexCreepers,
	idleDetectionCs map[stdtypes.NodeID]chan chan struct{},
	byzantineWeightedActions []actions.WeightedAction,
	networkWeightedActions []actions.WeightedAction,
	byzantineNodes []stdtypes.NodeID,
	globalEventsStream chan stdtypes.Event,
	logger logging.Logger,
) (*Adversary, error) {
	// TODO: add some sanity checks to make sure that there are as many cortexCreepers, idleDetectionCs, etc. as there are nodeIDs
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
		idleNodesMonitor:        NewIdleNodesMonitor(idleDetectionCs),
		undeliveredMsgs:         make(map[string]struct{}),
		byzantineNodes:          byzantineNodes,
		nodeIDs:                 nodeIDs,
		actionTrace:             NewActionTrace(),
		delayedEventsManager:    NewDelayedEventsManager(),
		globalEventsStreamOut:   globalEventsStream,
		logger:                  logging.Decorate(logger, "CA - "),
	}, nil
}

func (a *Adversary) RunCentralAdversary(maxEvents int, ctx context.Context) error {
	// slice of cortex creepers ordered by nodeIds to easily reference them to the select cases
	ccsNodeIds := maputil.GetSortedKeys(a.cortexCreepers)
	ccs := maputil.GetValuesOf(a.cortexCreepers, ccsNodeIds)

	// TODO: actually handle errors from idleNodesMonitor
	go func() {
		err := a.idleNodesMonitor.Run(ctx)
		if err != nil {
			a.logger.Log(logging.LevelError, "idle nodes monitor failed", err)
		}
	}()
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
					continue
				default:
					// TODO: handle delayed to msgs or similar
					if a.delayedEventsManager.Empty() {
						a.logger.Log(logging.LevelDebug, "Should shut down...")
						return IdleWithoutDelayedEvents
					}
					a.logger.Log(logging.LevelDebug, "Delayed events ready for dispatch")
					delayedEvents := a.delayedEventsManager.PopRandomDelayedEvents()
					a.pushEvents(delayedEvents.NodeID, delayedEvents.Events)
				}
			} else {
				a.logger.Log(logging.LevelDebug, "Nodes Idle BUT msgs in transit", "msg count", len(a.undeliveredMsgs))
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
				a.pushEvents(sourceNodeID, stdtypes.ListOf(event))
				continue
			}

			// otherwise pick an action to apply to this event
			var action actions.Action
			if isByzantineSourceNode {
				action = a.byzantineActionSelector.SelectAction()
			} else if isDeliverEvent {
				action = a.networkActionSelector.SelectAction()
			} else {
				a.pushEvents(sourceNodeID, stdtypes.ListOf(event))
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
				a.pushEvents(injectNodeID, injectEvents)
			}
		}
	}
}

func (a *Adversary) pushEvents(nodeID stdtypes.NodeID, events *stdtypes.EventList) {
	if a.globalEventsStreamOut != nil {
		evtsIter := events.Iterator()
		for evt := evtsIter.Next(); evt != nil; evt = evtsIter.Next() {
			// add node id to metadata, also creates copy of event which we want
			evtWithNodeID, err := evt.SetMetadata("node", nodeID)
			if err != nil {
				// should probably be handles better but for now panic is okay -> fuzzer should be able to deal with panics anyways
				panic(es.Errorf("pushEvents failed to set metadata: %v", err))
			}
			a.globalEventsStreamOut <- evtWithNodeID
		}
	}
	cc := a.cortexCreepers[nodeID]
	cc.PushEvents(events)
}

// TODO: should take ctx
func (a *Adversary) RunExperiment(ctx context.Context, puppeteerSchedule []actions.DelayedEvents, maxEvents int) error {
	// NOTE: the whole structure of this thing is a complete mess

	a.delayedEventsManager.Push(puppeteerSchedule...)

	err := a.RunCentralAdversary(maxEvents, ctx)

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
