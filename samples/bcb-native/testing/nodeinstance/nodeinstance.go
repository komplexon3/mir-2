package nodeinstance

import (
	"context"

	"github.com/filecoin-project/mir"
	"github.com/filecoin-project/mir/fuzzer/cortexcreeper"
	msgmetadata "github.com/filecoin-project/mir/fuzzer/interceptors/msgMetadata"
	"github.com/filecoin-project/mir/fuzzer/interceptors/nomulticast"
	"github.com/filecoin-project/mir/fuzzer/interceptors/vcinterceptor"
	"github.com/filecoin-project/mir/fuzzer/nodeinstance"
	mirCrypto "github.com/filecoin-project/mir/pkg/crypto"
	"github.com/filecoin-project/mir/pkg/deploytest"
	"github.com/filecoin-project/mir/pkg/eventlog"
	"github.com/filecoin-project/mir/pkg/idledetection"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/modules"
	trantorpbtypes "github.com/filecoin-project/mir/pkg/pb/trantorpb/types"
	"github.com/filecoin-project/mir/samples/bcb-native/modules/bcb"
	"github.com/filecoin-project/mir/stdtypes"
	es "github.com/go-errors/errors"
)

type BcbNodeInstance struct {
	node            *mir.Node
	transportModule *deploytest.FakeLink
	cortexCreeper   *cortexcreeper.CortexCreeper
	idleDetectionC  chan idledetection.IdleNotification
	eventLogger     *eventlog.Recorder
	nodeID          stdtypes.NodeID
	config          BcbNodeInstanceConfig
}

type BcbNodeInstanceConfig struct {
	Leader        stdtypes.NodeID
	InstanceUID   []byte
	NumberOfNodes int
}

func (bi *BcbNodeInstance) GetNode() *mir.Node {
	return bi.node
}

func (bi *BcbNodeInstance) GetIdleDetectionC() chan idledetection.IdleNotification {
	return bi.idleDetectionC
}

func (bi *BcbNodeInstance) Run(ctx context.Context) error {
	defer bi.eventLogger.Stop()
	go bi.cortexCreeper.Run(ctx) // ignoring error for now
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

func CreateBcbNodeInstance(nodeID stdtypes.NodeID, config BcbNodeInstanceConfig, transport *deploytest.FakeTransport, cortexCreeper *cortexcreeper.CortexCreeper, logPath string, logger logging.Logger) (nodeinstance.NodeInstance, error) {
	nodeIDs := make([]stdtypes.NodeID, config.NumberOfNodes)
	for i := 0; i < config.NumberOfNodes; i++ {
		nodeIDs[i] = stdtypes.NewNodeIDFromInt(i)
	}

	transportModule := &deploytest.FakeLink{
		FakeTransport: transport,
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

	eventLogger, err := eventlog.NewRecorder(nodeID, logPath, logger)
	if err != nil {
		return nil, es.Errorf("error setting up interceptor: %w", err)
	}

	msgMetadataInterceptorIn, msgMetadataInterceptorOut := msgmetadata.NewMsgMetadataInterceptorPair(logger, "vc", "msgID")

	interceptor := eventlog.MultiInterceptor(
		msgMetadataInterceptorIn,
		&nomulticast.NoMulticast{},
		cortexCreeper,
		vcinterceptor.New(nodeID),
		msgMetadataInterceptorOut,
		eventLogger,
	)

	// setup crypto
	keyPairs, err := mirCrypto.GenerateKeys(config.NumberOfNodes, 42)
	if err != nil {
		return nil, es.Errorf("error setting up key paris: %w", err)
	}
	crypto, err := mirCrypto.InsecureCryptoForTestingOnly(nodeIDs, nodeID, &keyPairs)
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
	node, idleDetectionC, err := mir.NewNodeWithIdleDetection(nodeID, mir.DefaultNodeConfig().WithLogger(logger), m, interceptor)
	if err != nil {
		return nil, es.Errorf("error creating a Mir node: %w", err)
	}

	instance := BcbNodeInstance{
		node:            node,
		nodeID:          nodeID,
		transportModule: transportModule,
		config:          config,
		cortexCreeper:   cortexCreeper,
		idleDetectionC:  idleDetectionC,
		eventLogger:     eventLogger,
	}

	return &instance, nil
}
