package nodeinstance

import (
	"time"

	"github.com/filecoin-project/mir"
	"github.com/filecoin-project/mir/adversary"
	"github.com/filecoin-project/mir/pkg/crypto"
	mirCrypto "github.com/filecoin-project/mir/pkg/crypto"
	"github.com/filecoin-project/mir/pkg/deploytest"
	"github.com/filecoin-project/mir/pkg/eventlog"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/modules"
	trantorpbtypes "github.com/filecoin-project/mir/pkg/pb/trantorpb/types"
	"github.com/filecoin-project/mir/pkg/utilinterceptors"
	broadcast "github.com/filecoin-project/mir/samples/broadcast/module"
	"github.com/filecoin-project/mir/stdtypes"
	es "github.com/go-errors/errors"
)

type BroadcastNodeInstance struct {
	node            *mir.Node
	nodeID          stdtypes.NodeID
	transportModule *deploytest.FakeLink
	config          BroadcastNodeInstanceConfig
}

type BroadcastNodeInstanceConfig struct {
	NumberOfNodes int
	FakeTransport *deploytest.FakeTransport
	LogPath       string
}

func (bi *BroadcastNodeInstance) GetNode() *mir.Node {
	return bi.node
}

func (bi *BroadcastNodeInstance) Setup() error {
	bi.transportModule.Connect(&trantorpbtypes.Membership{})
	return nil
}

func (bi *BroadcastNodeInstance) Cleanup() error {
	bi.transportModule.Stop()
	return nil
}

func CreateBroadcastNodeInstance(nodeID stdtypes.NodeID, config BroadcastNodeInstanceConfig, cortexCreeper adversary.CortexCreeper, logger logging.Logger) (adversary.NodeInstance, error) {
	nodeIDs := make([]stdtypes.NodeID, config.NumberOfNodes)
	for i := 0; i < config.NumberOfNodes; i++ {
		nodeIDs[i] = stdtypes.NewNodeIDFromInt(i)
	}

	transportModule := &deploytest.FakeLink{
		FakeTransport: config.FakeTransport,
		Source:        nodeID,
		DoneC:         make(chan struct{}),
	}

	if err := transportModule.Start(); err != nil {
		return nil, es.Errorf("could not start network transport: %w", err)
	}
	broadcastModule := broadcast.NewBroadcast(
		broadcast.ModuleParams{
			NodeID:   nodeID,
			AllNodes: nodeIDs,
		},
		broadcast.ModuleConfig{
			Self:     "broadcast",
			Consumer: "null",
			Net:      "net",
			Crypto:   "crypto",
		},
		logger,
	)

	eventLogger, err := eventlog.NewRecorder(nodeID, config.LogPath, logger, eventlog.TimeSourceOpt(func() int64 { return time.Now().UnixMilli() }))
	if err != nil {
		return nil, es.Errorf("error setting up interceptor: %w", err)
	}

	interceptor := eventlog.MultiInterceptor(&utilinterceptors.NodeIdMetadataInterceptor{NodeID: nodeID}, cortexCreeper, eventLogger)

	// setup crypto
	keyPairs, err := crypto.GenerateKeys(config.NumberOfNodes, 42)
	if err != nil {
		return nil, es.Errorf("error setting up key paris: %w", err)
	}
	crypto, err := crypto.InsecureCryptoForTestingOnly(nodeIDs, nodeID, &keyPairs)
	if err != nil {
		return nil, es.Errorf("error setting up crypto: %w", err)
	}

	m := map[stdtypes.ModuleID]modules.Module{
		"net":       transportModule,
		"crypto":    mirCrypto.New(crypto),
		"broadcast": broadcastModule,
		"null":      modules.NullPassive{}, // just sending delivers to null, will still be intercepted
	}

	// create a Mir node
	node, err := mir.NewNode(nodeID, mir.DefaultNodeConfig().WithLogger(logger), m, interceptor)
	if err != nil {
		return nil, es.Errorf("error creating a Mir node: %w", err)
	}

	instance := BroadcastNodeInstance{
		node:            node,
		nodeID:          nodeID,
		transportModule: transportModule,
		config:          config,
	}

	return &instance, nil
}
