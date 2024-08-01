package centraladversay

import (
	"context"
	"math/rand/v2"
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
	IdleWithoutDelayedEvents = es.Errorf("Shutting down because all nodes are idle and there are delayed events to dispatch.")
)

type Adversary struct {
	byzantineActionSelector actions.Actions
	networkActionSelector   actions.Actions
	logger                  logging.Logger
	actionTrace             *actionTrace
	idleNodesMonitor        *IdleNodesMonitor
	undeliveredMsgs         map[string]struct{}
	cortexCreepers          cortexcreeper.CortexCreepers
	delayedEventsManager    *delayedEventsManager
	globalEventsStreamOut   chan stdtypes.Event
	doneC                   chan struct{}
	isDoneC                 chan struct{}
	nodeIDs                 []stdtypes.NodeID
	byzantineNodes          []stdtypes.NodeID
}

func NewAdversary(
	nodeIDs []stdtypes.NodeID,
	cortexCreepers cortexcreeper.CortexCreepers,
	idleDetectionCs []chan chan struct{},
	byzantineWeightedActions []actions.WeightedAction,
	networkWeightedActions []actions.WeightedAction,
	byzantineNodes []stdtypes.NodeID,
	globalEventsStream chan stdtypes.Event,
	rand *rand.Rand,
	logger logging.Logger,
) (*Adversary, error) {
	// TODO: add some sanity checks to make sure that there are as many cortexCreepers, idleDetectionCs, etc. as there are nodeIDs
	byzantineActionSelector, err := actions.NewRandomActions(append(byzantineWeightedActions, networkWeightedActions...), rand)
	if err != nil {
		return nil, err
	}

	networkActionSelector, err := actions.NewRandomActions(networkWeightedActions, rand)
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
		doneC:                   make(chan struct{}),
		isDoneC:                 make(chan struct{}),
		logger:                  logging.Decorate(logger, "CA - "),
	}, nil
}

func (a *Adversary) Stop() {
	close(a.doneC)
	<-a.isDoneC
}

func (a *Adversary) RunCentralAdversary(ctx context.Context) error {
	defer close(a.isDoneC)
	// slice of cortex creepers ordered by nodeIds to easily reference them to the select cases
	ccsNodeIds := maputil.GetSortedKeys(a.cortexCreepers)
	ccs := maputil.GetValuesOf(a.cortexCreepers, ccsNodeIds)

	// TODO: actually handle errors from idleNodesMonitor
	go func() {
		err := a.idleNodesMonitor.Run(ctx)
		if err != nil {
			a.logger.Log(logging.LevelError, "idle nodes monitor failed", "err", err)
		}
	}()
	defer a.idleNodesMonitor.Stop()

	for {

		a.logger.Log(logging.LevelTrace, "current state", "undeliveredMsgs", len(a.undeliveredMsgs), "delayedEvents", len(a.delayedEventsManager.delayedEvents))
		extraSelectCases := 3
		// fan in events from different cortexCreepers + extra signals (e.g., ctx done)
		selectCases := make([]reflect.SelectCase, 0, len(ccs)+extraSelectCases)

		selectCases = append(selectCases, reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(ctx.Done()),
		})

		selectCases = append(selectCases, reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(a.doneC),
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

		// Select case pick
		ind, value, _ := reflect.Select(selectCases)

		switch ind {
		case 0: // ctx done
			return nil
		case 1: // cDone shutdown signal
			return nil
		case 2: // idle signal
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
					a.pushEvents(ctx, delayedEvents.NodeID, delayedEvents.Events)
				}
			} else {
				a.logger.Log(logging.LevelDebug, "Nodes Idle BUT msgs in transit", "msg count", len(a.undeliveredMsgs))
			}
		default: // process node event loops
			// process this event
			elIterator := value.Interface().(*stdtypes.EventList).Iterator()
			for event := elIterator.Next(); event != nil; event = elIterator.Next() {
				var err error

				sourceNodeID := ccsNodeIds[ind-extraSelectCases]
				isByzantineSourceNode := sliceutil.Contains(a.nodeIDs, sourceNodeID)
				isDeliverEvent := false

				// TODO: figure out how to actually do this filtering

				// hardcoded tmp solution
				switch evT := event.(type) {
				case *broadcastevents.BroadcastRequest:
				case *broadcastevents.Deliver:
				case *stdevents.SendMessage:
				case *stdevents.MessageReceived:
					// in case this is 'dropped' -> still gotta mark it
					isDeliverEvent = true
					msgID, err := evT.GetMetadata("msgID")
					if err != nil {
						return err
					}
					delete(a.undeliveredMsgs, msgID.(string))
				default:
					a.pushEvents(ctx, sourceNodeID, stdtypes.ListOf(event))
					continue
				}

				// otherwise pick an action to apply to this event
				var action actions.Action
				if isByzantineSourceNode {
					action = a.byzantineActionSelector.SelectAction()
				} else if isDeliverEvent {
					action = a.networkActionSelector.SelectAction()
				} else {
					a.pushEvents(ctx, sourceNodeID, stdtypes.ListOf(event))
					continue
				}

				actionLog, newEvents, delayedEvents, err := action(event, sourceNodeID, a.byzantineNodes)
				if err != nil {
					return err
				}

				if delayedEvents != nil {
					// // mark delivery already
					// for _, de := range delayedEvents {
					// 	evIterator := de.Events.Iterator()
					// 	for ev := evIterator.Next(); ev != nil; ev = evIterator.Next() {
					// 		switch evT := ev.(type) {
					// 		case *stdevents.MessageReceived:
					// 			msgID, err := evT.GetMetadata("msgID")
					// 			if err != nil {
					// 				return err
					// 			}
					// 			delete(a.undeliveredMsgs, msgID.(string))
					// 		}
					// 	}
					// }
					a.delayedEventsManager.Push(delayedEvents...)
				}

				// ignore "" bc this signifies noop - not the best solution but works for now
				if actionLog != "noop" && actionLog != "noop (network)" {
					a.actionTrace.Push(a.nodeIDs[ind-extraSelectCases], actionLog)
				}

				// TODO: check for nil?
				for injectNodeID, injectEvents := range newEvents {
					newEvents := stdtypes.EmptyList()
					evIterator := injectEvents.Iterator()
					for ev := evIterator.Next(); ev != nil; ev = evIterator.Next() {
						switch evT := ev.(type) {
						case *stdevents.SendMessage:
							msgID := utils.RandomString(10)
							ev, err = evT.SetMetadata("msgID", msgID)
							if err != nil {
								return err
							}
							a.undeliveredMsgs[msgID] = struct{}{}

							// case *stdevents.MessageReceived:
							// 	msgID, err := evT.GetMetadata("msgID")
							// 	if err != nil {
							// 		return err
							// 	}
							// 	delete(a.undeliveredMsgs, msgID.(string))
						}
						newEvents.PushBack(ev)
					}
					a.pushEvents(ctx, injectNodeID, newEvents)
				}
			}
		}
	}
}

func (a *Adversary) pushEvents(ctx context.Context, nodeID stdtypes.NodeID, events *stdtypes.EventList) {
	if a.globalEventsStreamOut != nil {
		evtsIter := events.Iterator()
		for evt := evtsIter.Next(); evt != nil; evt = evtsIter.Next() {
			// add node id to metadata, also creates copy of event which we want
			evtWithNodeID, err := evt.SetMetadata("node", nodeID)
			if err != nil {
				// should probably be handles better but for now panic is okay -> fuzzer should be able to deal with panics anyways
				panic(es.Errorf("pushEvents failed to set metadata: %v", err))
			}
			select {
			case a.globalEventsStreamOut <- evtWithNodeID:
			case <-a.doneC:
			case <-ctx.Done():
			}
		}
	}
	cc := a.cortexCreepers[nodeID]
	cc.PushEvents(events)
}

// TODO: should take ctx
func (a *Adversary) RunExperiment(ctx context.Context, puppeteerSchedule []actions.DelayedEvents) error {
	// NOTE: the whole structure of this thing is a complete mess

	a.delayedEventsManager.Push(puppeteerSchedule...)

	err := a.RunCentralAdversary(ctx)

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
