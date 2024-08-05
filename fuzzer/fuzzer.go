package fuzzer

import (
	"context"
	"fmt"
	"math/rand/v2"
	"os"
	"path"
	"time"

	"github.com/filecoin-project/mir/fuzzer/actions"
	"github.com/filecoin-project/mir/fuzzer/nodeinstance"
	"github.com/filecoin-project/mir/fuzzer/utils"
	"github.com/filecoin-project/mir/pkg/logging"
	es "github.com/go-errors/errors"

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
	isInterestingEvent func(event stdtypes.Event) bool
	byzantineActions   []actions.WeightedAction
	networkActions     []actions.WeightedAction
}

func NewFuzzer[T, S any](
	createNodeInstance nodeinstance.NodeInstanceCreationFunc[T],
	nodeConfigs nodeinstance.NodeConfigs[T],
	byzantineNodes []stdtypes.NodeID,
	puppeteerSchedule []actions.DelayedEvents,
	isInterestingEvent func(event stdtypes.Event) bool,
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
		isInterestingEvent: isInterestingEvent,
		byzantineActions:   byzantineActions,
		networkActions:     networkActions,
		createChecker:      createChecker,
		checkerParams:      checkerParams,
		reportsDir:         reportsDir,
	}, nil
}

func (f *Fuzzer[T, S]) Run(ctx context.Context, name string, runs int, timeout time.Duration, rand *rand.Rand, logger logging.Logger) (int, error) {
	err := os.MkdirAll(f.reportsDir, os.ModePerm)
	if err != nil {
		return 0, es.Errorf("failed to create fuzzing campaign report directory: %v", err)
	}

	reportFile, err := os.Create(path.Join(f.reportsDir, "report.txt"))
	if err != nil {
		return 0, es.Errorf("failed to create fuzzing campaign report file: %v", err)
	}
	defer reportFile.Close()

	countInteresting := 0

	for r := range runs {
		// TODO: every individual run should contain it's errors -> add panic handling
		runName := fmt.Sprintf("%d-%s", r, name)
		runReportDir := path.Join(f.reportsDir, runName)
		runCtx, runCancel := context.WithCancel(ctx)
		err = func() (err error) {
			defer func() {
				if r := recover(); r != nil {
					// convert panic to an error
					err = fmt.Errorf("panic occurred: %v", r)
				}
				// just in case I didn't handle this nicely downstream
				runCancel()
			}()

			fuzzerRun, err := newFuzzerRun(runName, f.createNodeInstance, f.nodeConfigs, f.byzantineNodes, f.puppeteerSchedule, f.isInterestingEvent, f.byzantineActions, f.networkActions, runReportDir, f.createChecker, f.checkerParams, rand, logger)
			if err != nil {
				return err
			}

			results, _ := fuzzerRun.Run(runCtx, runName, timeout, logger)
			// runErr is not necessaritly an err,err...
			if results == nil {
				return err
			}

			resultStr := runName
			if !results.allPassed {
				countInteresting++
				resultStr += " - INTERESTING"
			}
			resultStr += fmt.Sprintf("\nExit Reason: %v\n\n", results.exitReason)
			for label, res := range results.results {
				resultStr += fmt.Sprintln(utils.FormatResult(label, res))
			}
			resultStr += "\n-----------------------------------\n\n"

			_, err = fmt.Fprint(reportFile, resultStr)
			if err != nil {
				return es.Errorf("failed to write to fuzzing campaign report file: %v", err)
			}
			err = reportFile.Sync()
			if err != nil {
				return es.Errorf("failed to flush fuzzing campaign report file to disk: %v", err)
			}

			return nil
		}()
		if err != nil {
			fmt.Printf("Run %s failed: %v", runName, err)
			// don't fail/return, go on to next run
		}
	}
	return countInteresting, nil
}
