package fuzzer

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path"
	"slices"
	"strings"
	"time"

	es "github.com/go-errors/errors"

	"github.com/filecoin-project/mir/fuzzer/actions"
	centraladversay "github.com/filecoin-project/mir/fuzzer/centraladversary"
	"github.com/filecoin-project/mir/fuzzer/cortexcreeper"
	"github.com/filecoin-project/mir/fuzzer/nodeinstance"

	"github.com/filecoin-project/mir/checker"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/util/maputil"
	"github.com/filecoin-project/mir/stdtypes"
)

const (
	MAX_EVENTS = 200
)

type Fuzzer[T any] struct {
	logger            logging.Logger
	ca                *centraladversay.Adversary
	nodeInstances     nodeinstance.NodeInstances
	reportsDir        string
	puppeteerSchedule []actions.DelayedEvents
}

func NewFuzzer[T any](
	createNodeInstance nodeinstance.NodeInstanceCreationFunc[T],
	nodeConfigs nodeinstance.NodeConfigs[T],
	byzantineNodes []stdtypes.NodeID,
	puppeteerSchedule []actions.DelayedEvents,
	byzantineActions []actions.WeightedAction,
	networkActions []actions.WeightedAction,
	reportsDir string,
	logger logging.Logger,
) (*Fuzzer[T], error) {
	// TODO: create node instances here and run them from here as well (instead of in the ca)

	// setting up node instances and cortex creepers
	nodeIDs := maputil.GetKeys(nodeConfigs)
	for _, byzNodeID := range byzantineNodes {
		if !slices.Contains(nodeIDs, byzNodeID) {
			return nil, es.Errorf("cannot use node %s as a byzantine node as there is no node config for this node", byzNodeID)
		}
	}

	nodeInstances := make(nodeinstance.NodeInstances, len(nodeIDs))
	cortexCreepers := make(cortexcreeper.CortexCreepers, len(nodeConfigs))

	for nodeID, config := range nodeConfigs {
		nodeLogger := logging.Decorate(logger, string(nodeID)+" - ")
		cortexCreeper := cortexcreeper.NewCortexCreeper()
		cortexCreepers[nodeID] = cortexCreeper
		nodeInstance, err := createNodeInstance(nodeID, config, cortexCreeper, nodeLogger)
		if err != nil {
			return nil, es.Errorf("Failed to create node instance with id %s: %v", nodeID, err)
		}
		nodeInstances[nodeID] = nodeInstance
	}

	adv, err := centraladversay.NewAdversary(nodeIDs, cortexCreepers, byzantineActions, networkActions, byzantineNodes, logger)
	if err != nil {
		return nil, es.Errorf("failed to create adversary: %v", err)
	}

	// TODO: not too nice but checher needs this info
	logger.Log(logging.LevelDebug, "prepping delayed events")
	ps := make([]actions.DelayedEvents, 0, len(puppeteerSchedule))
	for _, de := range puppeteerSchedule {
		evtsWithNodeMetadata := stdtypes.EmptyList()
		evtsIterator := de.Events.Iterator()
		for evt := evtsIterator.Next(); evt != nil; evt = evtsIterator.Next() {
			evtWithNodeMetadata, err := evt.SetMetadata("node", de.NodeID)
			if err != nil {
				return nil, err
			}
			evtsWithNodeMetadata.PushBack(evtWithNodeMetadata)
		}
		ps = append(ps, actions.DelayedEvents{NodeID: de.NodeID, Events: evtsWithNodeMetadata})
	}
	logger.Log(logging.LevelDebug, "done - prepping delayed events")

	return &Fuzzer[T]{
		ca:                adv,
		puppeteerSchedule: ps,
		reportsDir:        reportsDir,
		logger:            logger,
		nodeInstances:     nodeInstances,
	}, nil
}

func (f *Fuzzer[T]) Run(ctx context.Context, name string, propertyChecker *checker.Checker, maxEvents int) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	reportDir := fmt.Sprintf("./report_%s_%s", time.Now().Format("2006-01-02_15-04-05"), strings.Join(strings.Split(name, " "), "_"))
	err := os.MkdirAll(reportDir, os.ModePerm)
	if err != nil {
		return es.Errorf("failed to create report directory: %v", err)
	}

	// TODO: look into handeling this better
	nodesErr := make(chan error)
	go func() {
		nodesErr <- nodeinstance.RunNodes(ctx, f.nodeInstances, f.logger)
		fmt.Println("DONEDONEDONEDONEDONEDONEDONEDONEDONE")
	}()

	err = f.ca.RunExperiment(f.puppeteerSchedule, propertyChecker, maxEvents)
	if err != nil {
		return err
	}
	fmt.Println("EXPERIMENT DONE EXPERIMENT DONE EXPERIMENT DONE ")
	cancel()

	results, _ := propertyChecker.GetResults()

	allPassed := true
	resultStr := fmt.Sprintf("Results: (%s)\n", reportDir)
	for label, res := range results {
		resultStr = fmt.Sprintf("%s%s: %s\n", resultStr, label, res.String())
		if res != checker.SUCCESS {
			allPassed = false
		}
	}

	if allPassed {
		f.logger.Log(logging.LevelInfo, "all properties passed")
	} else {
		f.logger.Log(logging.LevelInfo, resultStr)
	}

	commit := func() string {
		commit, err := exec.Command("git", "rev-parse", "HEAD").Output()
		if err != nil {
			return "no commit"
		}
		return string(commit)
	}()
	resultStr = fmt.Sprintf("%s\n\n%s\n\nCommit: %s\n\n", name, resultStr, commit)
	resultStr += f.ca.GetActionTrace().String()

	err = os.WriteFile(path.Join(reportDir, "report.txt"), []byte(resultStr), 0644)
	if err != nil {
		es.Errorf("failed to write report file: %v", err)
	}

	// delete report dir if all tests passed
	// if allPassed {
	// 	os.RemoveAll(reportDir)
	// }

	return nil
}
