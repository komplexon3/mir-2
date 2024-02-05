package utils

import (
	blockchainpbtypes "github.com/filecoin-project/mir/pkg-blockchain/blockchainpb/types"
	"github.com/mitchellh/hashstructure"
)

func HashBlock(block *blockchainpbtypes.Block) uint64 {
	hashBlock := *block
	hashBlock.BlockId = 0
	hash, err := hashstructure.Hash(hashBlock, nil)
	if err != nil {
		panic(err)
	}
	return hash
}
