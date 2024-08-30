package main

import (
	"context"
	"fmt"
	"math/rand/v2"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	es "github.com/go-errors/errors"

	"github.com/filecoin-project/mir/fuzzer"
	"github.com/filecoin-project/mir/fuzzer/actions"
	"github.com/filecoin-project/mir/fuzzer/checker"

	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/trantor/types"
	"github.com/filecoin-project/mir/samples/reliable-broadcast/testing/nodeinstance"
	"github.com/filecoin-project/mir/samples/reliable-broadcast/testing/properties"
	"github.com/filecoin-project/mir/stdtypes"
)

const (
	evaluationRuns = 100000
)

var (
	reportDir     = fmt.Sprintf("./evaluation_%s", time.Now().Format("2006-01-02_15-04-05"))
	propertyNames = []string{"integrity", "validity", "consistency", "integrity", "totality"}
)

func evaluationRun(
	name string,
	evaluationRunNr int,
	nodes []stdtypes.NodeID,
	byzantineNodes []stdtypes.NodeID,
	sender stdtypes.NodeID,
	byzantineActions []actions.WeightedAction,
	networkActions []actions.WeightedAction,
	rounds int,
	logLevel logging.LogLevel,
	resC chan map[string]checker.CheckerResult,
) (map[string]int, error) {
	// check that shit makes sense
	// rename to lower case

	nodeWeights := make(map[stdtypes.NodeID]types.VoteWeight, len(nodes))
	for i := range nodes {
		id := stdtypes.NewNodeIDFromInt(i)
		nodeWeights[id] = "1"
	}
	instanceUID := []byte("fuzzing instance")
	nodeConfigs := make(map[stdtypes.NodeID]nodeinstance.ReliableBroadcastNodeInstanceConfig, len(nodes))

	// TODO: the rounds should actually be part of fuzzer, and not handeled here...
	config := nodeinstance.ReliableBroadcastNodeInstanceConfig{InstanceUID: instanceUID, NumberOfNodes: len(nodes), Leader: sender}
	for _, nodeID := range nodes {
		nodeConfigs[nodeID] = config
	}

	checkerParams := properties.SystemConfig{
		AllNodes:       nodes,
		Sender:         sender,
		ByzantineNodes: byzantineNodes,
	}

	fuzzer, err := fuzzer.NewFuzzer(
		nodeinstance.CreateBroadcastNodeInstance,
		nodeConfigs,
		byzantineNodes,
		puppeteerEvents,
		isInterestingEvent,
		byzantineActions,
		networkActions,
		properties.CreateReliableBroadcastChecker,
		checkerParams,
		path.Join(reportDir, name),
	)
	if err != nil {
		return nil, es.Errorf("failed to create fuzzer: %v", err)
	}

	// TODO: properly deal with context
	ctx := context.Background()
	hitsAfter, err := fuzzer.RunEvaluation(ctx, name, rounds, MAX_RUN_DURATION, rand.New(rand.NewPCG(SEED1, SEED2+uint64(evaluationRunNr))), logLevel, resC)
	if err != nil {
		return nil, es.Errorf("fuzzer encountered an issue: %v", err)
	}

	return hitsAfter, nil
}

func Test_Evaluation(t *testing.T) {
	logLevel := logging.LevelInfo
	nodes := []stdtypes.NodeID{"0", "1", "2", "3"}
	byzantineNodes := []stdtypes.NodeID{"2", "3"}
	startTime := time.Now()
	resC := make(chan map[string]checker.CheckerResult, 100)

	evaluationPath := path.Join(reportDir, fmt.Sprintf("evaluation-intersection-%d-%d-%d.csv", len(nodes), len(byzantineNodes), evaluationRuns))
	err := os.MkdirAll(reportDir, os.ModePerm)
	if err != nil {
		fmt.Printf("failed to create report directory: %v", err)
		return
	}
	evaluationFile, err := os.Create(evaluationPath)
	if err != nil {
		fmt.Printf("failed to create evaluation file: %v", err)
		return
	}
	// write header
	fmt.Fprintln(evaluationFile, strings.Join(propertyNames, ", "))
	err = evaluationFile.Sync()
	if err != nil {
		fmt.Printf("failed to flushing to evaluation file: %v", err)
		return
	}
	defer evaluationFile.Close()
	go func() {
		_, err = evaluationRun(fmt.Sprintf("evaluation-%d", 0), 0, nodes, byzantineNodes, stdtypes.NodeID("0"), weightedActionsForByzantineNodes, weightedActionsForNetwork, evaluationRuns, logLevel, resC)
		close(resC)
		if err != nil {
			fmt.Println(err)
			return
		}
	}()

	for resMap := range resC {
		runStr := ""
		for _, property := range propertyNames {
			if resMap[property] == checker.FAILURE {
				runStr += "1,"
			} else {
				runStr += "0,"
			}
		}
		fmt.Fprintln(evaluationFile, runStr)
	}
	err = evaluationFile.Sync()
	if err != nil {
		fmt.Printf("failed to flushing to evaluation file: %v", err)
		return
	}
	duration := time.Since(startTime)
	fmt.Printf("=================================================================\nExecution time: %s\n", duration)
}
