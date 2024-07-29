package fuzzer

import (
	"context"
	"fmt"

	"github.com/filecoin-project/mir/fuzzer/actions"
	"github.com/filecoin-project/mir/fuzzer/nodeinstance"
	"github.com/filecoin-project/mir/pkg/logging"

	"github.com/filecoin-project/mir/fuzzer/checker"
	"github.com/filecoin-project/mir/stdtypes"
)

const (
	MAX_EVENTS = 200
)

type Fuzzer[T any] struct {
	createNodeInstance nodeinstance.NodeInstanceCreationFunc[T]
	nodeConfigs        nodeinstance.NodeConfigs[T]
	byzantineNodes     []stdtypes.NodeID
	puppeteerSchedule  []actions.DelayedEvents
	byzantineActions   []actions.WeightedAction
	networkActions     []actions.WeightedAction
	reportsDir         string
	properties         checker.Properties
}

func NewFuzzer[T any](
	createNodeInstance nodeinstance.NodeInstanceCreationFunc[T],
	nodeConfigs nodeinstance.NodeConfigs[T],
	byzantineNodes []stdtypes.NodeID,
	puppeteerSchedule []actions.DelayedEvents,
	byzantineActions []actions.WeightedAction,
	networkActions []actions.WeightedAction,
	reportsDir string,
	properties checker.Properties,
) (*Fuzzer[T], error) {
	return &Fuzzer[T]{
		createNodeInstance: createNodeInstance,
		nodeConfigs:        nodeConfigs,
		byzantineNodes:     byzantineNodes,
		puppeteerSchedule:  puppeteerSchedule,
		byzantineActions:   byzantineActions,
		networkActions:     networkActions,
		reportsDir:         reportsDir,
		properties:         properties,
	}, nil
}

func (f *Fuzzer[T]) Run(ctx context.Context, name string, runs int, maxEvents int, logger logging.Logger) error {
	for r := range runs {
		// TODO: every individual run should contain it's errors -> add panic handling
		runName := fmt.Sprintf("%s-%d", name, r)
		runCtx, runCancel := context.WithCancel(ctx)
		err := func() (err error) {
			defer func() {
				if r := recover(); r != nil {
					// convert panic to an error
					err = fmt.Errorf("panic occurred: %v", r)
				}
				// just in case I didn't hanle this nicely downstream
				runCancel()
			}()

			fuzzerRun, err := newFuzzerRun(runName, f.createNodeInstance, f.nodeConfigs, f.byzantineNodes, f.puppeteerSchedule, f.byzantineActions, f.networkActions, f.reportsDir, f.properties, logger)
			if err != nil {
				return err
			}

			err = fuzzerRun.Run(runCtx, runName, maxEvents, logger)
			if err != nil {
				return err
			}

			return nil
		}()
		if err != nil {
			fmt.Printf("Run %s failed: %v", runName, err)
			// don't fail/return, go on to next run
		}
	}
	return nil
}
