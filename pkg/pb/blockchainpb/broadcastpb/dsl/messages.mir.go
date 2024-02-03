// Code generated by Mir codegen. DO NOT EDIT.

package broadcastpbdsl

import (
	dsl "github.com/filecoin-project/mir/pkg/dsl"
	types "github.com/filecoin-project/mir/pkg/pb/blockchainpb/broadcastpb/types"
	types3 "github.com/filecoin-project/mir/pkg/pb/blockchainpb/types"
	dsl1 "github.com/filecoin-project/mir/pkg/pb/messagepb/dsl"
	types2 "github.com/filecoin-project/mir/pkg/pb/messagepb/types"
	types1 "github.com/filecoin-project/mir/pkg/types"
)

// Module-specific dsl functions for processing net messages.

func UponMessageReceived[W types.Message_TypeWrapper[M], M any](m dsl.Module, handler func(from types1.NodeID, msg *M) error) {
	dsl1.UponMessageReceived[*types2.Message_Broadcast](m, func(from types1.NodeID, msg *types.Message) error {
		w, ok := msg.Type.(W)
		if !ok {
			return nil
		}

		return handler(from, w.Unwrap())
	})
}

func UponNewBlockMessageReceived(m dsl.Module, handler func(from types1.NodeID, block *types3.Block) error) {
	UponMessageReceived[*types.Message_NewBlock](m, func(from types1.NodeID, msg *types.NewBlockMessage) error {
		return handler(from, msg.Block)
	})
}