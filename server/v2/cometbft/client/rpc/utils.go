package rpc

import (
	"fmt"

	tmtypes "buf.build/gen/go/tendermint/tendermint/protocolbuffers/go/tendermint/types"
	abciv1beta1 "cosmossdk.io/api/cosmos/base/abci/v1beta1"
	coretypes "github.com/cometbft/cometbft/rpc/core/types"
	gogoproto "github.com/cosmos/gogoproto/proto"
	protov2 "google.golang.org/protobuf/proto"
)

// formatBlockResults parses the indexed blocks into a slice of BlockResponse objects.
func formatBlockResults(resBlocks []*coretypes.ResultBlock) ([]*tmtypes.Block, error) {
	out := make([]*tmtypes.Block, len(resBlocks))
	for i := range resBlocks {
		out[i] = NewResponseResultBlock(resBlocks[i])
		if out[i] == nil {
			return nil, fmt.Errorf("unable to create response block from comet result block: %v", resBlocks[i])
		}
	}

	return out, nil
}

func NewSearchBlocksResult(totalCount, count, page, limit int64, blocks []*tmtypes.Block) *abciv1beta1.SearchBlocksResult {
	totalPages := calcTotalPages(totalCount, limit)
	return &abciv1beta1.SearchBlocksResult{
		TotalCount: totalCount,
		Count:      count,
		PageNumber: page,
		PageTotal:  totalPages,
		Limit:      limit,
		Blocks:     blocks,
	}
}

// NewResponseResultBlock returns a BlockResponse given a ResultBlock from CometBFT
func NewResponseResultBlock(res *coretypes.ResultBlock) *tmtypes.Block {
	blkProto, err := res.Block.ToProto()
	if err != nil {
		panic(err)
	}
	blkBz, err := gogoproto.Marshal(blkProto)
	if err != nil {
		panic(err)
	}

	blk := &tmtypes.Block{}
	err = protov2.Unmarshal(blkBz, blk)
	if err != nil {
		panic(err)
	}
	return blk
}

// calculate total pages in an overflow safe manner
func calcTotalPages(totalCount, limit int64) int64 {
	totalPages := int64(0)
	if totalCount != 0 && limit != 0 {
		if totalCount%limit > 0 {
			totalPages = totalCount/limit + 1
		} else {
			totalPages = totalCount / limit
		}
	}
	return totalPages
}
