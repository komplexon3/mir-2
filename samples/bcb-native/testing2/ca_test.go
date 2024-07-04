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
	centraladversay "github.com/filecoin-project/mir/fuzzer2/centraladversary"
	ni "github.com/filecoin-project/mir/fuzzer2/nodeinstance"
	"github.com/filecoin-project/mir/fuzzer2/puppeteer"
	broadcastevents "github.com/filecoin-project/mir/samples/broadcast/events"
	"github.com/stretchr/testify/assert"

	"github.com/filecoin-project/mir/checker"
	"github.com/filecoin-project/mir/pkg/deploytest"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/trantor/types"
	"github.com/filecoin-project/mir/samples/bcb-native/testing2/nodeinstance"
	"github.com/filecoin-project/mir/samples/bcb-native/testing2/properties"
	"github.com/filecoin-project/mir/samples/bcb-native/testing2/puppeteers"
	"github.com/filecoin-project/mir/stdtypes"
)

const MAX_EVENTS = 200
const MAX_HEARTBEATS_INACTIVE = 10

type CATestConfig struct {
	Nodes              []stdtypes.NodeID
	ByzantineNodes     []stdtypes.NodeID
	Sender             stdtypes.NodeID
	CreateNodeInstance ni.NodeInstanceCreationFunc[nodeinstance.BcbNodeInstanceConfig]
	Puppeteer          puppeteer.Puppeteer
	ExpectedResult     map[string]checker.CheckerResult
}

var catests = map[string]CATestConfig{
	"no byz": {
		Nodes:              []stdtypes.NodeID{"0", "1", "2", "3", "4"},
		ByzantineNodes:     []stdtypes.NodeID{"1", "2"},
		Sender:             stdtypes.NodeID("0"),
		CreateNodeInstance: nodeinstance.CreateBcbNodeInstance,
		Puppeteer:          puppeteers.NewOneNodeBroadcast(stdtypes.NodeID("0")),
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

		// Setup System & Adversary
		nodeWeights := make(map[stdtypes.NodeID]types.VoteWeight, len(test.Nodes))
		for i := range test.Nodes {
			id := stdtypes.NewNodeIDFromInt(i)
			nodeWeights[id] = "1"
		}
		instanceUID := []byte("testing instance")
		nodeConfigs := make(map[stdtypes.NodeID]nodeinstance.BcbNodeInstanceConfig, len(test.Nodes))
		fmt.Println(test.Nodes)
		config := nodeinstance.BcbNodeInstanceConfig{InstanceUID: instanceUID, NumberOfNodes: len(test.Nodes), Leader: test.Sender, FakeTransport: deploytest.NewFakeTransport(nodeWeights), ReportPath: reportDir}
		for _, nodeID := range test.Nodes {
			fmt.Println(nodeID)
			nodeConfigs[nodeID] = config
		}
		weightedActions := []actions.WeightedAction{
			actions.NewWeightedAction(func(e stdtypes.Event, sourceNode stdtypes.NodeID, byzantineNodes []stdtypes.NodeID) (string, map[stdtypes.NodeID]*stdtypes.EventList, error) {
				return "",
					map[stdtypes.NodeID]*stdtypes.EventList{
						sourceNode: stdtypes.ListOf(e),
					},
					nil
			}, 10),
			actions.NewWeightedAction(func(e stdtypes.Event, sourceNode stdtypes.NodeID, byzantineNodes []stdtypes.NodeID) (string, map[stdtypes.NodeID]*stdtypes.EventList, error) {
				return fmt.Sprintf("dropped event %s", e.ToString()),
					nil,
					nil
			}, 1),
			actions.NewWeightedAction(func(e stdtypes.Event, sourceNode stdtypes.NodeID, byzantineNodes []stdtypes.NodeID) (string, map[stdtypes.NodeID]*stdtypes.EventList, error) {
				e2, err := e.SetMetadata("duplicated", true)
				if err != nil {
					// TODO: should a failed action just be a "noop"
					return "", nil, err
				}
				return fmt.Sprintf("duplicated event %v", e.ToString()),
					map[stdtypes.NodeID]*stdtypes.EventList{
						sourceNode: stdtypes.ListOf(e, e2),
					},
					nil
			}, 1),
		}

		eventsOfInterest := []stdtypes.Event{&broadcastevents.Deliver{}, &broadcastevents.BroadcastRequest{}}

		// TODO: set maxByzantineNodes correctly
		adv, err := centraladversay.NewAdversary(test.CreateNodeInstance, nodeConfigs, eventsOfInterest, weightedActions, test.ByzantineNodes, logger)
		panicIfErr(err)

		systemConfig := &properties.SystemConfig{
			AllNodes:       test.Nodes,
			Sender:         test.Sender,
			ByzantineNodes: test.ByzantineNodes,
		}

		m := map[stdtypes.ModuleID]modules.Module{
			"validity":    properties.NewValidity(*systemConfig, logger),
			"integrity":   properties.NewIntegrity(*systemConfig, logger),
			"consistency": properties.NewConsistency(*systemConfig, logger),
		}

		// Setup Checker
		c, err := checker.NewChecker(m)
		panicIfErr(err)

		// Run experiment with live checker
		adv.RunExperiment(test.Puppeteer, c, MAX_EVENTS, MAX_HEARTBEATS_INACTIVE)

		results, _ := c.GetResults()

		allPassed := true
		resultStr := fmt.Sprintf("Results: (%s)\n", reportDir)
		for label, res := range results {
			resultStr = fmt.Sprintf("%s%s was %s - expected: %s\n", resultStr, label, res.String(), test.ExpectedResult[label])
			// NOTE: doesn't guarantee that the all expected res are checked... but that is ok for now
			assert.Equal(t, res, test.ExpectedResult[label])
			if res != checker.SUCCESS {
				allPassed = false
			}
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
		resultStr = fmt.Sprintf("%s\n\n%s\n\nCommit: %s\n\n", testName, resultStr, commit)
		resultStr += adv.GetActionLogString()
		panicIfErr(os.WriteFile(path.Join(reportDir, "report.txt"), []byte(resultStr), 0644))

		// delete report dir if all tests passed
		if allPassed {
			os.RemoveAll(reportDir)
		}
	})
}

func panicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}
