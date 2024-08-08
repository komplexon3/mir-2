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
	"github.com/filecoin-project/mir/samples/reliable-broadcast/modules/broadcast"
	"github.com/filecoin-project/mir/stdtypes"
	es "github.com/go-errors/errors"
)

type ReliableBroadcastNodeInstance struct {
	node            *mir.Node
	transportModule *deploytest.FakeLink
	cortexCreeper   *cortexcreeper.CortexCreeper
	idleDetectionC  chan idledetection.IdleNotification
	eventLogger     *eventlog.Recorder
	nodeID          stdtypes.NodeID
	config          ReliableBroadcastNodeInstanceConfig
}

type ReliableBroadcastNodeInstanceConfig struct {
	Leader        stdtypes.NodeID
	InstanceUID   []byte
	NumberOfNodes int
}

func (bi *ReliableBroadcastNodeInstance) GetNode() *mir.Node {
	return bi.node
}

func (bi *ReliableBroadcastNodeInstance) GetIdleDetectionC() chan idledetection.IdleNotification {
	return bi.idleDetectionC
}

func (bi *ReliableBroadcastNodeInstance) Run(ctx context.Context) error {
	defer bi.eventLogger.Stop()
	return bi.node.Run(ctx)
}

func (bi *ReliableBroadcastNodeInstance) Stop() {
	bi.cortexCreeper.AbortIntercepts()
	bi.node.Stop()
}

func (bi *ReliableBroadcastNodeInstance) Setup() error {
	bi.cortexCreeper.Setup(bi.node)
	bi.transportModule.Connect(&trantorpbtypes.Membership{})
	return nil
}

func (bi *ReliableBroadcastNodeInstance) Cleanup() error {
	bi.transportModule.Stop()
	return nil
}

func CreateBroadcastNodeInstance(nodeID stdtypes.NodeID, config ReliableBroadcastNodeInstanceConfig, transport *deploytest.FakeTransport, cortexCreeper *cortexcreeper.CortexCreeper, logPath string, logger logging.Logger) (nodeinstance.NodeInstance, error) {
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
	broadcastModule := broadcast.NewModule(
		broadcast.ModuleConfig{
			Self:     "broadcast",
			Consumer: "null",
			Net:      "net",
		},
		&broadcast.ModuleParams{
			InstanceUID: config.InstanceUID,
			AllNodes:    nodeIDs,
			Leader:      config.Leader,
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
		"broadcast": broadcastModule,
		"null":      modules.NullPassive{}, // just sending delivers to null, will still be intercepted
	}

	// create a Mir node
	node, idleDetectionC, err := mir.NewNodeWithIdleDetection(nodeID, mir.DefaultNodeConfig().WithLogger(logger), m, interceptor)
	if err != nil {
		return nil, es.Errorf("error creating a Mir node: %w", err)
	}

	instance := ReliableBroadcastNodeInstance{
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
