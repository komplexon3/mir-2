package main

import (
	"fmt"
	"testing"
	"time"

	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/stdtypes"
)

func Test_0(t *testing.T) {
	// defer goleak.VerifyNone(t)
	rounds := 100
	logger := logging.ConsoleWarnLogger
	logger = logging.Synchronize(logger)
	startTime := time.Now()
	fuzzBCB("test", []stdtypes.NodeID{"0", "1", "2", "3"}, []stdtypes.NodeID{"1"}, stdtypes.NodeID("0"), weightedActionsForByzantineNodes, weightedActionsForNetwork, rounds, logger)
	duration := time.Since(startTime)
	fmt.Printf("=================================================================\nExecution time: %s - for a total of %d rounds (avg per round: %s)\n", duration, rounds, time.Duration(int64(duration)/int64(rounds)))
}
