/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mir

import (
	"context"
	"fmt"
	"reflect"
	"sync"

	es "github.com/go-errors/errors"

	"github.com/filecoin-project/mir/pkg/eventlog"
	"github.com/filecoin-project/mir/pkg/idledetection"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/util/maputil"
	"github.com/filecoin-project/mir/stdevents"
	"github.com/filecoin-project/mir/stdtypes"
)

var ErrStopped = fmt.Errorf("stopped at caller request")

// Node is the local instance of Mir and the application's interface to the mir library.
type Node struct {
	ID     stdtypes.NodeID // Protocol-level node ID
	Config *NodeConfig     // Node-level (protocol-independent) configuration, like buffer sizes, logging, ...

	// Incoming events to be processed by the node.
	// E.g., all modules' output events are written in this channel,
	// from where the Node processor reads and redistributes the events to their respective pendingEvents buffers.
	// External events are also funneled through this channel towards the pendingEvents buffers.
	eventsIn chan *stdtypes.EventList

	// During debugging, Events that would normally be inserted in the pendingEvents event buffer
	// (and thus inserted in the event loop) are written to this channel instead if it is not nil.
	// If this channel is nil, those Events are discarded.
	debugOut chan *stdtypes.EventList

	// All the modules that are part of the node.
	modules modules.Modules

	// Interceptor of processed events.
	// If not nil, every event is passed to the interceptor (by calling its Intercept method)
	// just before being processed.
	interceptor eventlog.Interceptor

	// A buffer for storing outstanding events that need to be processed by the node.
	// It contains a separate sub-buffer for each type of event.
	pendingEvents eventBuffer

	// Channels for routing work items between modules.
	// Whenever pendingEvents contains events, those events will be written (by the process() method)
	// to the corresponding channel in workChans. When processed by the corresponding module,
	// the result of that processing (also a list of events) will also be written
	// to the appropriate channel in workChans, from which the process() method moves them to the pendingEvents buffer.
	workChans workChans

	// Used to synchronize the exit of the node's worker go routines.
	workErrNotifier *workErrNotifier

	// When true, importing events from ActiveModules is disabled.
	// This value is set to true when the internal event buffer exceeds a certain threshold
	// and reset to false when the buffer is drained under a certain threshold.
	inputPaused bool

	inputPausedCond *sync.Cond

	// If set to true, the node is in debug mode.
	// Only events received through the Step method are applied.
	// Events produced by the modules are, instead of being applied,
	debugMode bool

	// This channel is closed by the Run and Debug methods on returning.
	// Closing of the channel indicates that the node has stopped and the Stop method can also return.
	stopped chan struct{}

	// This data structure holds statistics about recently dispatched events.
	dispatchStats eventDispatchStats

	// Each module has a stopwatch that is measuring the time the module spent processing events.
	stopwatches map[stdtypes.ModuleID]*Stopwatch

	// Lock guarding event processing stats.
	// This data is accessed concurrently by both the main event loop and the stats monitoring goroutine.
	// Access to the event buffers is also guarded by this lock,
	// since they need to be accessed when generating statistics.
	statsLock sync.Mutex

	// no events in queue signal
	modulePauseChans         pauseChans
	importerPauseChans       pauseChans
	inactiveNotificationChan chan idledetection.IdleNotification
	continueNotificationChan chan idledetection.NoLongerIdleNotification
}

// NewNode creates a new node with ID id.
// The config parameter specifies Node-level (protocol-independent) configuration, like buffer sizes, logging, ...
// The modules parameter must contain initialized, ready-to-use modules that the new Node will use.
func NewNode(
	id stdtypes.NodeID,
	config *NodeConfig,
	m modules.Modules,
	interceptor eventlog.Interceptor,
) (*Node, error) {
	// Check that a valid configuration has been provided.
	if err := config.Validate(); err != nil {
		return nil, es.Errorf("invalid node configuration: %w", err)
	}

	// Make sure that the logger can be accessed concurrently (if there is any logger).
	if config.Logger == nil {
		config.Logger = logging.NilLogger
	} else {
		config.Logger = logging.Synchronize(config.Logger)
	}

	stopwatches := make(map[stdtypes.ModuleID]*Stopwatch, len(m))
	for mID := range m {
		stopwatches[mID] = &Stopwatch{}
	}

	numberOfActiveModules := 0
	for _, module := range m {
		switch module.(type) {
		case modules.ActiveModule:
			numberOfActiveModules++
		}
	}

	// Return a new Node.
	return &Node{
		ID:     id,
		Config: config,

		eventsIn: make(chan *stdtypes.EventList, len(m)),
		debugOut: make(chan *stdtypes.EventList),

		workChans:   newWorkChans(m),
		modules:     m,
		interceptor: interceptor,

		pendingEvents:   newEventBuffer(m),
		workErrNotifier: newWorkErrNotifier(),

		inputPaused:     false,
		inputPausedCond: sync.NewCond(&sync.Mutex{}),

		dispatchStats: newDispatchStats(maputil.GetKeys(m)),
		stopwatches:   stopwatches,

		stopped: make(chan struct{}),

		modulePauseChans:         newModulePauseChans(m),
		importerPauseChans:       newImporterPauseChans(m),
		inactiveNotificationChan: nil,
		continueNotificationChan: nil,
	}, nil
}

func NewNodeWithIdleDetection(id stdtypes.NodeID, config *NodeConfig, m modules.Modules, interceptor eventlog.Interceptor) (*Node, chan idledetection.IdleNotification, error) {
	n, err := NewNode(id, config, m, interceptor)
	if err != nil {
		return nil, nil, err
	}

	n.inactiveNotificationChan = make(chan idledetection.IdleNotification)
	return n, n.inactiveNotificationChan, nil
}

// Debug runs the Node in debug mode.
// The Node will ony process events submitted through the Step method.
// All internally generated events will be ignored
// and, if the eventsOut argument is not nil, written to eventsOut instead.
// Note that if the caller supplies such a channel, the caller is expected to read from it.
// Otherwise, the Node's execution might block while writing to the channel.
func (n *Node) Debug(ctx context.Context, eventsOut chan *stdtypes.EventList) error {
	// When done, indicate to the Stop method that it can return.
	defer close(n.stopped)

	// Enable debug mode
	n.debugMode = true

	// Set up channel for outputting internal events
	n.debugOut = eventsOut

	// Start processing of events.
	return n.process(ctx)
}

// InjectEvents inserts a list of Events in the Node.
func (n *Node) InjectEvents(ctx context.Context, events *stdtypes.EventList) error {
	// Enqueue event in a work channel to be handled by the processing thread.
	select {
	case n.eventsIn <- events:
		return nil
	case <-n.workErrNotifier.ExitStatusC():
		return n.workErrNotifier.Err()
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Run starts the Node.
// It first adds an Init event to the work items, giving the modules the possibility
// to perform any initialization.
// Run then launches the processing of incoming messages, and internal events.
// The node stops when the ctx is canceled.
// The function call is blocking and only returns when the node stops.
func (n *Node) Run(ctx context.Context) error {
	// When done, indicate to the Stop method that it can return.
	defer close(n.stopped)

	// Submit the Init event to the modules.
	n.eventsIn <- createInitEvents(n.modules)

	// Start processing of events.
	return n.process(ctx)
}

// Stop stops the Node and returns only after the node has stopped, i.e., after a call to Run or Debug returns.
// If neither Run nor Debug has been called,
// Stop blocks until either Run or Debug is called by another goroutine and returns.
func (n *Node) Stop() {
	// Indicate to the event processing loop that it should stop.
	n.workErrNotifier.Fail(ErrStopped)

	// Wait until event processing stops.
	<-n.stopped
}

// Performs all internal work of the node,
// which mostly consists of routing events between the node's modules.
// Stops and returns when ctx is canceled.
func (n *Node) process(ctx context.Context) error { //nolint:gocyclo
	n.Config.Logger.Log(logging.LevelInfo, "node process started")
	defer n.Config.Logger.Log(logging.LevelInfo, "node process finished")

	var wg sync.WaitGroup // Synchronizes all the worker functions

	// Wait for all worker threads (started by n.StartModules()) to finish when processing is done.
	// Watch out! If process() terminates unexpectedly (e.g. by panicking), this might get stuck!
	defer wg.Wait()

	// Make sure that the event importing goroutines do not get stuck if input is paused due to full buffers.
	// This defer statement must go after waiting for the worker goroutines to finish (wg.Wait() above).
	// so it is executed before wg.Wait() (since defers are stacked). Otherwise, we get into a deadlock.
	defer n.resumeInput()

	// Periodically log statistics about dispatched events and the state of the event buffers.
	if n.Config.Stats.Period > 0 {
		wg.Add(1)
		go n.monitorStats(n.Config.Stats.Period, &wg)
	}

	// Start processing module events.
	n.startModules(ctx, &wg)

	if n.inactiveNotificationChan != nil {
		defer func() {
			if n.continueNotificationChan != nil {
				close(n.continueNotificationChan)
			}
		}()
		defer close(n.inactiveNotificationChan)

		// observe if node is active
		// TODO: cancel/shutdown
		wg.Add(1)
		go func() {
			defer wg.Done()
			n.runIdleDetector(ctx)
		}()

	}

	// This loop shovels events between the appropriate channels, until a stopping condition is satisfied.
	var returnErr error
	for returnErr == nil {

		// TODO: push this stuff into module/funtion

		if n.inactiveNotificationChan != nil && n.pendingEvents.totalEvents == 0 {
			n.runIdleDetector(ctx)
		}

		// Initialize slices of select cases and the corresponding reactions to each case being selected.
		selectCases := make([]reflect.SelectCase, 0)
		selectReactions := make([]func(receivedVal reflect.Value), 0)

		// If the context has been canceled, set the corresponding stopping value at the Node's WorkErrorNotifier,
		// making the processing stop when the WorkErrorNotifier's channel is selected the next time.
		selectCases = append(selectCases, reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(ctx.Done()),
		})
		selectReactions = append(selectReactions, func(_ reflect.Value) {
			// TODO: Use a different error here to distinguish this case from calling Node.Stop()
			n.workErrNotifier.Fail(ErrStopped)
		})

		// Add events produced by modules and debugger to the eventBuffer buffers and handle logical time.

		selectCases = append(selectCases, reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(n.eventsIn),
		})
		selectReactions = append(selectReactions, func(newEventsVal reflect.Value) {
			n.statsLock.Lock()
			defer n.statsLock.Unlock()

			n.noLongerIdle(ctx)

			newEvents := newEventsVal.Interface().(*stdtypes.EventList)
			// Intercept the (stripped of all follow-ups) events that were emitted.
			// This is only for debugging / diagnostic purposes.
			interceptedEvents := n.interceptEvents(newEvents)
			// Add the intercepted events to the modules' event buffers
			if err := n.pendingEvents.Add(interceptedEvents); err != nil {
				n.workErrNotifier.Fail(err)
			}

			// Keep track of the size of the input buffer.
			// When it exceeds the PauseInputThreshold, pause the input from active modules.
			if n.pendingEvents.totalEvents > n.Config.PauseInputThreshold {
				n.pauseInput()
			}
		})

		// If an error occurred, stop processing.

		selectCases = append(selectCases, reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(n.workErrNotifier.ExitC()),
		})
		selectReactions = append(selectReactions, func(_ reflect.Value) {
			returnErr = n.workErrNotifier.Err()
		})

		// For each generic event buffer in eventBuffer that contains events to be submitted to its corresponding module,
		// create a selectCase for writing those events to the module's work channel.
		n.statsLock.Lock()
		for moduleID, buffer := range n.pendingEvents.buffers {
			if buffer.Len() > 0 {

				eventBatch := buffer.Head(n.Config.MaxEventBatchSize)
				numEvents := eventBatch.Len()

				// Create case for writing in the work channel.
				selectCases = append(selectCases, reflect.SelectCase{
					Dir:  reflect.SelectSend,
					Chan: reflect.ValueOf(n.workChans[moduleID]),
					Send: reflect.ValueOf(eventBatch),
				})

				// Create a copy of moduleID to use in the reaction function.
				// If we used moduleID directly in the function definition, it would correspond to the loop variable
				// and have the same value for all cases after the loop finishes iterating.
				mID := moduleID

				// React to writing to a work channel by emptying the corresponding event buffer
				// (i.e., removing events just written to the channel from the buffer).
				selectReactions = append(selectReactions, func(_ reflect.Value) {
					n.statsLock.Lock()
					defer n.statsLock.Unlock()

					n.noLongerIdle(ctx)

					n.pendingEvents.buffers[mID].RemoveFront(numEvents)

					// Keep track of the size of the event buffer.
					// Whenever it drops below the ResumeInputThreshold, resume input.
					n.pendingEvents.totalEvents -= numEvents
					if n.pendingEvents.totalEvents <= n.Config.ResumeInputThreshold {
						n.resumeInput()
					}

					n.dispatchStats.AddDispatch(mID, numEvents)
				})
			}
		}
		n.statsLock.Unlock()

		// Choose one case from above and execute the corresponding reaction.

		chosenCase, receivedValue, _ := reflect.Select(selectCases)
		selectReactions[chosenCase](receivedValue)
	}

	n.workErrNotifier.SetExitStatus(nil, nil) // TODO: change this when statuses are implemented.
	return returnErr
}

func (n *Node) startModules(ctx context.Context, wg *sync.WaitGroup) {
	// The modules mostly read events from their respective channels in n.workChans,
	// process them correspondingly, and write the results (also represented as events) in the appropriate channels.
	for moduleID, module := range n.modules {

		// For each module, we start a worker function reads a single work item (EventList) and processes it.
		wg.Add(1)
		go func(mID stdtypes.ModuleID, m modules.Module, workChan chan *stdtypes.EventList, pauseChan chan chan struct{}) {
			n.Config.Logger.Log(logging.LevelInfo, "module started", "ID", mID.String())
			defer n.Config.Logger.Log(logging.LevelInfo, "module finished", "ID", mID.String())
			defer wg.Done()

			// Create a context that is passed to the module event application function
			// and canceled when module processing is stopped (i.e. this function returns).
			// This is to make sure that the event processing is properly aborted if needed.
			processingCtx, cancelProcessing := context.WithCancel(ctx)
			defer cancelProcessing()

			continueProcessing := true
			var err error

			for continueProcessing {
				if n.debugMode {
					// In debug mode, all produced events are routed to the debug output.
					continueProcessing, err = n.processModuleEvents(
						processingCtx,
						mID,
						m,
						workChan,
						n.debugOut,
						pauseChan,
						n.stopwatches[mID],
					)
				} else {
					// During normal operation, feed all produced events back into the event loop.
					continueProcessing, err = n.processModuleEvents(
						processingCtx,
						mID,
						m,
						workChan,
						n.eventsIn,
						pauseChan,
						n.stopwatches[mID],
					)
				}
				if err != nil {
					n.workErrNotifier.Fail(es.Errorf("could not process PassiveModule (%v) events: %w", mID, err))
					return
				}
			}
		}(moduleID, module, n.workChans[moduleID], n.modulePauseChans[moduleID])

		// Depending on the module type (and the way output events are communicated back to the node),
		// start a goroutine importing the modules' output events
		switch m := module.(type) {
		case modules.PassiveModule:
			// Nothing else to be done for a PassiveModule
		case modules.ActiveModule:
			// Start a goroutine to import the ActiveModule's output events to workItemInput.
			wg.Add(1)
			inactiveNotificationChan := n.importerPauseChans[moduleID]
			go func() {
				defer wg.Done()
				if n.debugMode {
					// In debug mode, all produced events are routed to the debug output.
					n.importEvents(ctx, m.EventsOut(), n.debugOut, inactiveNotificationChan)
				} else {
					// During normal operation, feed all produced events back into the event loop.
					n.importEvents(ctx, m.EventsOut(), n.eventsIn, inactiveNotificationChan)
				}
			}()
		default:
			n.workErrNotifier.Fail(es.Errorf("unknown module type: %T", m))
		}
	}
}

// importEvents reads events from eventSource and writes them to the eventSink until
// - eventSource is closed or
// - ctx is canceled or
// - an error occurred in the Node and was announced through the Node's workErrorNotifier.
func (n *Node) importEvents(
	ctx context.Context,
	eventSource <-chan *stdtypes.EventList,
	eventSink chan<- *stdtypes.EventList,
	pause <-chan chan struct{},
) {
	for {

		n.waitForInputEnabled()

		// First, try to read events from the input.
		select {
		case continueChan := <-pause:
			n.Config.Logger.Log(logging.LevelTrace, "pausing importer")
			select {
			case <-continueChan:
			case <-ctx.Done():
				return
			case <-n.workErrNotifier.ExitC():
				return
			}
			n.Config.Logger.Log(logging.LevelTrace, "continue importer")
		case newEvents, ok := <-eventSource:

			// Return if input channel has been closed
			if !ok {
				return
			}

			// Skip writing events if there is no event sink.
			if eventSink == nil {
				continue
			}

			// If input events have been read, try to write them to the Node's central input channel.
			select {
			case eventSink <- newEvents:
			case <-ctx.Done():
				return
			case <-n.workErrNotifier.ExitC():
				return
			}

		case <-ctx.Done():
			return
		case <-n.workErrNotifier.ExitC():
			return
		}
	}
}

// If the interceptor module is present, passes events to it. Otherwise, does nothing.
// If an error occurs passing events to the interceptor, notifies the node by means of the workErrorNotifier.
// The interceptor has the ability to modify the EventList.
// The events returned by the interceptor are the events actually delivered to the system's modules
// Note: The passed Events should be free of any follow-up Events,
// as those will be intercepted separately when processed.
// Make sure to call the Strip method of the EventList before passing it to interceptEvents.
func (n *Node) interceptEvents(events *stdtypes.EventList) *stdtypes.EventList {
	// ATTENTION: n.interceptor is an interface type. If it is assigned the nil value of a concrete type,
	// this condition will evaluate to true, and Intercept(events) will be called on nil.
	// The implementation of the concrete type must make sure that calling Intercept even on the nil value
	// does not cause any problems.
	// For more explanation, see https://mangatmodi.medium.com/go-check-nil-interface-the-right-way-d142776edef1
	var err error
	if n.interceptor != nil {
		events, err = n.interceptor.Intercept(events)
		if err != nil {
			n.workErrNotifier.Fail(err)
		}
	}
	return events
}

func (n *Node) pauseInput() {
	n.inputPausedCond.L.Lock()
	n.inputPaused = true
	n.inputPausedCond.L.Unlock()
}

func (n *Node) resumeInput() {
	n.inputPausedCond.L.Lock()
	n.inputPaused = false
	n.inputPausedCond.Broadcast()
	n.inputPausedCond.L.Unlock()
}

func (n *Node) waitForInputEnabled() {
	n.inputPausedCond.L.Lock()
	for n.inputPaused {
		n.inputPausedCond.Wait()
	}
	n.inputPausedCond.L.Unlock()
}

func (n *Node) inputIsPaused() bool {
	n.inputPausedCond.L.Lock()
	defer n.inputPausedCond.L.Unlock()
	return n.inputPaused
}

func (n *Node) runIdleDetector(ctx context.Context) {
	pauseWg := sync.WaitGroup{}
	continueC := make(chan struct{})
	pauseCounter := make(chan struct{}, len(n.modulePauseChans)+len(n.importerPauseChans))
	defer close(pauseCounter)
	defer pauseWg.Wait()
	defer close(continueC) // unpausing modules/importers

	// no events in queues atm
	// TODO: don't (ab)use cancel for regular control flow - add "stop"/"done" channel
	for _, pauseC := range n.modulePauseChans {
		pauseWg.Add(1)
		go func() {
			defer pauseWg.Done()
			n.idleDetectorPauseWorker(ctx, continueC, pauseCounter, pauseC)
		}()
	}

	for _, pauseC := range n.importerPauseChans {
		pauseWg.Add(1)
		go func() {
			defer pauseWg.Done()
			n.idleDetectorPauseWorker(ctx, continueC, pauseCounter, pauseC)
		}()
	}

	count := 0

	for {
		select {
		case <-pauseCounter:
			count++
		case <-ctx.Done():
			// no need to restart modules as they should already be running
			return
		case <-n.workErrNotifier.ExitC():
			return
		}
		if len(n.eventsIn) > 0 {
			// abort
			return
		} else if count == len(n.modulePauseChans)+len(n.importerPauseChans) {
			// all modules are now inactive and there are no events in eventsIn
			break
		}
	}

	// all modules paused at this point, no events in eventsIn
	// TODO:  events in check
	n.Config.Logger.Log(logging.LevelDebug, "IDLE")
	idleNotification := idledetection.NewIdleNotification()
	n.continueNotificationChan = idleNotification.NoLongerIdleC
	select {
	case <-ctx.Done():
		return
	case <-n.workErrNotifier.ExitC():
		return
	case n.inactiveNotificationChan <- idleNotification:
	}
	select {
	case <-ctx.Done():
		return
	case <-n.workErrNotifier.ExitC():
		return
	case <-idleNotification.Ack:
	}
}

func (n *Node) idleDetectorPauseWorker(ctx context.Context, continueC chan struct{}, pauseCounter chan struct{}, pauseC chan chan struct{}) {
	select {
	case <-ctx.Done():
	case <-continueC:
	case pauseC <- continueC:
		// if this (or another) module created events during the pausing procedure, abort
		select {
		case <-ctx.Done():
		case pauseCounter <- struct{}{}:
		case <-continueC:
		}
	}
}

func (n *Node) noLongerIdle(ctx context.Context) {
	if n.continueNotificationChan != nil {
		n.Config.Logger.Log(logging.LevelDebug, "No longer IDLE")
		noLongerIdleNotification := idledetection.NewNoLongerIdleNotification()
		select {
		case n.continueNotificationChan <- noLongerIdleNotification:
		case <-n.workErrNotifier.ExitC():
		case <-ctx.Done():
		}
		select {
		case <-noLongerIdleNotification.Ack:
		case <-n.workErrNotifier.ExitC():
		case <-ctx.Done():
		}
		close(n.continueNotificationChan)
		n.continueNotificationChan = nil
	}
}

func createInitEvents(m modules.Modules) *stdtypes.EventList {
	initEvents := stdtypes.EmptyList()
	for moduleID := range m {
		initEvents.PushBack(stdevents.NewInit(moduleID))
	}
	return initEvents
}

func newImporterPauseChans(m modules.Modules) pauseChans {
	pcs := make(map[stdtypes.ModuleID]chan chan struct{})

	for moduleID, module := range m {
		switch module.(type) {
		case modules.ActiveModule:
			pcs[moduleID] = make(chan chan struct{})
		}
	}

	return pcs
}
