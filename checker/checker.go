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

type Checker struct {
	conditions []*condition
	executed   bool // NOTE: is this actually providing any value?
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

type condition struct {
	name         string
	runnerModule modules.PassiveModule
	eventChan    chan stdtypes.Event
	doneC        chan struct{}
	done         bool
	result       CheckerResult
}

func newCondition(name string, module modules.PassiveModule) *condition {
	return &condition{
		name,
		module,
		make(chan stdtypes.Event),
		make(chan struct{}),
		false,
		NOT_READY,
	}

}

func NewChecker(conditions modules.Modules) (*Checker, error) {
	checker := &Checker{
		conditions: make([]*condition, 0, len(conditions)),
		executed:   false,
	}

	for key, cond := range conditions {
		if passiveModule, ok := cond.(modules.PassiveModule); ok {
			checker.conditions = append(checker.conditions, newCondition(string(key), passiveModule))
		} else {
			return nil, es.Errorf("module %s must be a passive module", key)
		}
	}

	return checker, nil
}

func (c *Checker) GetResults() (map[string]CheckerResult, error) {
	if !c.executed {
		return nil, fmt.Errorf("no results available, run analysis first")
	}

	results := make(map[string]CheckerResult, len(c.conditions))

	for _, condition := range c.conditions {
		results[condition.name] = condition.result
	}

	return results, nil
}

func (c *Checker) RunAnalysis(eventChan chan stdtypes.Event) error {
	if len(c.conditions) == 0 {
		return fmt.Errorf("no conditions registered")
	}

	var wg sync.WaitGroup

	for _, cc := range c.conditions {
		wg.Add(1)
		go func(cc *condition) {
			defer func() {
				cc.done = true
				close(cc.doneC)
				wg.Done()
			}()

			for e := range cc.eventChan {
				outEvents, err := safelyApplyEvents(cc.runnerModule, stdtypes.ListOf(e))
				if err != nil {
					fmt.Printf("condition %s failed to apply events: %v", cc.name, err)
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
						fmt.Printf("condition returned unsupported event: %T (supported: SuccessEvent, FailureEvent)", ev)
					}
				}
			}

		}(cc)
	}

	for e := range eventChan {
		for _, condition := range c.conditions {
			if !condition.done {
				// TODO: this seeems very 'meh...' -> read up on 'closing' patterns
				select {
				case <-condition.doneC:
				case condition.eventChan <- e:
				}
			}
		}
	}

	// send done event to all -> initiate post processing if necessary
	for _, condition := range c.conditions {
		if !condition.done {
			condition.eventChan <- events.NewFinalEvent()
		}
	}

	// close all channels
	// will stop any condition runner loop that hadn't terminated bc of success/failure yet
	for _, condition := range c.conditions {
		close(condition.eventChan)
	}

	// nomore events, stop all condition runners
	wg.Wait()
	c.executed = true

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
