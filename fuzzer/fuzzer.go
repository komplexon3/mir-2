package fuzzer

import (
	"context"
	"fmt"
	"io"
	"math/rand/v2"
	"os"
	"path"
	"strings"
	"time"

	"github.com/filecoin-project/mir/fuzzer/actions"
	centraladversay "github.com/filecoin-project/mir/fuzzer/centraladversary"
	"github.com/filecoin-project/mir/fuzzer/nodeinstance"
	"github.com/filecoin-project/mir/fuzzer/utils"
	"github.com/filecoin-project/mir/pkg/logging"
	es "github.com/go-errors/errors"
	"github.com/gosuri/uilive"

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

// jankily modified to get metrics
func (f *Fuzzer[T, S]) RunEvaluation(ctx context.Context, name string, runs int, timeout time.Duration, rand *rand.Rand, logLevel logging.LogLevel, resC chan map[string]checker.CheckerResult) (map[string]int, error) {
	err := os.MkdirAll(f.reportsDir, os.ModePerm)
	if err != nil {
		return nil, es.Errorf("failed to create fuzzing campaign report directory: %v", err)
	}

	reportFile, err := os.Create(path.Join(f.reportsDir, "report.txt"))
	if err != nil {
		return nil, es.Errorf("failed to create fuzzing campaign report file: %v", err)
	}
	_, err = fmt.Fprintf(reportFile, "count nodes: %d, byzantine nodes: %v\n\n", len(f.nodeConfigs), f.byzantineNodes)
	if err != nil {
		return nil, es.Errorf("failed to write to fuzzing campaign report file: %v", err)
	}
	defer reportFile.Close()

	updateWriter := NewUpdateWriter(runs, 100)
	updateWriter.Start()
	updateWriter.Update(0, 0, 0, 0, 0)

	propertiesHit := 0
	propertiesFirstHitInRun := make(map[string]int)

	countInteresting := 0
	countIdleExit := 0
	countTimeoutExit := 0
	countOtherExit := 0
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

			err = os.MkdirAll(runReportDir, os.ModePerm)
			if err != nil {
				return es.Errorf("failed to create report directory: %v", err)
			}

			fLogger, err := logging.NewFileLogger(logLevel, path.Join(runReportDir, "log.txt"))
			if err != nil {
				return es.Errorf("failed to create log file: %v", err)
			}

			defer fLogger.Stop() // close file
			logger := logging.Synchronize(fLogger)
			// logger := logging.NilLogger

			fuzzerRun, err := newFuzzerRun(runName, f.createNodeInstance, f.nodeConfigs, f.byzantineNodes, f.puppeteerSchedule, f.isInterestingEvent, f.byzantineActions, f.networkActions, runReportDir, f.createChecker, f.checkerParams, rand, logger)
			if err != nil {
				return err
			}

			results, runErr := fuzzerRun.Run(runCtx, runName, timeout, logger)
			// runErr is not necessaritly an err,err...
			switch runErr {
			case ErrorTimeout:
				countTimeoutExit++
			case centraladversay.ErrShutdownIdleWithoutDelayedEvents:
				countIdleExit++
			default:
				countOtherExit++
			}
			// no results means something really went wrong
			// TODO: should probably have fuzzerrun return 2 errors - 'real' errors and run return codes
			if results == nil {
				return err
			}

			resultStr := runName
			if !results.allPassed {
				countInteresting++
				resultStr += " - INTERESTING"

				// ignore timeout errors
				if results.exitReason != ErrorTimeout {
					for propertyName, result := range results.results {
						if result == checker.FAILURE {
							if _, ok := propertiesFirstHitInRun[propertyName]; !ok {
								propertiesFirstHitInRun[propertyName] = r + 1
								propertiesHit++
							}
						}
					}
				}

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

			resC <- results.results
			return nil
		}()
		if err != nil {
			// TODO: this will not work with the update writer, add a way to output this...
			fmt.Printf("Run %s failed: %v", runName, err)
			// don't fail/return, go on to next run
		}
		if (r+1)%10 == 0 {
			updateWriter.Update(r+1, countInteresting, countTimeoutExit, countIdleExit, countOtherExit)
		}
	}
	updateWriter.Stop()
	return propertiesFirstHitInRun, nil
}

func (f *Fuzzer[T, S]) Run(ctx context.Context, name string, runs int, timeout time.Duration, rand *rand.Rand, logLevel logging.LogLevel) (int, error) {
	err := os.MkdirAll(f.reportsDir, os.ModePerm)
	if err != nil {
		return 0, es.Errorf("failed to create fuzzing campaign report directory: %v", err)
	}

	reportFile, err := os.Create(path.Join(f.reportsDir, "report.txt"))
	if err != nil {
		return 0, es.Errorf("failed to create fuzzing campaign report file: %v", err)
	}
	defer reportFile.Close()

	updateWriter := NewUpdateWriter(runs, 100)
	updateWriter.Start()
	updateWriter.Update(0, 0, 0, 0, 0)

	countInteresting := 0
	countIdleExit := 0
	countTimeoutExit := 0
	countOtherExit := 0
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

			err = os.MkdirAll(runReportDir, os.ModePerm)
			if err != nil {
				return es.Errorf("failed to create report directory: %v", err)
			}

			fLogger, err := logging.NewFileLogger(logLevel, path.Join(runReportDir, "log.txt"))
			if err != nil {
				return es.Errorf("failed to create log file: %v", err)
			}

			defer fLogger.Stop() // close file
			logger := logging.Synchronize(fLogger)
			// logger := logging.NilLogger

			fuzzerRun, err := newFuzzerRun(runName, f.createNodeInstance, f.nodeConfigs, f.byzantineNodes, f.puppeteerSchedule, f.isInterestingEvent, f.byzantineActions, f.networkActions, runReportDir, f.createChecker, f.checkerParams, rand, logger)
			if err != nil {
				return err
			}

			results, runErr := fuzzerRun.Run(runCtx, runName, timeout, logger)
			// runErr is not necessaritly an err,err...
			switch runErr {
			case ErrorTimeout:
				countTimeoutExit++
			case centraladversay.ErrShutdownIdleWithoutDelayedEvents:
				countIdleExit++
			default:
				countOtherExit++
			}
			// no results means something really went wrong
			// TODO: should probably have fuzzerrun return 2 errors - 'real' errors and run return codes
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
			// TODO: this will not work with the update writer, add a way to output this...
			fmt.Printf("Run %s failed: %v", runName, err)
			// don't fail/return, go on to next run
		}
		updateWriter.Update(r+1, countInteresting, countTimeoutExit, countIdleExit, countOtherExit)
	}
	updateWriter.Stop()
	return countInteresting, nil
}

type UpdateWriter struct {
	startTime        time.Time
	secondLine       io.Writer
	thirdLine        io.Writer
	fourthLine       io.Writer
	writer           *uilive.Writer
	progressBarWidth int
	totalRuns        int
}

func NewUpdateWriter(totalRuns, progressbarWidth int) *UpdateWriter {
	writer := uilive.New()

	return &UpdateWriter{
		progressBarWidth: progressbarWidth,
		totalRuns:        totalRuns,
		writer:           writer,
		secondLine:       writer.Newline(),
		thirdLine:        writer.Newline(),
		fourthLine:       writer.Newline(),
	}
}

func (uw *UpdateWriter) Start() {
	uw.writer.Start()
	uw.startTime = time.Now()
}

func (uw *UpdateWriter) Stop() {
	uw.writer.Stop()
}

func (uw *UpdateWriter) Update(runsCompleted, interstingCases, timoutExits, idleExits, otherExits int) {
	runsPaddingNeeded := len(string(uw.totalRuns))
	fractionCompleted := float64(runsCompleted) / float64(uw.totalRuns)
	progressTicksCheck := int(float64(uw.progressBarWidth) * fractionCompleted)

	fmt.Fprintf(uw.writer, "Fuzzing Target - Total Runs: %*d, Completed Runs: %*d, Time Elapsed: %s\n", runsPaddingNeeded, uw.totalRuns, runsPaddingNeeded, runsCompleted, time.Since(uw.startTime).Round(time.Second))
	fmt.Fprintf(uw.secondLine, "[%s%s] %3d%%\n", strings.Repeat("=", progressTicksCheck), strings.Repeat("_", uw.progressBarWidth-progressTicksCheck), int(fractionCompleted*100))
	fmt.Fprintf(uw.thirdLine, "Timout Exits: %*d\t\tIdle Exits: %*d\t\tOther Exits: %*d\n", runsPaddingNeeded, timoutExits, runsPaddingNeeded, idleExits, runsPaddingNeeded, otherExits)
	fmt.Fprintf(uw.fourthLine, "Interesing Cases: %*d\n", runsPaddingNeeded, interstingCases)
}
