package properties

import (
	"fmt"

	"github.com/filecoin-project/mir/fuzzer/checker"
	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/stdtypes"
)

func CreateBCBChecker(sc SystemConfig, logger logging.Logger) (*checker.Checker, error) {
	checkerProperties := checker.Properties{
		"validity":    NewValidity(sc, logger),
		"integrity":   NewIntegrity(sc, logger),
		"consistency": NewConsistency(sc, logger),
	}

	return checker.NewChecker(checkerProperties)
}

type SystemConfig struct {
	Sender         stdtypes.NodeID
	ByzantineNodes []stdtypes.NodeID
	AllNodes       []stdtypes.NodeID
}

func getNodeIdFromMetadata(e stdtypes.Event) stdtypes.NodeID {
	node, err := e.GetMetadata("node")
	if err != nil {
		panic("node not in metadata")
	}

	// TODO: just converting without checking - will fail nastily if not string
	switch nodeT := node.(type) {
	case int:
		return stdtypes.NewNodeIDFromInt(nodeT)
	case string:
		return stdtypes.NodeID(nodeT)
	case stdtypes.NodeID:
		return nodeT
	default:
		panic(fmt.Errorf("cannot convert %T into stdtypes.NodeID", nodeT))
	}
}
