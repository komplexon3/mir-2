package nodeinstance

import (
	"context"

	"github.com/filecoin-project/mir"
	"github.com/filecoin-project/mir/fuzzer/cortexcreeper"
	"github.com/filecoin-project/mir/fuzzer/interceptors"
	"github.com/filecoin-project/mir/fuzzer/nodeinstance"
	"github.com/filecoin-project/mir/pkg/deploytest"
	"github.com/filecoin-project/mir/pkg/eventlog"
	"github.com/filecoin-project/mir/pkg/idledetection"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/modules"
	trantorpbtypes "github.com/filecoin-project/mir/pkg/pb/trantorpb/types"
	"github.com/filecoin-project/mir/samples/authenticated-broadcast/modules/abroadcast"
	"github.com/filecoin-project/mir/stdtypes"
	es "github.com/go-errors/errors"
)

type ABroadcastNodeInstance struct {
	node            *mir.Node
	transportModule *deploytest.FakeLink
	cortexCreeper   *cortexcreeper.CortexCreeper
	idleDetectionC  chan idledetection.IdleNotification
	eventLogger     *eventlog.Recorder
	nodeID          stdtypes.NodeID
	config          ABroadcastNodeInstanceConfig
}

type ABroadcastNodeInstanceConfig struct {
	Leader        stdtypes.NodeID
	InstanceUID   []byte
	NumberOfNodes int
}

func (i *ABroadcastNodeInstance) GetNode() *mir.Node {
	return i.node
}

func (i *ABroadcastNodeInstance) GetIdleDetectionC() chan idledetection.IdleNotification {
	return i.idleDetectionC
}

func (i *ABroadcastNodeInstance) Run(ctx context.Context) error {
	defer i.eventLogger.Stop()
	return i.node.Run(ctx)
}

func (i *ABroadcastNodeInstance) Stop() {
	i.cortexCreeper.AbortIntercepts()
	i.node.Stop()
}

func (i *ABroadcastNodeInstance) Setup() error {
	i.cortexCreeper.Setup(i.node)
	i.transportModule.Connect(&trantorpbtypes.Membership{})
	return nil
}

func (i *ABroadcastNodeInstance) Cleanup() error {
	i.transportModule.Stop()
	return nil
}

func CreateABroadcastNodeInstance(nodeID stdtypes.NodeID, config ABroadcastNodeInstanceConfig, transport *deploytest.FakeTransport, cortexCreeper *cortexcreeper.CortexCreeper, logPath string, logger logging.Logger) (nodeinstance.NodeInstance, error) {
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
	aBroadcastModule := abroadcast.NewModule(
		abroadcast.ModuleConfig{
			Self:     "broadcast",
			Consumer: "null",
			Net:      "net",
			Crypto:   "crypto",
		},
		&abroadcast.ModuleParams{
			AllNodes: nodeIDs,
			Leader:   config.Leader,
		},
		nodeID,
		logger,
	)

	eventLogger, err := eventlog.NewRecorder(nodeID, logPath, logger)
	if err != nil {
		return nil, es.Errorf("error setting up event logger interceptor: %w", err)
	}

	interceptor, err := interceptors.NewFuzzerInterceptor(nodeID, cortexCreeper, logger, eventLogger)
	if err != nil {
		return nil, es.Errorf("error setting up fuzzer interceptor: %w", err)
	}

	m := map[stdtypes.ModuleID]modules.Module{
		"net":       transportModule,
		"broadcast": aBroadcastModule,
		"null":      modules.NullPassive{}, // just sending delivers to null, will still be intercepted
	}

	// create a Mir node
	node, idleDetectionC, err := mir.NewNodeWithIdleDetection(nodeID, mir.DefaultNodeConfig().WithLogger(logger), m, interceptor)
	if err != nil {
		return nil, es.Errorf("error creating a Mir node: %w", err)
	}

	instance := ABroadcastNodeInstance{
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
