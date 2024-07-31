package nodeinstance

import (
	"context"

	"github.com/filecoin-project/mir"
	"github.com/filecoin-project/mir/fuzzer/cortexcreeper"
	idledetection "github.com/filecoin-project/mir/pkg/idleDetection"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/stdtypes"
)

type NodeInstance interface {
	GetNode() *mir.Node
	GetIdleDetectionC() chan idledetection.IdleNotification
	Setup() error
	Cleanup() error
	Run(ctx context.Context) error
	Stop()
}

type NodeInstanceCreationFunc[T any] func(nodeID stdtypes.NodeID, config T, cortexCreeper *cortexcreeper.CortexCreeper, logger logging.Logger) (NodeInstance, error)

type NodeConfigs[T any] map[stdtypes.NodeID]T

type NodeInstances map[stdtypes.NodeID]NodeInstance
