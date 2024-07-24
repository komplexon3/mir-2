package fuzzer2

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	es "github.com/go-errors/errors"

	"github.com/filecoin-project/mir/fuzzer2/actions"
	centraladversay "github.com/filecoin-project/mir/fuzzer2/centraladversary"
	"github.com/filecoin-project/mir/fuzzer2/nodeinstance"

	"github.com/filecoin-project/mir/checker"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/stdtypes"
)

const (
	MAX_EVENTS              = 200
	MAX_HEARTBEATS_INACTIVE = 10
)

type Fuzzer[T any] struct {
	ca                *centraladversay.Adversary
	puppeteerSchedule []actions.DelayedEvents
	reportsDir        string
	logger            logging.Logger
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
	adv, err := centraladversay.NewAdversary(createNodeInstance, nodeConfigs, byzantineActions, networkActions, byzantineNodes, logger)
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
	}, nil
}

func (f *Fuzzer[T]) Run(name string, propertyChecker *checker.Checker, maxEvents, maxInactiveHeartbeats int) error {
	reportDir := fmt.Sprintf("./report_%s_%s", time.Now().Format("2006-01-02_15-04-05"), strings.Join(strings.Split(name, " "), "_"))
	err := os.MkdirAll(reportDir, os.ModePerm)
	if err != nil {
		return es.Errorf("failed to create report directory: %v", err)
	}

	f.ca.RunExperiment(f.puppeteerSchedule, propertyChecker, maxEvents, maxInactiveHeartbeats)

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
	resultStr += f.ca.GetActionLogString()

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
