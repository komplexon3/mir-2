package testmodules

import "github.com/filecoin-project/mir/stdtypes"

type SystemConfig struct {
	ByzantineNodes []stdtypes.NodeID
	AllNodes       []stdtypes.NodeID
}

type broadcastRequest struct {
	data []byte
	node stdtypes.NodeID
}

func getNodeIdFromMetadata(e stdtypes.Event) stdtypes.NodeID {
	nodeId, err := e.GetMetadata("node")
	if err != nil {
		panic("handleDeliver - node not in metadata")
	}

  // TODO: just converting without checking - will fail nastily if not string
  return stdtypes.NodeID(nodeId.(string))
}
