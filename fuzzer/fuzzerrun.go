package fuzzer

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/filecoin-project/mir/fuzzer/actions"
	centraladversay "github.com/filecoin-project/mir/fuzzer/centraladversary"
	"github.com/filecoin-project/mir/fuzzer/checker"
	"github.com/filecoin-project/mir/fuzzer/cortexcreeper"
	"github.com/filecoin-project/mir/fuzzer/nodeinstance"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/util/maputil"
	"github.com/filecoin-project/mir/stdtypes"
	es "github.com/go-errors/errors"
)

type fuzzerRun[T any] struct {
	name              string
	ca                *centraladversay.Adversary
	nodeInstances     nodeinstance.NodeInstances
	reportsDir        string
	puppeteerSchedule []actions.DelayedEvents
	propertyChecker   *checker.Checker
	eventsToCheckerC  chan stdtypes.Event
}

func newFuzzerRun[T any](
	name string,
	createNodeInstance nodeinstance.NodeInstanceCreationFunc[T],
	nodeConfigs nodeinstance.NodeConfigs[T],
	byzantineNodes []stdtypes.NodeID,
	puppeteerSchedule []actions.DelayedEvents,
	byzantineActions []actions.WeightedAction,
	networkActions []actions.WeightedAction,
	reportsDir string,
	properties checker.Properties,
	baseLogger logging.Logger,
) (*fuzzerRun[T], error) {
	// TODO: create node instances here and run them from here as well (instead of in the ca)

	// setting up node instances and cortex creepers
	nodeIDs := maputil.GetKeys(nodeConfigs)
	for _, byzNodeID := range byzantineNodes {
		if !slices.Contains(nodeIDs, byzNodeID) {
			return nil, es.Errorf("cannot use node %s as a byzantine node as there is no node config for this node", byzNodeID)
		}
	}

	nodeInstances := make(nodeinstance.NodeInstances, len(nodeIDs))
	cortexCreepers := make(cortexcreeper.CortexCreepers, len(nodeIDs))
	idleDetectionCs := make([]chan chan struct{}, 0, len(nodeIDs))

	for nodeID, config := range nodeConfigs {
		nodeLogger := logging.Decorate(baseLogger, string(nodeID)+" - ")
		cortexCreeper := cortexcreeper.NewCortexCreeper()
		cortexCreepers[nodeID] = cortexCreeper
		nodeInstance, err := createNodeInstance(nodeID, config, cortexCreeper, nodeLogger)
		if err != nil {
			return nil, es.Errorf("Failed to create node instance with id %s: %v", nodeID, err)
		}
		nodeInstances[nodeID] = nodeInstance
		idleDetectionCs = append(idleDetectionCs, nodeInstance.GetIdleDetectionC())
	}

	eventsToCheckerChan := make(chan stdtypes.Event)
	adv, err := centraladversay.NewAdversary(nodeIDs, cortexCreepers, idleDetectionCs, byzantineActions, networkActions, byzantineNodes, eventsToCheckerChan, baseLogger)
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

	propertyChecker, err := checker.NewChecker(properties)
	if err != nil {
		return nil, es.Errorf("fauled to create checker: %v", err)
	}

	return &fuzzerRun[T]{
		name:              name,
		ca:                adv,
		puppeteerSchedule: ps,
		reportsDir:        reportsDir,
		nodeInstances:     nodeInstances,
		propertyChecker:   propertyChecker,
		eventsToCheckerC:  eventsToCheckerChan,
	}, nil
}

// TODO: make fuzz run dir and fuzzerRun dirs inside, general fuzz run report file in fuzz run dir
func (r *fuzzerRun[T]) Run(ctx context.Context, name string, maxEvents int, logger logging.Logger) error {
	wg := sync.WaitGroup{}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	reportDir := fmt.Sprintf("./report_%s_%s", time.Now().Format("2006-01-02_15-04-05"), strings.Join(strings.Split(name, " "), "_"))
	err := os.MkdirAll(reportDir, os.ModePerm)
	if err != nil {
		return es.Errorf("failed to create report directory: %v", err)
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
		fmt.Println("NODES DONE NODES DONE NODES DONE NODES DONE NODES DONE ")
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(checkerErr)
		select {
		case checkerErr <- r.propertyChecker.Run(ctx, r.eventsToCheckerC):
		default:
		}
		fmt.Println("CHECKER DONE CHECKER DONE CHECKER DONE")
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(caErr)
		select {
		case caErr <- r.ca.RunExperiment(ctx, r.puppeteerSchedule, maxEvents):
		default:
		}
		fmt.Println("CA DONE CA DONE CA DONE CA DONE CA DONE ")
	}()

	select {
	case err = <-nodesErr:
		if err != nil {
			err = es.Errorf("Nodes error: %v", err)
		}
	case err = <-checkerErr:
		if err != nil {
			err = es.Errorf("Checker error: %v", err)
		}
	case err = <-caErr:
		if err != nil {
			err = es.Errorf("Central Adversary error: %v", err)
		}
	}

	// stop everything
	nodesRunner.Stop()
	r.propertyChecker.Stop()
	// TODO: add mechanism to stop ca
	// cancel()

	// wait until everything stopped
	wg.Wait()

	close(r.eventsToCheckerC)

	// TODO: make sure this call is actually safe -> should be the case I think
	results, _ := r.propertyChecker.GetResults()

	allPassed := true
	resultStr := fmt.Sprintf("Results: (%s)\n", reportDir)
	for label, res := range results {
		resultStr = fmt.Sprintf("%s%s: %s\n", resultStr, label, res.String())
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
	resultStr = fmt.Sprintf("%s\n\n%s\n\nCommit: %s\n\n", name, resultStr, commit)
	resultStr += r.ca.GetActionTrace().String()

	err = os.WriteFile(path.Join(reportDir, "report.txt"), []byte(resultStr), 0644)
	if err != nil {
		es.Errorf("failed to write report file: %v", err)
	}

	// delete report dir if all tests passed
	if allPassed {
		os.RemoveAll(reportDir)
	}

	return nil
}
