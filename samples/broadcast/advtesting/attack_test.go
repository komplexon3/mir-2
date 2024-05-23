package advtesting_test

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
	"testing"
	"time"

	broadcastevents "github.com/filecoin-project/mir/samples/broadcast/events"
	"github.com/filecoin-project/mir/stdevents"
	"github.com/stretchr/testify/assert"

	testmodules "github.com/filecoin-project/mir/samples/broadcast/properties"

	"github.com/filecoin-project/mir/adversary"
	"github.com/filecoin-project/mir/checker"
	"github.com/filecoin-project/mir/pkg/deploytest"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/trantor/types"
	"github.com/filecoin-project/mir/samples/broadcast/advtesting/nodeinstance"
	"github.com/filecoin-project/mir/samples/broadcast/advtesting/plugins"
	"github.com/filecoin-project/mir/samples/broadcast/advtesting/puppeteers"
	"github.com/filecoin-project/mir/stdtypes"
)

type TestConfig struct {
	NonByzantineNodes  []stdtypes.NodeID
	ByzantineNodes     []stdtypes.NodeID
	CreateNodeInstance adversary.NodeInstanceCreationFunc[nodeinstance.BroadcastNodeInstanceConfig]
	Plugin             *adversary.Plugin
	Puppeteer          adversary.Puppeteer
	Description        string // for report, should describe this config
	ExpectedResult     map[string]checker.CheckerResult
}

var rrPuppeteer, _ = puppeteers.NewRoundRobinPuppeteer(10, time.Second/100)

var tests = map[string]TestConfig{
	"no byz": {
		NonByzantineNodes:  []stdtypes.NodeID{"0", "1", "2", "3", "4"},
		ByzantineNodes:     []stdtypes.NodeID{},
		CreateNodeInstance: nodeinstance.CreateBroadcastNodeInstance,
		Plugin:             nil,
		Puppeteer:          rrPuppeteer,
		Description:        "0/5 byzantine nodes, round robin puppeteer (4, 0.2s), reference broadcast",
		ExpectedResult: map[string]checker.CheckerResult{
			"validity":    checker.SUCCESS,
			"integrity":   checker.SUCCESS,
			"consistency": checker.SUCCESS,
		},
	},
	"threshold no echo": {
		NonByzantineNodes:  []stdtypes.NodeID{"0", "1", "2"},
		ByzantineNodes:     []stdtypes.NodeID{"3", "4"},
		CreateNodeInstance: nodeinstance.CreateBroadcastNodeInstance,
		Plugin:             plugins.NewNoEcho(),
		Puppeteer:          rrPuppeteer,
		Description:        "2/5 byzantine nodes, round robin puppeteer (4, 0.2s), reference broadcast",
		ExpectedResult: map[string]checker.CheckerResult{
			"validity":    checker.FAILURE,
			"integrity":   checker.SUCCESS,
			"consistency": checker.SUCCESS,
		},
	},
}

func Test_NoByz(t *testing.T) {
	RunTest(t, "no byz")
}

func Test_ThresholdNoEchoAttack(t *testing.T) {
	RunTest(t, "threshold no echo")
}

func RunTest(t *testing.T, testName string) {
	test, ok := tests[testName]
	if !ok {
		panic(fmt.Sprintf("test with name %s not found", testName))
	}

	// create directory for report
	reportDir := fmt.Sprintf("./report_%s_%s", strings.Join(strings.Split(testName, " "), "_"), time.Now().Format("2006-01-02_15-04-05"))
	err := os.MkdirAll(reportDir, os.ModePerm)
	panicIfErr(err)

	t.Run(testName, func(tt *testing.T) {
		logger := logging.ConsoleWarnLogger
		allNodes := append(test.NonByzantineNodes, test.ByzantineNodes...)
		nodeWeights := make(map[stdtypes.NodeID]types.VoteWeight, len(allNodes))
		for i := range allNodes {
			id := stdtypes.NewNodeIDFromInt(i)
			nodeWeights[id] = "1"
		}
		nodeConfigs := make(map[stdtypes.NodeID]nodeinstance.BroadcastNodeInstanceConfig, len(allNodes))
		config := nodeinstance.BroadcastNodeInstanceConfig{NumberOfNodes: len(allNodes), FakeTransport: deploytest.NewFakeTransport(nodeWeights), LogPath: reportDir}
		for _, nodeID := range allNodes {
			nodeConfigs[nodeID] = config
		}
		adv, err := adversary.NewAdversary(test.CreateNodeInstance, nodeConfigs, test.ByzantineNodes, logger)
		panicIfErr(err)
		adv.RunExperiment(test.Plugin, test.Puppeteer)

		time.Sleep(time.Second)

		// property checking
		files := make([]string, 0, len(allNodes))
		for i := range allNodes {
			files = append(files, path.Join(reportDir, fmt.Sprintf("eventlog0_%d.gz", i)))
		}

		history := checker.GetEventsFromFileSortedByVectorClock([]func([]byte) (stdtypes.Event, error){broadcastevents.Deserialize, stdevents.Deserialize}, files...)
		eventChan := make(chan stdtypes.Event)

		systemConfig := &testmodules.SystemConfig{
			AllNodes:       allNodes,
			ByzantineNodes: test.ByzantineNodes,
		}

		m := map[stdtypes.ModuleID]modules.Module{
			"validity":    testmodules.NewValidity(*systemConfig, logger),
			"integrity":   testmodules.NewIntegrity(*systemConfig, logger),
			"consistency": testmodules.NewConsistency(*systemConfig, logger),
		}

		c, err := checker.NewChecker(m)
		panicIfErr(err)

		fmt.Println("Starting analysis")
		analysisTraceFile, err := os.Create(path.Join(reportDir, "trace.txt"))
		panicIfErr(err)
		defer analysisTraceFile.Close()
		go func() {
			for _, ele := range history {
				eventChan <- ele
				nodeId, err := ele.GetMetadata("node")
				if err != nil {
					nodeId = "?"
				}
				analysisTraceFile.WriteString(fmt.Sprintf("%s - %s\n", nodeId.(string), ele.ToString()))
			}
			close(eventChan)
		}()
		analysisTraceFile.Sync()

		err = c.RunAnalysis(eventChan)
		panicIfErr(err)

		results, _ := c.GetResults()

		resultStr := "Results:\n"
		for label, res := range results {
			resultStr = fmt.Sprintf("%s%s was %s - expected: %s\n", resultStr, label, res.String(), test.ExpectedResult[label])
			// NOTE: doesn't guarantee that the all expected res are checked... but that is ok for now
			assert.Equal(t, res, test.ExpectedResult[label])
		}
		fmt.Print(resultStr)
		// write into file

		// append description and commit

		commit := func() string {
			commit, err := exec.Command("git", "rev-parse", "HEAD").Output()
			if err != nil {
				return "no commit"
			}
			return string(commit)
		}()
		resultStr = fmt.Sprintf("%s\n\n%s\n\nDescription: %s\n\nCommit: %s", testName, resultStr, test.Description, commit)
		panicIfErr(os.WriteFile(path.Join(reportDir, "report.txt"), []byte(resultStr), 0644))

	})
}

func panicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}
