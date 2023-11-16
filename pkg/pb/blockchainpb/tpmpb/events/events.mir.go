// Code generated by Mir codegen. DO NOT EDIT.

package tpmpbevents

import (
	types2 "github.com/filecoin-project/mir/pkg/pb/blockchainpb/tpmpb/types"
	types1 "github.com/filecoin-project/mir/pkg/pb/eventpb/types"
	types "github.com/filecoin-project/mir/pkg/types"
)

func NewBlockRequest(destModule types.ModuleID, headId uint64) *types1.Event {
	return &types1.Event{
		DestModule: destModule,
		Type: &types1.Event_Tpm{
			Tpm: &types2.Event{
				Type: &types2.Event_NewBlockRequest{
					NewBlockRequest: &types2.NewBlockRequest{
						HeadId: headId,
					},
				},
			},
		},
	}
}
