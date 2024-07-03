package localnet

import (
	"fmt"
	"sync"

	"github.com/filecoin-project/mir/pkg/logging"
	"github.com/filecoin-project/mir/pkg/pb/messagepb"
	messagepbtypes "github.com/filecoin-project/mir/pkg/pb/messagepb/types"
	transportpbevents "github.com/filecoin-project/mir/pkg/pb/transportpb/events"
	trantorpbtypes "github.com/filecoin-project/mir/pkg/pb/trantorpb/types"
	"github.com/filecoin-project/mir/pkg/trantor/types"
	"github.com/filecoin-project/mir/stdevents"
	"github.com/filecoin-project/mir/stdtypes"
)

type LocalNetwork struct {
	// Buffers is source x dest
	Buffers          map[stdtypes.NodeID]map[stdtypes.NodeID]chan *stdtypes.EventList
	NodeSinks        map[stdtypes.NodeID]chan *stdtypes.EventList
	Connectivity     map[stdtypes.NodeID]map[stdtypes.NodeID]bool // whether or not there is connectivity from node A to node B
	connectivityLock *sync.RWMutex
	logger           logging.Logger
	nodeIDsWeight    map[stdtypes.NodeID]types.VoteWeight
}

func NewLocalNetwork(nodeIDsWeight map[stdtypes.NodeID]types.VoteWeight) *LocalNetwork {
	buffers := make(map[stdtypes.NodeID]map[stdtypes.NodeID]chan *stdtypes.EventList)
	nodeSinks := make(map[stdtypes.NodeID]chan *stdtypes.EventList)
	for sourceID := range nodeIDsWeight {
		buffers[sourceID] = make(map[stdtypes.NodeID]chan *stdtypes.EventList)
		for destID := range nodeIDsWeight {
			if sourceID == destID {
				continue
			}
			buffers[sourceID][destID] = make(chan *stdtypes.EventList, 10000)
		}
		nodeSinks[sourceID] = make(chan *stdtypes.EventList)
	}

	return &LocalNetwork{
		Buffers:       buffers,
		NodeSinks:     nodeSinks,
		logger:        logging.ConsoleErrorLogger,
		nodeIDsWeight: nodeIDsWeight,
	}
}

func (ft *LocalNetwork) SendRawMessage(sourceNode, destNode stdtypes.NodeID, destModule stdtypes.ModuleID, message stdtypes.Message) error {

	ft.connectivityLock.RLock()
	if !ft.Connectivity[sourceNode][destNode] {
		ft.logger.Log(logging.LevelDebug, "Message dropped because connectivity from source to destination is disabled", "source", sourceNode, "dest", destNode)
		return nil
	}
	ft.connectivityLock.RUnlock()

	select {
	case ft.Buffers[sourceNode][destNode] <- stdtypes.ListOf(
		stdevents.NewMessageReceived(destModule, sourceNode, message),
	):
	default:
		fmt.Printf("Warning: Dropping message %v from %s to %s\n", message, sourceNode, destNode)
	}

	return nil
}

func (ft *LocalNetwork) Send(sourceNode, destNode stdtypes.NodeID, msg *messagepb.Message) {
	ft.connectivityLock.RLock()
	if !ft.Connectivity[sourceNode][destNode] {
		ft.logger.Log(logging.LevelDebug, "Message dropped because connectivity from source to destination is disabled", "source", sourceNode, "dest", destNode)
		return
	}
	ft.connectivityLock.RUnlock()

	select {
	case ft.Buffers[sourceNode][destNode] <- stdtypes.ListOf(
		transportpbevents.MessageReceived(stdtypes.ModuleID(msg.DestModule), sourceNode, messagepbtypes.MessageFromPb(msg)).Pb(),
	):
	default:
		fmt.Printf("Warning: Dropping message %T from %s to %s\n", msg.Type, sourceNode, destNode)
	}
}

// func (ft *LocalNetwork) Link(source stdtypes.NodeID) (net.Transport, error) {
// 	return &FakeLink{
// 		Source:        source,
// 		FakeTransport: ft,
// 		DoneC:         make(chan struct{}),
// 	}, nil
// }

func (ft *LocalNetwork) Membership() *trantorpbtypes.Membership {
	membership := &trantorpbtypes.Membership{Nodes: make(map[stdtypes.NodeID]*trantorpbtypes.NodeIdentity)}
	// Dummy addresses. Never actually used.
	for nID := range ft.Buffers {
		membership.Nodes[nID] = &trantorpbtypes.NodeIdentity{
			Id:     nID,
			Addr:   "",
			Key:    nil,
			Weight: ft.nodeIDsWeight[nID],
		}
	}

	return membership
}

func (ft *LocalNetwork) Close() {}

func (ft *LocalNetwork) RecvChannel(dest stdtypes.NodeID) <-chan *stdtypes.EventList {
	return ft.NodeSinks[dest]
}
