package main

import (
	"context"
	"fmt"
	"os"
	"time"

	es "github.com/go-errors/errors"
	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/filecoin-project/mir/samples/broadcast/controlmodule"
	broadcast "github.com/filecoin-project/mir/samples/broadcast/module"
	"github.com/filecoin-project/mir/samples/broadcast/customattacks/sharedkeysattack/sharedkeysmodule"
	"github.com/filecoin-project/mir/stdtypes"

	"github.com/filecoin-project/mir"
	"github.com/filecoin-project/mir/pkg/crypto"
	mirCrypto "github.com/filecoin-project/mir/pkg/crypto"
	"github.com/filecoin-project/mir/pkg/dsl"
	"github.com/filecoin-project/mir/pkg/eventlog"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/net/grpc"
	trantorpbtypes "github.com/filecoin-project/mir/pkg/pb/trantorpb/types"
	grpctools "github.com/filecoin-project/mir/pkg/util/grpc"
	"github.com/filecoin-project/mir/pkg/util/sliceutil"
	"github.com/filecoin-project/mir/pkg/utilinterceptors"
	"github.com/filecoin-project/mir/pkg/vcinterceptor"
)

const (

	// Base port number for the nodes to listen to messages from each other.
	// The nodes will listen on ports starting from nodeBasePort through nodeBasePort+3.
	nodeBasePort = 10000
)

// parsedArgs represents parsed command-line parameters passed to the program.
type parsedArgs struct {

	// ID of this node.
	// The package github.com/hyperledger-labs/mir/pkg/types defines this and other types used by the library.
	OwnID stdtypes.NodeID

	// number of nodes
	NodeNumber int

	// numner of byzantine nodes
	ByzantineNodesNumber int

	// If set, print debug output to stdout.
	Verbose bool

	// If set, print trace output to stdout.
	Trace bool
}

func main() {
	if err := run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func run() error {
	args := parseArgs(os.Args)

	// Initialize logger that will be used throughout the code to print log messages.
	var logger logging.Logger
	if args.Trace {
		logger = logging.ConsoleTraceLogger // Print trace-level info.
	} else if args.Verbose {
		logger = logging.ConsoleDebugLogger // Print debug-level info in verbose mode.
	} else {
		logger = logging.ConsoleWarnLogger // Only print errors and warnings by default.
	}

	// IDs of nodes that are part of the system.
	// This example uses a static configuration of nodeNumber nodes.
	nodeIDs := make([]stdtypes.NodeID, args.NodeNumber)
	byzantineNodeIDs := make([]stdtypes.NodeID, args.ByzantineNodesNumber)
	for i := 0; i < args.NodeNumber; i++ {
		nodeIDs[i] = stdtypes.NewNodeIDFromInt(i)
		if i < args.ByzantineNodesNumber {
			byzantineNodeIDs[i] = nodeIDs[i]
		}
	}

	// Construct membership, remembering own address.
	membership := &trantorpbtypes.Membership{make(map[stdtypes.NodeID]*trantorpbtypes.NodeIdentity)} // nolint:govet
	var ownAddr stdtypes.NodeAddress
	for i := range nodeIDs {
		id := stdtypes.NewNodeIDFromInt(i)
		addr := grpctools.NewDummyMultiaddr(i + nodeBasePort)

		if id == args.OwnID {
			ownAddr = addr
		}

		membership.Nodes[id] = &trantorpbtypes.NodeIdentity{
			Id:     id,
			Addr:   addr.String(),
			Key:    nil,
			Weight: "1",
		}
	}

	transportModule, err := grpc.NewTransport(grpc.DefaultParams(), args.OwnID, ownAddr.String(), logger, nil)
	if err != nil {
		return es.Errorf("failed to get network transport %w", err)
	}
	if err := transportModule.Start(); err != nil {
		return es.Errorf("could not start network transport: %w", err)
	}
	transportModule.Connect(membership)

	eventLogger, err := eventlog.NewRecorder(args.OwnID, ".", logger, eventlog.TimeSourceOpt(func() int64 { return time.Now().UnixMilli() }))
	if err != nil {
		return es.Errorf("error setting up interceptor: %w", err)
	}

	interceptor := eventlog.MultiInterceptor(&utilinterceptors.NodeIdMetadataInterceptor{NodeID: args.OwnID}, vcinterceptor.New(args.OwnID), &utilinterceptors.EventPrinter{NodeID: args.OwnID}, eventLogger)

	// control module reads the user input from the console and processes it.
	control := controlmodule.NewControlModule(controlmodule.ControlModuleConfig{Self: "control", Broadcast: "broadcast"})

	// setup crypto
	keyPairs, err := crypto.GenerateKeys(args.NodeNumber, 42)
	if err != nil {
		return es.Errorf("error setting up key paris: %w", err)
	}
	m := map[stdtypes.ModuleID]modules.Module{
		"net": transportModule,
		// "crypto":    mirCrypto.New(crypto),
		// "broadcast": broadcastModule,
		"control": control,
	}

	// only used by byzantine nodes
	byzCryptoModules := make(map[stdtypes.NodeID]stdtypes.ModuleID)
	for i := 0; i < args.ByzantineNodesNumber; i++ {
		name := stdtypes.ModuleID(fmt.Sprint("crypto", i))
		byzCryptoModules[byzantineNodeIDs[i]] = name
		crypto, err := crypto.InsecureCryptoForTestingOnly(nodeIDs, byzantineNodeIDs[i], &keyPairs)
		if err != nil {
			return es.Errorf("error setting up crypto: %w", err)
		}
		m[name] = mirCrypto.New(crypto)

	}

	var broadcastModule dsl.Module
	if sliceutil.Contains(byzantineNodeIDs, args.OwnID) {
		broadcastModule = sharedkeysmodule.NewBroadcast(
			sharedkeysmodule.ModuleParams{
				NodeID:         args.OwnID,
				AllNodes:       nodeIDs,
				ByzantineNodes: byzantineNodeIDs,
			},
			sharedkeysmodule.ModuleConfig{
				Self:     "broadcast",
				Consumer: "control",
				Net:      "net",
				Crypto:   byzCryptoModules,
			},
			logger,
		)

	} else {
		broadcastModule = broadcast.NewBroadcast(
			broadcast.ModuleParams{
				NodeID:   args.OwnID,
				AllNodes: nodeIDs,
			},
			broadcast.ModuleConfig{
				Self:     "broadcast",
				Consumer: "control",
				Net:      "net",
				Crypto:   "crypto",
			},
			logger,
		)
		crypto, err := crypto.InsecureCryptoForTestingOnly(nodeIDs, args.OwnID, &keyPairs)
		if err != nil {
			return es.Errorf("error setting up crypto: %w", err)
		}
    m["crypto"] = mirCrypto.New(crypto)
	}

	m["broadcast"] = broadcastModule
	// create a Mir node
	node, err := mir.NewNode("client", mir.DefaultNodeConfig().WithLogger(logger), m, interceptor)
	if err != nil {
		return es.Errorf("error creating a Mir node: %w", err)
	}

	// run the node
	err = node.Run(context.Background())
	if err != nil {
		return es.Errorf("error running node: %w", err)
	}

	return nil
}

// Parses the command-line arguments and returns them in a params struct.
func parseArgs(args []string) *parsedArgs {
	app := kingpin.New("broadcast", "Small broadcast application")
	verbose := app.Flag("verbose", "Verbose mode.").Short('v').Bool()
	trace := app.Flag("trace", "Very verbose mode.").Bool()
	ownID := app.Arg("id", "ID of this node").Required().String()
	numberOfNodes := app.Arg("nodeCount", "number of nodes").Required().Int()
	numberOfByzantineNodes := app.Arg("byzNodeCount", "number of byzantine nodes").Required().Int()

	if _, err := app.Parse(args[1:]); err != nil { // Skip args[0], which is the name of the program, not an argument.
		app.FatalUsage("could not parse arguments: %v\n", err)
	}

	return &parsedArgs{
		OwnID:                stdtypes.NodeID(*ownID),
		NodeNumber:           *numberOfNodes,
		ByzantineNodesNumber: *numberOfByzantineNodes,
		Verbose:              *verbose,
		Trace:                *trace,
	}
}
