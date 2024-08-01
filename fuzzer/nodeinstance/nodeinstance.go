package nodeinstance

import (
	"context"

	"github.com/filecoin-project/mir"
	"github.com/filecoin-project/mir/fuzzer/cortexcreeper"
	"github.com/filecoin-project/mir/pkg/deploytest"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/stdtypes"
)

type NodeInstance interface {
	GetNode() *mir.Node
	GetIdleDetectionC() chan chan struct{}
	Setup() error
	Cleanup() error
	Run(ctx context.Context) error
	Stop()
}

type NodeInstanceCreationFunc[T any] func(nodeID stdtypes.NodeID, config T, transport *deploytest.FakeTransport, cortexCreeper *cortexcreeper.CortexCreeper, logPath string, logger logging.Logger) (NodeInstance, error)

type NodeConfigs[T any] map[stdtypes.NodeID]T

type NodeInstances map[stdtypes.NodeID]NodeInstance
