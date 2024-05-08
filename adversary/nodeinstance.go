package adversary

import (
	"github.com/filecoin-project/mir"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/stdtypes"
)

type NodeInstance interface {
	GetNode() *mir.Node
	Setup() error
	Cleanup() error
}

type NodeInstanceCreationFunc[T interface{}] func(nodeID stdtypes.NodeID, config T, cortexCreeper CortexCreeper, logger logging.Logger) (NodeInstance, error)

type NodeConfigs[T interface{}] map[stdtypes.NodeID]T
