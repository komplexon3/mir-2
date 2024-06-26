package testing

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/filecoin-project/mir/fuzzer2/actions"
	"github.com/filecoin-project/mir/fuzzer2/centraladversary"
	"github.com/filecoin-project/mir/fuzzer2/heartbeat"
	ni "github.com/filecoin-project/mir/fuzzer2/nodeinstance"
	"github.com/filecoin-project/mir/fuzzer2/puppeteer"
	broadcastevents "github.com/filecoin-project/mir/samples/broadcast/events"
	"github.com/filecoin-project/mir/stdevents"
	"github.com/stretchr/testify/assert"

	testmodules "github.com/filecoin-project/mir/samples/broadcast/properties"

	"github.com/filecoin-project/mir/checker"
	"github.com/filecoin-project/mir/pkg/deploytest"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/trantor/types"
	"github.com/filecoin-project/mir/samples/bcb-native/testing2/nodeinstance"
	"github.com/filecoin-project/mir/samples/bcb-native/testing2/puppeteers"
	"github.com/filecoin-project/mir/stdtypes"
)

const MAX_EVENTS = 200
const MAX_HEARTBEATS_INACTIVE = 10

type CATestConfig struct {
	Nodes              []stdtypes.NodeID
	Sender             stdtypes.NodeID
	CreateNodeInstance ni.NodeInstanceCreationFunc[nodeinstance.BcbNodeInstanceConfig]
	Puppeteer          puppeteer.Puppeteer
	Description        string // for report, should describe this config
	ExpectedResult     map[string]checker.CheckerResult
}

var catests = map[string]CATestConfig{
	"no byz": {
		Nodes:              []stdtypes.NodeID{"0", "1", "2", "3", "4"},
		Sender:             stdtypes.NodeID("0"),
		CreateNodeInstance: nodeinstance.CreateBcbNodeInstance,
		Puppeteer:          puppeteers.NewOneNodeBroadcast(stdtypes.NodeID("0")),
		Description:        "0/5 byzantine nodes, node 0 sender (non-byz)",
		ExpectedResult: map[string]checker.CheckerResult{
			"validity":    checker.SUCCESS,
			"integrity":   checker.SUCCESS,
			"consistency": checker.SUCCESS,
		},
	},
}

func Test_CANoByz(t *testing.T) {
	RunCATest(t, "no byz")
}

func RunCATest(t *testing.T, testName string) {
	test, ok := catests[testName]
	if !ok {
		panic(fmt.Sprintf("test with name %s not found", testName))
	}

	// create directory for report
	reportDir := fmt.Sprintf("./report_%s_%s", strings.Join(strings.Split(testName, " "), "_"), time.Now().Format("2006-01-02_15-04-05"))
	err := os.MkdirAll(reportDir, os.ModePerm)
	panicIfErr(err)

	t.Run(testName, func(_ *testing.T) {
		logger := logging.ConsoleWarnLogger
		nodeWeights := make(map[stdtypes.NodeID]types.VoteWeight, len(test.Nodes))
		for i := range test.Nodes {
			id := stdtypes.NewNodeIDFromInt(i)
			nodeWeights[id] = "1"
		}
		instanceUID := []byte("testing instance")
		nodeConfigs := make(map[stdtypes.NodeID]nodeinstance.BcbNodeInstanceConfig, len(test.Nodes))
		config := nodeinstance.BcbNodeInstanceConfig{InstanceUID: instanceUID, NumberOfNodes: len(test.Nodes), Leader: test.Sender, FakeTransport: deploytest.NewFakeTransport(nodeWeights), LogPath: reportDir}
		for _, nodeID := range test.Nodes {
			nodeConfigs[nodeID] = config
		}
		weightedActions := []actions.WeightedAction{actions.NewWeightedAction(func(e stdtypes.Event) (*stdtypes.EventList, error) { return stdtypes.ListOf(e), nil }, 1)}
		panicIfErr(err)
		adv, err := centraladversay.NewAdversary(test.CreateNodeInstance, nodeConfigs, weightedActions, logger)
		panicIfErr(err)

		adv.RunExperiment(test.Puppeteer, MAX_EVENTS, MAX_HEARTBEATS_INACTIVE)
		fmt.Println("Experiment done")

		// time.Sleep(time.Second)

		// property checking
		files := make([]string, 0, len(test.Nodes))
		for i := range test.Nodes {
			files = append(files, path.Join(reportDir, fmt.Sprintf("eventlog0_%d.gz", i)))
		}

		history := checker.GetEventsFromFileSortedByVectorClock([]func([]byte) (stdtypes.Event, error){broadcastevents.Deserialize, stdevents.Deserialize, heartbeat.Deserialize}, files...)
		eventChan := make(chan stdtypes.Event)

		systemConfig := &testmodules.SystemConfig{
			AllNodes:       test.Nodes,
			ByzantineNodes: []stdtypes.NodeID{},
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
