package rosetta

import (
	"context"
	"crypto/sha256"

	crgtypes "github.com/tendermint/cosmos-rosetta-gateway/types"

	rosettatypes "github.com/coinbase/rosetta-sdk-go/types"

	crgerrs "github.com/tendermint/cosmos-rosetta-gateway/errors"
)

const (
	deliverTxSize           = sha256.Size
	beginEndBlockTxSize     = deliverTxSize + 1
	endBlockHashStart       = 0x0
	beginBlockHashStart     = 0x1
	burnerAddressIdentifier = "burner"
)

func (c *Client) getTx(ctx context.Context, txHash []byte) (*rosettatypes.Transaction, error) {
	rawTx, err := c.clientCtx.Client.Tx(ctx, txHash, true)
	if err != nil {
		return nil, crgerrs.WrapError(crgerrs.ErrUnknown, err.Error())
	}
	return sdkTxToRosettaTx(c.clientCtx.TxConfig.TxDecoder(), rawTx.Tx, &rawTx.TxResult)
}

func (c *Client) beginBlockTx(ctx context.Context, blockHash []byte) (*rosettatypes.Transaction, error) {
	// get block height by hash
	block, err := c.clientCtx.Client.BlockByHash(ctx, blockHash)
	if err != nil {
		return nil, crgerrs.WrapError(crgerrs.ErrUnknown, err.Error())
	}

	// get block txs
	fullBlock, err := c.blockTxs(ctx, &block.Block.Height)
	if err != nil {
		return nil, err
	}

	return fullBlock.Transactions[0], nil
}

func (c *Client) endBlockTx(ctx context.Context, blockHash []byte) (*rosettatypes.Transaction, error) {
	// get block height by hash
	block, err := c.clientCtx.Client.BlockByHash(ctx, blockHash)
	if err != nil {
		return nil, crgerrs.WrapError(crgerrs.ErrUnknown, err.Error())
	}

	// get block txs
	fullBlock, err := c.blockTxs(ctx, &block.Block.Height)
	if err != nil {
		return nil, err
	}

	// get last tx
	return fullBlock.Transactions[len(fullBlock.Transactions)-1], nil
}

func (c *Client) blockTxs(ctx context.Context, height *int64) (crgtypes.BlockTransactionsResponse, error) {
	// get block info
	blockInfo, err := c.clientCtx.Client.Block(ctx, height)
	if err != nil {
		return crgtypes.BlockTransactionsResponse{}, err
	}
	// get block events
	blockResults, err := c.clientCtx.Client.BlockResults(ctx, height)
	if err != nil {
		return crgtypes.BlockTransactionsResponse{}, err
	}

	if len(blockResults.TxsResults) != len(blockInfo.Block.Txs) {
		// wtf?
		panic("block results transactions do now match block transactions")
	}
	// process begin and end block txs
	beginBlockTx := &rosettatypes.Transaction{
		TransactionIdentifier: &rosettatypes.TransactionIdentifier{Hash: beginBlockTxHash(blockInfo.BlockID.Hash)},
		Operations: normalizeOperationIndexes(
			nil,
			sdkEventsToBalanceOperations(StatusTxSuccess, blockResults.BeginBlockEvents),
		),
	}

	endBlockTx := &rosettatypes.Transaction{
		TransactionIdentifier: &rosettatypes.TransactionIdentifier{Hash: endBlockTxHash(blockInfo.BlockID.Hash)},
		Operations: normalizeOperationIndexes(
			nil,
			sdkEventsToBalanceOperations(StatusTxSuccess, blockResults.EndBlockEvents),
		),
	}

	deliverTx := make([]*rosettatypes.Transaction, len(blockInfo.Block.Txs))
	// process normal txs
	for i, tx := range blockInfo.Block.Txs {
		rosTx, err := sdkTxToRosettaTx(c.clientCtx.TxConfig.TxDecoder(), tx, blockResults.TxsResults[i])
		if err != nil {
			return crgtypes.BlockTransactionsResponse{}, err
		}
		deliverTx[i] = rosTx
	}

	finalTxs := make([]*rosettatypes.Transaction, 0, 2+len(deliverTx))
	finalTxs = append(finalTxs, beginBlockTx)
	finalTxs = append(finalTxs, deliverTx...)
	finalTxs = append(finalTxs, endBlockTx)

	return crgtypes.BlockTransactionsResponse{
		BlockResponse: tmResultBlockToRosettaBlockResponse(blockInfo),
		Transactions:  finalTxs,
	}, nil
}
