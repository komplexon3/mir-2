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
	"github.com/filecoin-project/mir/fuzzer2/puppeteer"

	"github.com/filecoin-project/mir/checker"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/stdtypes"
)

const (
	MAX_EVENTS              = 200
	MAX_HEARTBEATS_INACTIVE = 10
)

type Fuzzer[T any] struct {
	ca         *centraladversay.Adversary
	puppeteer  puppeteer.Puppeteer
	reportsDir string
	logger     logging.Logger
}

func NewFuzzer[T any](
	createNodeInstance nodeinstance.NodeInstanceCreationFunc[T],
	nodeConfigs nodeinstance.NodeConfigs[T],
	byzantineNodes []stdtypes.NodeID,
	puppeteer puppeteer.Puppeteer,
	byzantineActions []actions.WeightedAction,
	networkActions []actions.WeightedAction,
	reportsDir string,
	logger logging.Logger,
) (*Fuzzer[T], error) {
	adv, err := centraladversay.NewAdversary(createNodeInstance, nodeConfigs, byzantineActions, byzantineNodes, logger)
	if err != nil {
		return nil, es.Errorf("failed to create adversary: %v", err)
	}

	return &Fuzzer[T]{
		ca:         adv,
		puppeteer:  puppeteer,
		reportsDir: reportsDir,
		logger:     logger,
	}, nil
}

func (f *Fuzzer[T]) Run(name string, propertyChecker *checker.Checker, maxEvents, maxInactiveHeartbeats int) error {
	reportDir := fmt.Sprintf("./report_%s_%s", strings.Join(strings.Split(name, " "), "_"), time.Now().Format("2006-01-02_15-04-05"))
	err := os.MkdirAll(reportDir, os.ModePerm)
	if err != nil {
		return es.Errorf("failed to create report directory: %v", err)
	}

	f.ca.RunExperiment(f.puppeteer, propertyChecker, maxEvents, maxInactiveHeartbeats)

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
	if allPassed {
		os.RemoveAll(reportDir)
	}

	return nil
}
