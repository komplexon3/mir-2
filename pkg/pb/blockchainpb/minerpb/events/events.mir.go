// Code generated by Mir codegen. DO NOT EDIT.

package minerpbevents

import (
	types2 "github.com/filecoin-project/mir/pkg/pb/blockchainpb/minerpb/types"
	types "github.com/filecoin-project/mir/pkg/pb/blockchainpb/payloadpb/types"
	types1 "github.com/filecoin-project/mir/pkg/pb/eventpb/types"
	stdtypes "github.com/filecoin-project/mir/stdtypes"
)

func BlockRequest(destModule stdtypes.ModuleID, headId uint64, payload *types.Payload) *types1.Event {
	return &types1.Event{
		DestModule: destModule,
		Type: &types1.Event_Miner{
			Miner: &types2.Event{
				Type: &types2.Event_BlockRequest{
					BlockRequest: &types2.BlockRequest{
						HeadId:  headId,
						Payload: payload,
					},
				},
			},
		},
	}
}

func NewHead(destModule stdtypes.ModuleID, headId uint64) *types1.Event {
	return &types1.Event{
		DestModule: destModule,
		Type: &types1.Event_Miner{
			Miner: &types2.Event{
				Type: &types2.Event_NewHead{
					NewHead: &types2.NewHead{
						HeadId: headId,
					},
				},
			},
		},
	}
}
