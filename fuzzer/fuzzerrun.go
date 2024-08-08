package fuzzer

import (
	"context"
	"fmt"
	"math/rand/v2"
	"os"
	"os/exec"
	"path"
	"slices"
	"sync"
	"time"

	"github.com/filecoin-project/mir/fuzzer/actions"
	centraladversay "github.com/filecoin-project/mir/fuzzer/centraladversary"
	"github.com/filecoin-project/mir/fuzzer/checker"
	"github.com/filecoin-project/mir/fuzzer/cortexcreeper"
	"github.com/filecoin-project/mir/fuzzer/nodeinstance"
	"github.com/filecoin-project/mir/fuzzer/utils"
	"github.com/filecoin-project/mir/pkg/deploytest"
	"github.com/filecoin-project/mir/pkg/idledetection"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/trantor/types"
	"github.com/filecoin-project/mir/pkg/util/maputil"
	"github.com/filecoin-project/mir/stdtypes"
	es "github.com/go-errors/errors"
)

var ErrorTimeout = fmt.Errorf("TIMEOUT")

type runResult struct {
	results    map[string]checker.CheckerResult
	exitReason error
	allPassed  bool
}

type fuzzerRun struct {
	ca                *centraladversay.Adversary
	nodeInstances     nodeinstance.NodeInstances
	propertyChecker   *checker.Checker
	eventsToCheckerC  chan stdtypes.Event
	name              string
	reportDir         string
	puppeteerSchedule []actions.DelayedEvents
}

func newFuzzerRun[T, S any](
	name string,
	createNodeInstance nodeinstance.NodeInstanceCreationFunc[T],
	nodeConfigs nodeinstance.NodeConfigs[T],
	byzantineNodes []stdtypes.NodeID,
	puppeteerSchedule []actions.DelayedEvents,
	isInterestingEvent func(event stdtypes.Event) bool,
	byzantineActions []actions.WeightedAction,
	networkActions []actions.WeightedAction,
	reportDir string,
	createChecker checker.CreateCheckerFunc[S],
	checkerParams S,
	rand *rand.Rand,
	baseLogger logging.Logger,
) (*fuzzerRun, error) {
	// TODO: create node instances here and run them from here as well (instead of in the ca)

	// setting up node instances and cortex creepers
	nodeIDs := maputil.GetKeys(nodeConfigs)
	for _, byzNodeID := range byzantineNodes {
		if !slices.Contains(nodeIDs, byzNodeID) {
			return nil, es.Errorf("cannot use node %s as a byzantine node as there is no node config for this node", byzNodeID)
		}
	}

	// setup transport
	nodeWeights := make(map[stdtypes.NodeID]types.VoteWeight, len(nodeIDs))
	for _, nodeID := range nodeIDs {
		nodeWeights[nodeID] = "1"
	}
	fakeTransport := deploytest.NewFakeTransport(nodeWeights)

	nodeInstances := make(nodeinstance.NodeInstances, len(nodeIDs))
	cortexCreepers := make(cortexcreeper.CortexCreepers, len(nodeIDs))
	idleDetectionCs := make([]chan idledetection.IdleNotification, 0, len(nodeIDs))

	for nodeID, config := range nodeConfigs {
		nodeLogger := logging.Decorate(baseLogger, string(nodeID)+" - ")
		cortexCreeper := cortexcreeper.NewCortexCreeper()
		cortexCreepers[nodeID] = cortexCreeper
		nodeInstance, err := createNodeInstance(nodeID, config, fakeTransport, cortexCreeper, reportDir, nodeLogger)
		if err != nil {
			return nil, es.Errorf("Failed to create node instance with id %s: %v", nodeID, err)
		}
		nodeInstances[nodeID] = nodeInstance
		idleDetectionCs = append(idleDetectionCs, nodeInstance.GetIdleDetectionC())
	}

	eventsToCheckerChan := make(chan stdtypes.Event)
	adv, err := centraladversay.NewAdversary(nodeIDs, cortexCreepers, idleDetectionCs, isInterestingEvent, byzantineActions, networkActions, byzantineNodes, eventsToCheckerChan, rand, baseLogger)
	if err != nil {
		return nil, es.Errorf("failed to create adversary: %v", err)
	}

	// TODO: not too nice but checher needs this info
	ps := make([]actions.DelayedEvents, 0, len(puppeteerSchedule))
	for _, de := range puppeteerSchedule {
		evtsWithNodeMetadata := stdtypes.EmptyList()
		evtsIterator := de.Events.Iterator()
		for evt := evtsIterator.Next(); evt != nil; evt = evtsIterator.Next() {
			evtWithNodeMetadata, err := evt.SetMetadata("node", de.NodeID) // creates copy of events as well
			if err != nil {
				return nil, err
			}
			evtsWithNodeMetadata.PushBack(evtWithNodeMetadata)
		}
		ps = append(ps, actions.DelayedEvents{NodeID: de.NodeID, Events: evtsWithNodeMetadata})
	}

	checkerLogger := logging.Decorate(baseLogger, "Checker - ")
	propertyChecker, err := createChecker(checkerParams, checkerLogger)
	if err != nil {
		return nil, es.Errorf("fauled to create checker: %v", err)
	}

	return &fuzzerRun{
		name:              name,
		ca:                adv,
		puppeteerSchedule: ps,
		reportDir:         reportDir,
		nodeInstances:     nodeInstances,
		propertyChecker:   propertyChecker,
		eventsToCheckerC:  eventsToCheckerChan,
	}, nil
}

// TODO: make fuzz run dir and fuzzerRun dirs inside, general fuzz run report file in fuzz run dir
func (r *fuzzerRun) Run(ctx context.Context, name string, timeout time.Duration, logger logging.Logger) (*runResult, error) {
	wg := sync.WaitGroup{}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	err := os.MkdirAll(r.reportDir, os.ModePerm)
	if err != nil {
		return nil, es.Errorf("failed to create report directory: %v", err)
	}

	nodesRunner := nodeinstance.NewNodesRunner(r.nodeInstances)
	// TODO: look into handeling this better
	nodesErr := make(chan error)
	checkerErr := make(chan error)
	caErr := make(chan error)
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(nodesErr)
		select {
		case nodesErr <- nodesRunner.Run(ctx, logger):
		default:
		}
		logger.Log(logging.LevelDebug, "Nodes shut down")
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(checkerErr)
		select {
		case checkerErr <- r.propertyChecker.Run(ctx, r.eventsToCheckerC):
		default:
		}
		logger.Log(logging.LevelDebug, "Checker shut down")
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(caErr)
		select {
		case caErr <- r.ca.RunExperiment(ctx, r.puppeteerSchedule):
		default:
		}
		logger.Log(logging.LevelDebug, "Central adversary shut down")
	}()

	var exitErr error
	select {
	case exitErr = <-nodesErr:
	case exitErr = <-checkerErr:
	case exitErr = <-caErr:
	case <-time.After(timeout):
		exitErr = ErrorTimeout
	}

	if exitErr != nil {
		// TODO: error is expected - filter for "expected exit errors" and give log an info enty instead
		logger.Log(logging.LevelError, "Error occured during fuzzer run:", "error", exitErr)
	}

	// stop everything
	logger.Log(logging.LevelDebug, "Sending Shutdown Nodes Runner")
	// causes 'idle but already indicating idle' error in idle nodes monitor during shutdown
	nodesRunner.Stop()
	logger.Log(logging.LevelDebug, "Sending Shutdown Central Adversary")
	r.ca.Stop()
	logger.Log(logging.LevelDebug, "Sending Shutdown Property Checker")
	r.propertyChecker.Stop()

	// wait until everything stopped
	wg.Wait()
	logger.Log(logging.LevelDebug, "Fuzzer routines shut down complete")

	close(r.eventsToCheckerC)

	// TODO: make sure this call is actually safe -> should be the case I think
	results, _ := r.propertyChecker.GetResults()

	allPassed := true
	resultStr := fmt.Sprintf("Results: (%s)\n", r.reportDir)
	for label, res := range results {
		resultStr += fmt.Sprintf("%s\n", utils.FormatResult(label, res))
		if res != checker.SUCCESS {
			allPassed = false
		}
	}

	if allPassed {
		logger.Log(logging.LevelInfo, "all properties passed")
	} else {
		logger.Log(logging.LevelInfo, resultStr)
	}

	commit := func() string {
		commit, err := exec.Command("git", "rev-parse", "HEAD").Output()
		if err != nil {
			return "no commit"
		}
		return string(commit)
	}()
	resultStr = fmt.Sprintf("%s\n\n%s\n\nCommit: %s\n\nExit reason: %v\n\n===================================\n\n", name, resultStr, commit, exitErr)
	resultStr += r.ca.GetActionTrace().String()

	err = os.WriteFile(path.Join(r.reportDir, "report.txt"), []byte(resultStr), 0644)
	if err != nil {
		es.Errorf("failed to write report file: %v", err)
		return nil, err
	}

	// delete report dir if all tests passed, and shutdown was 'normal'
	if allPassed && exitErr == centraladversay.ErrShutdownIdleWithoutDelayedEvents {
		os.RemoveAll(r.reportDir)
	}

	return &runResult{
		results:    results,
		exitReason: exitErr,
		allPassed:  allPassed,
	}, exitErr
}
