package nodeinstance

import (
	"context"

	"github.com/filecoin-project/mir"
	"github.com/filecoin-project/mir/fuzzer/cortexcreeper"
	"github.com/filecoin-project/mir/fuzzer/nodeinstance"
	"github.com/filecoin-project/mir/pkg/crypto"
	mirCrypto "github.com/filecoin-project/mir/pkg/crypto"
	"github.com/filecoin-project/mir/pkg/deploytest"
	"github.com/filecoin-project/mir/pkg/eventlog"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/modules"
	trantorpbtypes "github.com/filecoin-project/mir/pkg/pb/trantorpb/types"
	"github.com/filecoin-project/mir/pkg/utilinterceptors"
	"github.com/filecoin-project/mir/pkg/vcinterceptor"
	"github.com/filecoin-project/mir/samples/bcb-native/modules/bcb"
	"github.com/filecoin-project/mir/stdtypes"
	es "github.com/go-errors/errors"
)

type BcbNodeInstance struct {
	node            *mir.Node
	nodeID          stdtypes.NodeID
	transportModule *deploytest.FakeLink
	config          BcbNodeInstanceConfig
	cortexCreeper   cortexcreeper.CortexCreeper
}

type BcbNodeInstanceConfig struct {
	InstanceUID   []byte
	NumberOfNodes int
	Leader        stdtypes.NodeID
	FakeTransport *deploytest.FakeTransport
	LogPath       string
}

func (bi *BcbNodeInstance) GetNode() *mir.Node {
	return bi.node
}

func (bi *BcbNodeInstance) Run(ctx context.Context) error {
	go bi.cortexCreeper.RunInjector(ctx) // ignoring error for now
	return bi.node.Run(ctx)
}

func (bi *BcbNodeInstance) Stop() {
	bi.cortexCreeper.StopInjector()
	bi.node.Stop()
}

func (bi *BcbNodeInstance) Setup() error {
	bi.cortexCreeper.Setup(bi.node)
	bi.transportModule.Connect(&trantorpbtypes.Membership{})
	return nil
}

func (bi *BcbNodeInstance) Cleanup() error {
	bi.transportModule.Stop()
	return nil
}

func CreateBcbNodeInstance(nodeID stdtypes.NodeID, config BcbNodeInstanceConfig, cortexCreeper cortexcreeper.CortexCreeper, logger logging.Logger) (nodeinstance.NodeInstance, error) {
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
	bcbModule := bcb.NewModule(
		bcb.ModuleConfig{
			Self:     "bcb",
			Consumer: "null",
			Net:      "net",
			Crypto:   "crypto",
		},
		&bcb.ModuleParams{
			InstanceUID: config.InstanceUID,
			AllNodes:    nodeIDs,
			Leader:      config.Leader,
		},
		nodeID,
		logger,
	)

	eventLogger, err := eventlog.NewRecorder(nodeID, config.LogPath, logger)
	if err != nil {
		return nil, es.Errorf("error setting up interceptor: %w", err)
	}

	interceptor := eventlog.MultiInterceptor(vcinterceptor.New(nodeID), &utilinterceptors.NodeIdMetadataInterceptor{NodeID: nodeID}, eventLogger, cortexCreeper)

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
		"net":    transportModule,
		"crypto": mirCrypto.New(crypto),
		"bcb":    bcbModule,
		"null":   modules.NullPassive{}, // just sending delivers to null, will still be intercepted
	}

	// create a Mir node
	node, err := mir.NewNode(nodeID, mir.DefaultNodeConfig().WithLogger(logger), m, interceptor)
	if err != nil {
		return nil, es.Errorf("error creating a Mir node: %w", err)
	}

	instance := BcbNodeInstance{
		node:            node,
		nodeID:          nodeID,
		transportModule: transportModule,
		config:          config,
		cortexCreeper:   cortexCreeper,
	}

	return &instance, nil
}
