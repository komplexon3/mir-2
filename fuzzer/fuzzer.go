package fuzzer

import (
	"context"
	"fmt"
	"math/rand/v2"
	"path"
	"time"

	"github.com/filecoin-project/mir/fuzzer/actions"
	"github.com/filecoin-project/mir/fuzzer/nodeinstance"
	"github.com/filecoin-project/mir/pkg/logging"

	"github.com/filecoin-project/mir/fuzzer/checker"
	"github.com/filecoin-project/mir/stdtypes"
)

const (
	MAX_EVENTS = 200
)

type Fuzzer[T, S any] struct {
	checkerParams      S
	createNodeInstance nodeinstance.NodeInstanceCreationFunc[T]
	nodeConfigs        nodeinstance.NodeConfigs[T]
	createChecker      checker.CreateCheckerFunc[S]
	reportsDir         string
	byzantineNodes     []stdtypes.NodeID
	puppeteerSchedule  []actions.DelayedEvents
	byzantineActions   []actions.WeightedAction
	networkActions     []actions.WeightedAction
}

func NewFuzzer[T, S any](
	createNodeInstance nodeinstance.NodeInstanceCreationFunc[T],
	nodeConfigs nodeinstance.NodeConfigs[T],
	byzantineNodes []stdtypes.NodeID,
	puppeteerSchedule []actions.DelayedEvents,
	byzantineActions []actions.WeightedAction,
	networkActions []actions.WeightedAction,
	createChecker checker.CreateCheckerFunc[S],
	checkerParams S,
	reportsDir string,
) (*Fuzzer[T, S], error) {
	return &Fuzzer[T, S]{
		createNodeInstance: createNodeInstance,
		nodeConfigs:        nodeConfigs,
		byzantineNodes:     byzantineNodes,
		puppeteerSchedule:  puppeteerSchedule,
		byzantineActions:   byzantineActions,
		networkActions:     networkActions,
		createChecker:      createChecker,
		checkerParams:      checkerParams,
		reportsDir:         reportsDir,
	}, nil
}

func (f *Fuzzer[T, S]) Run(ctx context.Context, name string, runs int, timeout time.Duration, rand *rand.Rand, logger logging.Logger) error {
	for r := range runs {
		// TODO: every individual run should contain it's errors -> add panic handling
		runName := fmt.Sprintf("%d-%s", r, name)
		runReportDir := path.Join(f.reportsDir, runName)
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

			fuzzerRun, err := newFuzzerRun(runName, f.createNodeInstance, f.nodeConfigs, f.byzantineNodes, f.puppeteerSchedule, f.byzantineActions, f.networkActions, runReportDir, f.createChecker, f.checkerParams, rand, logger)
			if err != nil {
				return err
			}

			err = fuzzerRun.Run(runCtx, runName, timeout, logger)
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
