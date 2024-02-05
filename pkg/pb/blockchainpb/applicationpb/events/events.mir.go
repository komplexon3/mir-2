// Code generated by Mir codegen. DO NOT EDIT.

package applicationpbevents

import (
	types3 "github.com/filecoin-project/mir/pkg/pb/blockchainpb/applicationpb/types"
	types4 "github.com/filecoin-project/mir/pkg/pb/blockchainpb/payloadpb/types"
	types "github.com/filecoin-project/mir/pkg/pb/blockchainpb/statepb/types"
	types1 "github.com/filecoin-project/mir/pkg/pb/blockchainpb/types"
	types2 "github.com/filecoin-project/mir/pkg/pb/eventpb/types"
	stdtypes "github.com/filecoin-project/mir/stdtypes"
)

func VerifyBlocksRequest(destModule stdtypes.ModuleID, checkpointState *types.State, chainCheckpointToStart []*types1.Block, chainToVerify []*types1.Block) *types2.Event {
	return &types2.Event{
		DestModule: destModule,
		Type: &types2.Event_Application{
			Application: &types3.Event{
				Type: &types3.Event_VerifyBlockRequest{
					VerifyBlockRequest: &types3.VerifyBlocksRequest{
						CheckpointState:        checkpointState,
						ChainCheckpointToStart: chainCheckpointToStart,
						ChainToVerify:          chainToVerify,
					},
				},
			},
		},
	}
}

func VerifyBlocksResponse(destModule stdtypes.ModuleID, verifiedBlocks []*types1.Block) *types2.Event {
	return &types2.Event{
		DestModule: destModule,
		Type: &types2.Event_Application{
			Application: &types3.Event{
				Type: &types3.Event_VerifyBlockResponse{
					VerifyBlockResponse: &types3.VerifyBlocksResponse{
						VerifiedBlocks: verifiedBlocks,
					},
				},
			},
		},
	}
}

func PayloadRequest(destModule stdtypes.ModuleID, headId uint64) *types2.Event {
	return &types2.Event{
		DestModule: destModule,
		Type: &types2.Event_Application{
			Application: &types3.Event{
				Type: &types3.Event_PayloadRequest{
					PayloadRequest: &types3.PayloadRequest{
						HeadId: headId,
					},
				},
			},
		},
	}
}

func PayloadResponse(destModule stdtypes.ModuleID, headId uint64, payload *types4.Payload) *types2.Event {
	return &types2.Event{
		DestModule: destModule,
		Type: &types2.Event_Application{
			Application: &types3.Event{
				Type: &types3.Event_PayloadResponse{
					PayloadResponse: &types3.PayloadResponse{
						HeadId:  headId,
						Payload: payload,
					},
				},
			},
		},
	}
}

func HeadChange(destModule stdtypes.ModuleID, removedChain []*types1.Block, addedChain []*types1.Block, checkpointToForkRoot []*types1.Block, checkpointState *types.State) *types2.Event {
	return &types2.Event{
		DestModule: destModule,
		Type: &types2.Event_Application{
			Application: &types3.Event{
				Type: &types3.Event_HeadChange{
					HeadChange: &types3.HeadChange{
						RemovedChain:         removedChain,
						AddedChain:           addedChain,
						CheckpointToForkRoot: checkpointToForkRoot,
						CheckpointState:      checkpointState,
					},
				},
			},
		},
	}
}

func MessageInput(destModule stdtypes.ModuleID, text string) *types2.Event {
	return &types2.Event{
		DestModule: destModule,
		Type: &types2.Event_Application{
			Application: &types3.Event{
				Type: &types3.Event_MessageInput{
					MessageInput: &types3.MessageInput{
						Text: text,
					},
				},
			},
		},
	}
}
