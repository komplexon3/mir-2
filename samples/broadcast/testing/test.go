package main

import (
	"fmt"
	"os"
	"slices"

	"github.com/filecoin-project/mir/checker"
	"github.com/filecoin-project/mir/pkg/eventlog"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/pb/eventpb"
	"github.com/filecoin-project/mir/pkg/pb/recordingpb"
	broadcastevents "github.com/filecoin-project/mir/samples/broadcast/events"
	"github.com/filecoin-project/mir/samples/broadcast/testing/properties"
	"github.com/filecoin-project/mir/stdtypes"
	"gopkg.in/alecthomas/kingpin.v2"
)

// Will analyze the logs in the parent directory. Not the best place to put it but the first thing that I could come up with.

type parsedArgs struct {
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
    files = append(files, fmt.Sprintf("../eventlog0_%d.gz", i))
	}

	logger := logging.ConsoleDebugLogger
	history := getEventsFromFile(files...)
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
			fmt.Print("x")
			eventChan <- ele
		}
		close(eventChan)
		fmt.Println()
	}()

	c.RunAnalysis(eventChan)

	results, _ := c.GetResults()

	fmt.Println("Results:")
	for label, res := range results {
		fmt.Printf("%s was %s\n", label, res.String())
	}

}

func failOnErr(err error) {
	if err != nil {
		panic(err)
	}
}

func getEventsFromFile(filenames ...string) []stdtypes.Event {
	recordingLog := make([]*recordingpb.Entry, 0)

	for _, fn := range filenames {
		file, err := os.Open(fn)
		failOnErr(err)
		defer file.Close() // ignore close errors

		reader, err := eventlog.NewReader(file)
		failOnErr(err)

		entries, err := reader.ReadAllEntries()
		failOnErr(err)

		recordingLog = append(recordingLog, entries...)
	}

	// actually, we want to interleave the histories
	// for simplicity we join them and then sort them

	// keeping original order of items
	slices.SortStableFunc(recordingLog, func(a, b *recordingpb.Entry) int {
		return int(a.GetTime() - b.GetTime())
	})

	eventLog := make([]stdtypes.Event, 0, len(recordingLog))
	for _, rle := range recordingLog {
		for _, e := range rle.Events {
			var event stdtypes.Event = e
			if serializedEvent, ok := event.(*eventpb.Event).Type.(*eventpb.Event_Serialized); ok {
				decodedEvent, err := broadcastevents.Deserialize(serializedEvent.Serialized.GetData())
				if err != nil {
					// TODO: why is (almost?) every event a serialized event?
					// fmt.Println("failed to deserialize event, just passing it on...")
				} else {
					event = decodedEvent
				}
			}
			eventLog = append(eventLog, event)
			// event.SetMetadata("node", stdtypes.NodeID(rle.NodeId))
		}
	}

	return eventLog
}

func parseArgs(args []string) *parsedArgs {
	app := kingpin.New("broadcast testing", "testing some properties - enter number of nodes and ids of byzantine nodes")
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
	}
}
