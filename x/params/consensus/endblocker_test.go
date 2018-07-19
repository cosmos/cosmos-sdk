package consensus

import (
	"testing"

	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params/store"
)

type flatConsensusParams struct {
	blockMaxBytes int32
	blockMaxTxs   int32
	blockMaxGas   int64
	txMaxBytes    int32
	txMaxGas      int64
	partSizeBytes int32
}

func flat(params *abci.ConsensusParams) (res flatConsensusParams) {
	if params == nil {
		return
	}

	blockSize := params.BlockSize
	if blockSize != nil {
		res.blockMaxBytes = blockSize.MaxBytes
		res.blockMaxTxs = blockSize.MaxTxs
		res.blockMaxGas = blockSize.MaxGas
	}
	txSize := params.TxSize
	if txSize != nil {
		res.txMaxBytes = txSize.MaxBytes
		res.txMaxGas = txSize.MaxGas
	}
	gossip := params.BlockGossip
	if gossip != nil {
		res.partSizeBytes = gossip.BlockPartSizeBytes
	}

	return
}

func override(original flatConsensusParams, updates flatConsensusParams) (res flatConsensusParams) {
	res = original

	if updates.blockMaxBytes != 0 {
		res.blockMaxBytes = updates.blockMaxBytes
	}
	if updates.blockMaxTxs != 0 {
		res.blockMaxTxs = updates.blockMaxTxs
	}
	if updates.blockMaxGas != 0 {
		res.blockMaxGas = updates.blockMaxGas
	}
	if updates.txMaxBytes != 0 {
		res.txMaxBytes = updates.txMaxBytes
	}
	if updates.txMaxGas != 0 {
		res.txMaxGas = updates.txMaxGas
	}
	if updates.partSizeBytes != 0 {
		res.partSizeBytes = updates.partSizeBytes
	}

	return
}

func setParams(ctx sdk.Context, store store.Store, params flatConsensusParams) {
	if params.blockMaxBytes != 0 {
		store.Set(ctx, blockMaxBytesKey, params.blockMaxBytes)
	}
	if params.blockMaxTxs != 0 {
		store.Set(ctx, blockMaxTxsKey, params.blockMaxTxs)
	}
	if params.blockMaxGas != 0 {
		store.Set(ctx, blockMaxGasKey, params.blockMaxGas)
	}
	if params.txMaxBytes != 0 {
		store.Set(ctx, txMaxBytesKey, params.txMaxBytes)
	}
	if params.txMaxGas != 0 {
		store.Set(ctx, txMaxGasKey, params.txMaxGas)
	}
	if params.partSizeBytes != 0 {
		store.Set(ctx, blockPartSizeBytesKey, params.partSizeBytes)
	}
}

func TestEndBlocker(t *testing.T) {
	ctx, store, commit := store.DefaultTestComponents(t)

	empty := &abci.ConsensusParams{
		BlockSize: &abci.BlockSize{
			MaxBytes: 0,
			MaxTxs:   0,
			MaxGas:   0,
		},
		TxSize: &abci.TxSize{
			MaxBytes: 0,
			MaxGas:   0,
		},
		BlockGossip: &abci.BlockGossip{
			BlockPartSizeBytes: 0,
		},
	}

	cases := []*abci.ConsensusParams{
		nil,
		&abci.ConsensusParams{
			BlockSize:   nil,
			TxSize:      nil,
			BlockGossip: nil,
		},
		empty,
		&abci.ConsensusParams{
			BlockSize:   &abci.BlockSize{1, 2, 3},
			TxSize:      &abci.TxSize{4, 5},
			BlockGossip: &abci.BlockGossip{6},
		},
		&abci.ConsensusParams{
			BlockSize:   nil,
			TxSize:      nil,
			BlockGossip: &abci.BlockGossip{10},
		},
	}

	current := flat(empty)
	for _, tc := range cases {
		flatten := flat(tc)
		setParams(ctx, store, flat(tc))
		updates := EndBlock(ctx, store)
		updated := override(current, flat(updates))
		current = override(current, flatten)
		require.Equal(t, current, updated)
		commit()
	}
}
