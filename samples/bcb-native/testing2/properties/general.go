package properties

import (
	"fmt"

	"github.com/filecoin-project/mir/stdtypes"
)

type SystemConfig struct {
	ByzantineNodes []stdtypes.NodeID
	AllNodes       []stdtypes.NodeID
	Sender         stdtypes.NodeID
}

func getNodeIdFromMetadata(e stdtypes.Event) stdtypes.NodeID {
	node, err := e.GetMetadata("node")
	if err != nil {
		panic("handleDeliver - node not in metadata")
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
