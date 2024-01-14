// Code generated by Mir codegen. DO NOT EDIT.

package bcmpbevents

import (
	blockchainpb "github.com/filecoin-project/mir/pkg/pb/blockchainpb"
	types2 "github.com/filecoin-project/mir/pkg/pb/blockchainpb/bcmpb/types"
	statepb "github.com/filecoin-project/mir/pkg/pb/blockchainpb/statepb"
	types1 "github.com/filecoin-project/mir/pkg/pb/eventpb/types"
	types "github.com/filecoin-project/mir/pkg/types"
)

func NewBlock(destModule types.ModuleID, block *blockchainpb.Block) *types1.Event {
	return &types1.Event{
		DestModule: destModule,
		Type: &types1.Event_Bcm{
			Bcm: &types2.Event{
				Type: &types2.Event_NewBlock{
					NewBlock: &types2.NewBlock{
						Block: block,
					},
				},
			},
		},
	}
}

func NewChain(destModule types.ModuleID, blocks []*blockchainpb.Block) *types1.Event {
	return &types1.Event{
		DestModule: destModule,
		Type: &types1.Event_Bcm{
			Bcm: &types2.Event{
				Type: &types2.Event_NewChain{
					NewChain: &types2.NewChain{
						Blocks: blocks,
					},
				},
			},
		},
	}
}

func GetBlockRequest(destModule types.ModuleID, requestId string, sourceModule types.ModuleID, blockId uint64) *types1.Event {
	return &types1.Event{
		DestModule: destModule,
		Type: &types1.Event_Bcm{
			Bcm: &types2.Event{
				Type: &types2.Event_GetBlockRequest{
					GetBlockRequest: &types2.GetBlockRequest{
						RequestId:    requestId,
						SourceModule: sourceModule,
						BlockId:      blockId,
					},
				},
			},
		},
	}
}

func GetBlockResponse(destModule types.ModuleID, requestId string, found bool, block *blockchainpb.Block) *types1.Event {
	return &types1.Event{
		DestModule: destModule,
		Type: &types1.Event_Bcm{
			Bcm: &types2.Event{
				Type: &types2.Event_GetBlockResponse{
					GetBlockResponse: &types2.GetBlockResponse{
						RequestId: requestId,
						Found:     found,
						Block:     block,
					},
				},
			},
		},
	}
}

func GetChainRequest(destModule types.ModuleID, requestId string, sourceModule types.ModuleID, endBlockId uint64, sourceBlockIds []uint64) *types1.Event {
	return &types1.Event{
		DestModule: destModule,
		Type: &types1.Event_Bcm{
			Bcm: &types2.Event{
				Type: &types2.Event_GetChainRequest{
					GetChainRequest: &types2.GetChainRequest{
						RequestId:      requestId,
						SourceModule:   sourceModule,
						EndBlockId:     endBlockId,
						SourceBlockIds: sourceBlockIds,
					},
				},
			},
		},
	}
}

func GetChainResponse(destModule types.ModuleID, requestId string, success bool, chain []*blockchainpb.Block) *types1.Event {
	return &types1.Event{
		DestModule: destModule,
		Type: &types1.Event_Bcm{
			Bcm: &types2.Event{
				Type: &types2.Event_GetChainResponse{
					GetChainResponse: &types2.GetChainResponse{
						RequestId: requestId,
						Success:   success,
						Chain:     chain,
					},
				},
			},
		},
	}
}

func GetHeadToCheckpointChainRequest(destModule types.ModuleID, requestId string, sourceModule types.ModuleID) *types1.Event {
	return &types1.Event{
		DestModule: destModule,
		Type: &types1.Event_Bcm{
			Bcm: &types2.Event{
				Type: &types2.Event_GetHeadToCheckpointChainRequest{
					GetHeadToCheckpointChainRequest: &types2.GetHeadToCheckpointChainRequest{
						RequestId:    requestId,
						SourceModule: sourceModule,
					},
				},
			},
		},
	}
}

func GetHeadToCheckpointChainResponse(destModule types.ModuleID, requestId string, chain []*blockchainpb.BlockInternal) *types1.Event {
	return &types1.Event{
		DestModule: destModule,
		Type: &types1.Event_Bcm{
			Bcm: &types2.Event{
				Type: &types2.Event_GetHeadToCheckpointChainResponse{
					GetHeadToCheckpointChainResponse: &types2.GetHeadToCheckpointChainResponse{
						RequestId: requestId,
						Chain:     chain,
					},
				},
			},
		},
	}
}

func RegisterCheckpoint(destModule types.ModuleID, blockId uint64, state *statepb.State) *types1.Event {
	return &types1.Event{
		DestModule: destModule,
		Type: &types1.Event_Bcm{
			Bcm: &types2.Event{
				Type: &types2.Event_RegisterCheckpoint{
					RegisterCheckpoint: &types2.RegisterCheckpoint{
						BlockId: blockId,
						State:   state,
					},
				},
			},
		},
	}
}

func InitBlockchain(destModule types.ModuleID, initialState *statepb.State) *types1.Event {
	return &types1.Event{
		DestModule: destModule,
		Type: &types1.Event_Bcm{
			Bcm: &types2.Event{
				Type: &types2.Event_InitBlockchain{
					InitBlockchain: &types2.InitBlockchain{
						InitialState: initialState,
					},
				},
			},
		},
	}
}
