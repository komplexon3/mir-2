package centraladversay

import (
	"context"
	"math/rand/v2"
	"reflect"
	"sync"

	"github.com/filecoin-project/mir/fuzzer/actions"
	"github.com/filecoin-project/mir/fuzzer/cortexcreeper"
	"github.com/filecoin-project/mir/fuzzer/utils"
	"github.com/filecoin-project/mir/pkg/idledetection"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/util/sliceutil"
	"github.com/filecoin-project/mir/stdevents"
	"github.com/filecoin-project/mir/stdtypes"

	es "github.com/go-errors/errors"
)

var (
	ErrorShutdown                       = es.Errorf("Shutdown")
	ErrShutdownIdleWithoutDelayedEvents = es.Errorf("Shutting down because all nodes are idle and there are delayed events to dispatch.")
)

type Adversary struct {
	byzantineActionSelector actions.Actions
	networkActionSelector   actions.Actions
	logger                  logging.Logger
	cortexCreepers          cortexcreeper.CortexCreepers
	undeliveredMsgs         map[string]struct{}
	actionTrace             *actionTrace
	delayedEventsManager    *delayedEventsManager
	globalEventsStreamOut   chan stdtypes.Event
	doneC                   chan struct{}
	isDoneC                 chan struct{}
	isInterestingEvent      func(event stdtypes.Event) bool
	nodeIDs                 []stdtypes.NodeID
	byzantineNodes          []stdtypes.NodeID
	idleDetectionCs         []chan idledetection.IdleNotification
}

func NewAdversary(
	nodeIDs []stdtypes.NodeID,
	cortexCreepers cortexcreeper.CortexCreepers,
	idleDetectionCs []chan idledetection.IdleNotification,
	isInterestingEvent func(event stdtypes.Event) bool,
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
		undeliveredMsgs:         make(map[string]struct{}),
		byzantineNodes:          byzantineNodes,
		nodeIDs:                 nodeIDs,
		actionTrace:             NewActionTrace(),
		delayedEventsManager:    NewDelayedEventsManager(),
		isInterestingEvent:      isInterestingEvent,
		globalEventsStreamOut:   globalEventsStream,
		doneC:                   make(chan struct{}),
		isDoneC:                 make(chan struct{}),
		idleDetectionCs:         idleDetectionCs,
		logger:                  logging.Decorate(logger, "CA - "),
	}, nil
}

func (a *Adversary) Stop() {
	close(a.doneC)
	<-a.isDoneC
}

func (a *Adversary) RunCentralAdversary(ctx context.Context) error {
	wg := sync.WaitGroup{}
	defer close(a.isDoneC)
	defer wg.Wait()

	continueCs := make([]chan idledetection.NoLongerIdleNotification, len(a.idleDetectionCs))
	isIdleStates := make([]bool, len(a.idleDetectionCs))

	var err error
	for err == nil {

		a.logger.Log(logging.LevelTrace, "current state", "undeliveredMsgs", len(a.undeliveredMsgs), "delayedEvents", len(a.delayedEventsManager.delayedEvents))
		var ackC chan struct{}
		extraSelectCases := 2
		// fan in events from different cortexCreepers + extra signals (e.g., ctx done)
		selectCases := make([]reflect.SelectCase, 0, 3*len(a.cortexCreepers)+extraSelectCases)
		selectReactions := make([]func(receivedVal reflect.Value, ok bool) error, 0, 3*len(a.cortexCreepers)+extraSelectCases)

		selectCases = append(selectCases, reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(ctx.Done()),
		})
		selectReactions = append(selectReactions, func(_ reflect.Value, _ bool) error {
			return ErrorShutdown
		})

		selectCases = append(selectCases, reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(a.doneC),
		})
		selectReactions = append(selectReactions, func(_ reflect.Value, _ bool) error {
			return ErrorShutdown
		})

		for nodeID, cc := range a.cortexCreepers {
			selectCases = append(selectCases, reflect.SelectCase{
				Dir:  reflect.SelectRecv,
				Chan: reflect.ValueOf(cc.GetEventsIn()),
			})

			selectReactions = append(selectReactions, func(value reflect.Value, ok bool) error {
				evtsAck := value.Interface().(*cortexcreeper.EventsAck)
				a.logger.Log(logging.LevelTrace, "processing evts", "node", nodeID, "events", evtsAck.Events)
				defer a.logger.Log(logging.LevelTrace, "done processing evts", "node", nodeID, "events", evtsAck.Events)
				defer close(evtsAck.Ack)
				dispatchEvents := make(map[stdtypes.NodeID]*stdtypes.EventList, len(a.nodeIDs))
				for _, nodeID := range a.nodeIDs {
					dispatchEvents[nodeID] = stdtypes.EmptyList()
				}
				elIterator := evtsAck.Events.Iterator()
				for event := elIterator.Next(); event != nil; event = elIterator.Next() {
					var err error

					isByzantineSourceNode := sliceutil.Contains(a.byzantineNodes, nodeID)
					isDeliverEvent := false

					// hardcoded tmp solution
					switch evT := event.(type) {
					case *stdevents.MessageReceived:
						// in case this is 'dropped' -> still gotta mark it
						isDeliverEvent = true
						var msgID interface{}
						msgID, err = evT.GetMetadata("msgID")
						if err != nil {
							return err
						}
						delete(a.undeliveredMsgs, msgID.(string))
					}

					if !a.isInterestingEvent(event) {
						// not an event of interest, just let it through
						switch eventT := event.(type) {
						case *stdevents.SendMessage:
							msgID := utils.RandomString(10)
							event, err = eventT.SetMetadata("msgID", msgID)
							if err != nil {
								return err
							}
							a.undeliveredMsgs[msgID] = struct{}{}
						}
						dispatchEvents[nodeID].PushBack(event)
						continue
					}

					// otherwise pick an action to apply to this event
					var action actions.Action
					if isByzantineSourceNode {
						action = a.byzantineActionSelector.SelectAction()
					} else if isDeliverEvent {
						action = a.networkActionSelector.SelectAction()
					} else {
						switch eventT := event.(type) {
						case *stdevents.SendMessage:
							msgID := utils.RandomString(10)
							event, err = eventT.SetMetadata("msgID", msgID)
							if err != nil {
								return err
							}
							a.undeliveredMsgs[msgID] = struct{}{}
						}
						a.logger.Log(logging.LevelTrace, "def - pushing evt", "evt", event.ToString())
						dispatchEvents[nodeID].PushBack(event)
						a.logger.Log(logging.LevelTrace, "done def - pushing evt", "evt", event.ToString())
						continue
					}

					actionLog, newEvents, delayedEvents, err := action(event, nodeID, a.byzantineNodes)
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
						a.actionTrace.Push(nodeID, actionLog)
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
						dispatchEvents[injectNodeID].PushBackList(newEvents)
					}
				}
				defer a.dispatchMultiNode(ctx, dispatchEvents)
				return nil
			})
		}

		for i, continueC := range continueCs {
			selectCases = append(selectCases, reflect.SelectCase{
				Chan: reflect.ValueOf(continueC),
				Dir:  reflect.SelectRecv,
			})
			selectReactions = append(selectReactions, func(_noLongerIdleNotification reflect.Value, ok bool) error {
				if !ok {
					return nil
				}
				noLongerIdleNotification := _noLongerIdleNotification.Interface().(idledetection.NoLongerIdleNotification)
				ackC = noLongerIdleNotification.Ack
				if !isIdleStates[i] {
					err = es.Errorf("node indicating not idle but is already registered as such (node index %d)", i)
				}
				isIdleStates[i] = false
				continueCs[i] = nil
				return nil
			})
		}

		for i, idleDetectionC := range a.idleDetectionCs {
			selectCases = append(selectCases, reflect.SelectCase{
				Chan: reflect.ValueOf(idleDetectionC),
				Dir:  reflect.SelectRecv,
			})
			selectReactions = append(selectReactions, func(_idleNotification reflect.Value, ok bool) error {
				if !ok {
					return nil
				}
				idleNotification := _idleNotification.Interface().(idledetection.IdleNotification)
				ackC = idleNotification.Ack
				if isIdleStates[i] {
					err = es.Errorf("node indicating idle but is already registered as such (node index %d)", i)
				}

				isIdleStates[i] = true
				continueCs[i] = idleNotification.NoLongerIdleC

				if !sliceutil.Contains(isIdleStates, false) {
					// handle idle case
					if a.noUndeliveredMsgs() {
						// idle notification
						a.logger.Log(logging.LevelDebug, "All nodes IDLE")
						// TODO: handle delayed to msgs or similar
						if a.delayedEventsManager.Empty() {
							a.logger.Log(logging.LevelDebug, "Should shut down...")
							return ErrShutdownIdleWithoutDelayedEvents
						}
						a.logger.Log(logging.LevelDebug, "Delayed events ready for dispatch")
						delayedEvents := a.delayedEventsManager.PopRandomDelayedEvents()
						delayedEventsListWithMsgID := stdtypes.EmptyList()
						evIterator := delayedEvents.Events.Iterator()
						for ev := evIterator.Next(); ev != nil; ev = evIterator.Next() {
							switch evT := ev.(type) {
							case *stdevents.SendMessage:
								msgID := utils.RandomString(10)
								var err error
								ev, err = evT.SetMetadata("msgID", msgID)
								if err != nil {
									return err
								}
								a.undeliveredMsgs[msgID] = struct{}{}
							}
							delayedEventsListWithMsgID.PushBack(ev)
						}
						a.dispatch(ctx, delayedEvents.NodeID, delayedEventsListWithMsgID)
					} else {
						a.logger.Log(logging.LevelDebug, "Nodes Idle BUT msgs in transit", "msg count", len(a.undeliveredMsgs))
					}
				}
				return nil
			})
		}

		// Select case pick
		ind, value, ok := reflect.Select(selectCases)
		err = selectReactions[ind](value, ok)
		if ackC != nil {
			close(ackC)
		}

	}
	return err
}

func (a *Adversary) dispatchMultiNode(ctx context.Context, events map[stdtypes.NodeID]*stdtypes.EventList) error {
	for nodeID, evts := range events {
		if err := a.dispatch(ctx, nodeID, evts); err != nil {
			return err
		}
	}
	return nil
}

func (a *Adversary) dispatch(ctx context.Context, nodeID stdtypes.NodeID, events *stdtypes.EventList) error {
	if events.Len() == 0 {
		// no events, return directly
		return nil
	}
	if a.globalEventsStreamOut != nil {
		evtsIter := events.Iterator()
		for evt := evtsIter.Next(); evt != nil; evt = evtsIter.Next() {
			// add node id to metadata, also creates copy of event which we want
			evtWithNodeID, err := evt.SetMetadata("node", nodeID)
			if err != nil {
				// should probably be handles better but for now panic is okay -> fuzzer should be able to deal with panics anyways
				return (es.Errorf("dispatch failed to set metadata: %v", err))
			}
			select {
			case a.globalEventsStreamOut <- evtWithNodeID:
			case <-a.doneC:
				// TODO: add 'shutdown' error?
				return nil
			case <-ctx.Done():
				return nil
			}
		}
		a.cortexCreepers[nodeID].PushEvents(ctx, events)
	}
	return nil
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
