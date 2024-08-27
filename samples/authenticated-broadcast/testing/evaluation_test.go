package main

import (
	"context"
	"fmt"
	"math/rand/v2"
	"os"
	"path"
	"testing"
	"time"

	es "github.com/go-errors/errors"

	"github.com/filecoin-project/mir/fuzzer"
	"github.com/filecoin-project/mir/fuzzer/actions"

	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/trantor/types"
	"github.com/filecoin-project/mir/samples/authenticated-broadcast/testing/nodeinstance"
	"github.com/filecoin-project/mir/samples/authenticated-broadcast/testing/properties"
	"github.com/filecoin-project/mir/stdtypes"
)

const (
	runsPerEvaluationRun = 10000
	evaluationRuns       = 1000
)

var (
	reportDir     = fmt.Sprintf("./evaluation_%s", time.Now().Format("2006-01-02_15-04-05"))
	propertyNames = []string{"validity"} //, "consistency", "integrity"}
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
) (map[string]int, error) {
	// check that shit makes sense
	// rename to lower case

	nodeWeights := make(map[stdtypes.NodeID]types.VoteWeight, len(nodes))
	for i := range nodes {
		id := stdtypes.NewNodeIDFromInt(i)
		nodeWeights[id] = "1"
	}
	instanceUID := []byte("fuzzing instance")
	nodeConfigs := make(map[stdtypes.NodeID]nodeinstance.ABroadcastNodeInstanceConfig, len(nodes))

	// TODO: the rounds should actually be part of fuzzer, and not handeled here...
	config := nodeinstance.ABroadcastNodeInstanceConfig{InstanceUID: instanceUID, NumberOfNodes: len(nodes), Leader: sender}
	for _, nodeID := range nodes {
		nodeConfigs[nodeID] = config
	}

	checkerParams := properties.SystemConfig{
		AllNodes:       nodes,
		Sender:         sender,
		ByzantineNodes: byzantineNodes,
	}

	fuzzer, err := fuzzer.NewFuzzer(
		nodeinstance.CreateABroadcastNodeInstance,
		nodeConfigs,
		byzantineNodes,
		puppeteerEvents,
		isInterestingEvent,
		byzantineActions,
		networkActions,
		properties.CreateABroadcastChecker,
		checkerParams,
		path.Join(reportDir, name),
	)
	if err != nil {
		return nil, es.Errorf("failed to create fuzzer: %v", err)
	}

	// TODO: properly deal with context
	ctx := context.Background()
	hitsAfter, err := fuzzer.RunEvaluation(ctx, name, rounds, MAX_RUN_DURATION, rand.New(rand.NewPCG(SEED1, SEED2+uint64(evaluationRunNr))), len(propertyNames), logLevel)
	if err != nil {
		return nil, es.Errorf("fuzzer encountered an issue: %v", err)
	}

	return hitsAfter, nil
}

func Test_Evaluation(t *testing.T) {
	logLevel := logging.LevelInfo
	nodes := []stdtypes.NodeID{"0", "1", "2", "3"}
	byzantineNodes := []stdtypes.NodeID{"1", "2"}
	startTime := time.Now()

	evaluationPath := path.Join(reportDir, fmt.Sprintf("evaluation-%d-%d-%d-%d.csv", len(nodes), len(byzantineNodes), evaluationRuns, runsPerEvaluationRun))
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
	fmt.Fprint(evaluationFile, "validity, consistency, integrity\n")
	err = evaluationFile.Sync()
	if err != nil {
		fmt.Printf("failed to flushing to evaluation file: %v", err)
		return
	}
	defer evaluationFile.Close()
	for r := range evaluationRuns {
		hits, err := evaluationRun(fmt.Sprintf("evaluation-%d", r), r, nodes, byzantineNodes, stdtypes.NodeID("0"), weightedActionsForByzantineNodes, weightedActionsForNetwork, runsPerEvaluationRun, logLevel)
		if err != nil {
			fmt.Println(err)
			return
		}

		runStr := ""
		for _, property := range propertyNames {
			res, ok := hits[property]
			if ok {
				runStr += fmt.Sprintf("%d,", res)
			} else {
				runStr += ","
			}
		}

		fmt.Fprintln(evaluationFile, runStr)
		err = evaluationFile.Sync()
		if err != nil {
			fmt.Printf("failed to flushing to evaluation file: %v", err)
			return
		}
	}
	duration := time.Since(startTime)
	fmt.Printf("=================================================================\nExecution time: %s - for a total of %d rounds (avg per round: %s)\n", duration, runsPerEvaluationRun*evaluationRuns, time.Duration(int64(duration)/int64(runsPerEvaluationRun*evaluationRuns)))
}

// func Test_Ressources(t *testing.T) {
// 	defer goleak.VerifyNone(t)
// 	rounds := 10
// 	logLevel := logging.LevelDebug
// 	hits, err := fuzz("test", []stdtypes.NodeID{"0", "1", "2", "3", "4"}, []stdtypes.NodeID{"1", "2"}, stdtypes.NodeID("0"), weightedActionsForByzantineNodes, weightedActionsForNetwork, rounds, logLevel)
// 	if err != nil {
// 		fmt.Println(err)
// 	}
// 	fmt.Printf("=================================================================\nInteresting cases %d (out of %d - %.3f%%)\n", hits, rounds, float32(hits)/float32(rounds)*100)
// }
