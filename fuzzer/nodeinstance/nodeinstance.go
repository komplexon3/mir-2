package nodeinstance

import (
	"context"

	"github.com/filecoin-project/mir"
	"github.com/filecoin-project/mir/fuzzer/centraladversary/cortexcreeper"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/stdtypes"
)

type NodeInstance interface {
	GetNode() *mir.Node
	Setup() error
	Cleanup() error
	Run(ctx context.Context) error
	Stop()
}

type NodeInstanceCreationFunc[T interface{}] func(nodeID stdtypes.NodeID, config T, cortexCreeper *cortexcreeper.CortexCreeper, logger logging.Logger) (NodeInstance, error)

type NodeConfigs[T interface{}] map[stdtypes.NodeID]T
