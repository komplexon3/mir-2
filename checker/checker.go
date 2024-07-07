package checker

import (
	"fmt"
	"runtime/debug"
	"sync"

	es "github.com/go-errors/errors"

	"github.com/filecoin-project/mir/checker/events"
	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/stdtypes"
)

type Property = modules.PassiveModule
type Properties = map[string]modules.PassiveModule

type checkerStatus int64

const (
	NOT_STARTED checkerStatus = iota
	RUNNING
	FINISHED
)

type Checker struct {
	properties []*property
	status     checkerStatus
	cDone      chan struct{}
}

type CheckerResult int64

const (
	NOT_READY CheckerResult = iota
	SUCCESS
	FAILURE
)

func (cr CheckerResult) String() string {
	switch cr {
	case NOT_READY:
		return "not ready"
	case SUCCESS:
		return "success"
	case FAILURE:
		return "failure"
	}

	// TODO: are golang enums really this weak?
	return "unknown"
}

type Decoder func([]byte) (stdtypes.Event, error)

type property struct {
	name         string
	runnerModule modules.PassiveModule
	eventChan    chan stdtypes.Event
	doneC        chan struct{}
	done         bool
	result       CheckerResult
}

func newProperty(name string, module modules.PassiveModule) *property {
	return &property{
		name,
		module,
		make(chan stdtypes.Event),
		make(chan struct{}, 1),
		false,
		NOT_READY,
	}

}

func NewChecker(properties Properties) (*Checker, error) {
	checker := &Checker{
		properties: make([]*property, 0, len(properties)),
		status:     NOT_STARTED,
		cDone:      make(chan struct{}),
	}

	for key, cond := range properties {
		if passiveModule, ok := cond.(modules.PassiveModule); ok {
			checker.properties = append(checker.properties, newProperty(string(key), passiveModule))
		} else {
			return nil, es.Errorf("module %s must be a passive module", key)
		}
	}

	return checker, nil
}

func (c *Checker) GetResults() (map[string]CheckerResult, error) {
	if c.status != FINISHED {
		return nil, fmt.Errorf("no results available, run analysis first")
	}

	results := make(map[string]CheckerResult, len(c.properties))

	for _, property := range c.properties {
		results[property.name] = property.result
	}

	return results, nil
}

func (c *Checker) Start() error {
	if c.status != NOT_STARTED {
		return es.Errorf("Cannot start checker. Checker is either finished or already running.")
	}

	if len(c.properties) == 0 {
		return fmt.Errorf("no properties registered")
	}

	c.status = RUNNING
	var wg sync.WaitGroup

	for _, p := range c.properties {
		wg.Add(1)
		go func(cc *property) {
			defer func() {
				cc.done = true
				close(cc.doneC)
				wg.Done()
			}()

			for e := range cc.eventChan {
				outEvents, err := safelyApplyEvents(cc.runnerModule, stdtypes.ListOf(e))
				if err != nil {
					fmt.Printf("property %s failed to apply events: %v", cc.name, err)
					return
				}

				iter := outEvents.Iterator()
				for ev := iter.Next(); ev != nil; ev = iter.Next() {
					switch ev.(type) {
					case *events.SuccessEvent:
						cc.result = SUCCESS
						return
					case *events.FailureEvent:
						cc.result = FAILURE
						return
					default:
						// TODO: add logging stuff
						fmt.Printf("property returned unsupported event: %T (supported: SuccessEvent, FailureEvent)", ev)
					}
				}
			}

		}(p)
	}

	wg.Wait()

	c.cDone <- struct{}{}
	c.status = FINISHED

	return nil
}

func (c *Checker) Stop() error {
	if c.status == FINISHED {
		return es.Errorf("Already finished")
	}

	if c.status == NOT_STARTED {
		return es.Errorf("Analysis has not started yet.")
	}

	// send done event to all -> initiate post processing if necessary
	for _, property := range c.properties {
		if !property.done {
			property.eventChan <- events.NewFinalEvent()
		}
	}

	// close all channels
	// will stop any property runner loop that hadn't terminated bc of success/failure yet
	for _, property := range c.properties {
		close(property.eventChan)
	}

	// wait on done signal from "runtime"
	<-c.cDone

	return nil
}

func (c *Checker) NextEvent(event stdtypes.Event) error {
	if c.status != RUNNING {
		return es.Errorf("Cannot process event. Checker is not running.")
	}

	for _, property := range c.properties {
		if !property.done {
			// TODO: this seeems very 'meh...' -> read up on 'closing' patterns
			select {
			case <-property.doneC:
			case property.eventChan <- event:
			}
		}
	}

	return nil
}

// "legacy" to not break some of the old tests
func (c *Checker) RunAnalysis(eventChan chan stdtypes.Event) error {
	errChan := make(chan error)
	go func() {
		errChan <- c.Start()
	}()

	for event := range eventChan {
		c.NextEvent(event)
	}

	c.Stop()

	err := <-errChan
	if err != nil {
		return err
	}

	return nil
}

// TODO - duplicated code...
func safelyApplyEvents(
	module modules.PassiveModule,
	events *stdtypes.EventList,
) (result *stdtypes.EventList, err error) {
	defer func() {
		if r := recover(); r != nil {
			if rErr, ok := r.(error); ok {
				err = es.Errorf("module panicked: %w\nStack trace:\n%s", rErr, string(debug.Stack()))
			} else {
				err = es.Errorf("module panicked: %v\nStack trace:\n%s", r, string(debug.Stack()))
			}
		}
	}()

	return module.ApplyEvents(events)
}
