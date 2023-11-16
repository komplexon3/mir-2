// handles all commmunication between nodes

package main

import (
	"github.com/filecoin-project/mir/pkg/dsl"
	"github.com/filecoin-project/mir/pkg/modules"
	"github.com/filecoin-project/mir/pkg/pb/blockchainpb"
	bcmpbdsl "github.com/filecoin-project/mir/pkg/pb/blockchainpb/bcmpb/dsl"
	communicationpbdsl "github.com/filecoin-project/mir/pkg/pb/blockchainpb/communicationpb/dsl"
	communicationpbmsgs "github.com/filecoin-project/mir/pkg/pb/blockchainpb/communicationpb/msgs"
	transportpbdsl "github.com/filecoin-project/mir/pkg/pb/transportpb/dsl"
	t "github.com/filecoin-project/mir/pkg/types"
)

func NewCommunication(otherNodes []t.NodeID, mangle bool) modules.PassiveModule {
	m := dsl.NewModule("communication")

	dsl.UponInit(m, func() error {
		return nil
	})

	communicationpbdsl.UponNewBlock(m, func(block *blockchainpb.Block) error {
		// take the block and send it to all other nodes

		if mangle {
			// send via mangles
			for _, node := range otherNodes {
				transportpbdsl.SendMessage(m, "mangler", communicationpbmsgs.NewBlockMessage("communication", block), []t.NodeID{node})
			}
		} else {
			transportpbdsl.SendMessage(m, "transport", communicationpbmsgs.NewBlockMessage("communication", block), otherNodes)
		}

		return nil
	})

	communicationpbdsl.UponNewBlockMessageReceived(m, func(from t.NodeID, block *blockchainpb.Block) error {
		// take the block and add it to the blockchain
		// could add some randomization here - delay/drop (drop implemented in send)

		bcmpbdsl.NewBlock(m, "bcm", block)
		return nil
	})

	return m
}
