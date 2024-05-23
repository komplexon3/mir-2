package main

import (
	"fmt"
	"os"
	"path"

	"github.com/filecoin-project/mir/checker"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/modules"
	broadcastevents "github.com/filecoin-project/mir/samples/broadcast/events"
	testmodules "github.com/filecoin-project/mir/samples/broadcast/properties"
	"github.com/filecoin-project/mir/stdevents"
	"github.com/filecoin-project/mir/stdtypes"
	"gopkg.in/alecthomas/kingpin.v2"
)

// Will analyze the logs in the parent directory. Not the best place to put it but the first thing that I could come up with.

type parsedArgs struct {
	TraceDir       string
	NumberOfNodes  int
	ByzantineNodes []stdtypes.NodeID
}

func main() {
	args := parseArgs(os.Args)

	byzantineNodes := args.ByzantineNodes
	nodes := make([]stdtypes.NodeID, 0, args.NumberOfNodes)
	files := make([]string, 0, args.NumberOfNodes)
	for i := 0; i < args.NumberOfNodes; i++ {
		nodes = append(nodes, stdtypes.NewNodeIDFromInt(i))
		files = append(files, path.Join(args.TraceDir, fmt.Sprintf("eventlog0_%d.gz", i)))
	}

	logger := logging.ConsoleDebugLogger
	history := checker.GetEventsFromFileSortedByVectorClock([]func([]byte) (stdtypes.Event, error){broadcastevents.Deserialize, stdevents.Deserialize}, files...)
	eventChan := make(chan stdtypes.Event)

	systemConfig := &testmodules.SystemConfig{
		AllNodes:       nodes,
		ByzantineNodes: byzantineNodes,
	}

	m := map[stdtypes.ModuleID]modules.Module{
		"validity":    testmodules.NewValidity(*systemConfig, logger),
		"integrity":   testmodules.NewIntegrity(*systemConfig, logger),
		"consistency": testmodules.NewConsistency(*systemConfig, logger),
	}

	c, err := checker.NewChecker(m)
	if err != nil {
		panic(err)
	}

	fmt.Println("Starting analysis")
	go func() {
		for _, ele := range history {
			eventChan <- ele
		}
		close(eventChan)
	}()

	c.RunAnalysis(eventChan)

	results, _ := c.GetResults()

	fmt.Println("Results:")
	for label, res := range results {
		fmt.Printf("%s was %s\n", label, res.String())
	}

}

func parseArgs(args []string) *parsedArgs {
	app := kingpin.New("broadcast testing", "testing some properties - enter number of nodes and ids of byzantine nodes")
	traceDir := app.Arg("traceDir", "directory in which the trace files are").Required().String()
	numberOfNodes := app.Arg("count", "number of nodes").Required().Int()
	byzantineNodes := app.Arg("bNodes", "ids of byzantine nodes").Strings()

	if _, err := app.Parse(args[1:]); err != nil { // Skip args[0], which is the name of the program, not an argument.
		app.FatalUsage("could not parse arguments: %v\n", err)
	}

	byzantineNodeIDs := make([]stdtypes.NodeID, 0, len(*byzantineNodes))
	for _, b := range *byzantineNodes {
		byzantineNodeIDs = append(byzantineNodeIDs, stdtypes.NodeID(b))
	}

	return &parsedArgs{
		NumberOfNodes:  *numberOfNodes,
		ByzantineNodes: byzantineNodeIDs,
		TraceDir:       *traceDir,
	}
}
