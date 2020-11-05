package conversion

import (
	"fmt"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/cosmos/cosmos-sdk/server/rosetta"
	sdk "github.com/cosmos/cosmos-sdk/types"
	tmcoretypes "github.com/tendermint/tendermint/rpc/core/types"
	tmtypes "github.com/tendermint/tendermint/types"
	"time"
)

// TimeToMilliseconds converts time to milliseconds timestamp
func TimeToMilliseconds(t time.Time) int64 {
	return t.UnixNano() / (int64(time.Millisecond) / int64(time.Nanosecond))
}

// CoinsToBalance converts sdk.Coins to rosetta.Amounts
func CoinsToBalance(coins []sdk.Coin) []*types.Amount {
	amounts := make([]*types.Amount, len(coins))

	for i, coin := range coins {
		amounts[i] = &types.Amount{
			Value: coin.Amount.String(),
			Currency: &types.Currency{
				Symbol: coin.Denom,
			},
		}
	}

	return amounts
}

// ResultTxSearchToTransaction converts tendermint search transactions to rosetta ones
func ResultTxSearchToTransaction(txs []*rosetta.SdkTxWithHash) []*types.Transaction {
	converted := make([]*types.Transaction, len(txs))
	for i, tx := range txs {
		converted[i] = &types.Transaction{
			TransactionIdentifier: &types.TransactionIdentifier{Hash: tx.HexHash},
			Operations:            SdkTxToOperations(tx.Tx),
			Metadata:              nil,
		}
	}
	return converted
}

// SdkTxResponseToOperations converts a tx response to operations
func SdkTxToOperations(tx sdk.Tx) []*types.Operation {
	msgs := tx.GetMsgs()
	ops := make([]*types.Operation, len(msgs))
	for i, msg := range msgs {
		// TODO: assert if balance change op?
		ops[i] = &types.Operation{
			OperationIdentifier: &types.OperationIdentifier{
				Index:        int64(i),
				NetworkIndex: nil,
			},
			RelatedOperations: nil,
			Type:              msg.Type(),
			Status:            "",
			Account: &types.AccountIdentifier{
				Address:    "",
				SubAccount: nil,
				Metadata:   nil,
			},
			Amount:     nil,
			CoinChange: nil,
			Metadata:   nil,
		}
	}
	return ops
}

// TendermintTxsToTxIdentifiers converts a tendermint raw transaction into a rosetta tx identifier
func TendermintTxsToTxIdentifiers(txs []tmtypes.Tx) []*types.TransactionIdentifier {
	converted := make([]*types.TransactionIdentifier, len(txs))
	for i, tx := range txs {
		converted[i] = &types.TransactionIdentifier{Hash: fmt.Sprintf("%x", tx.Hash())} // TODO hash is sha256, so we hex it?
	}
	return converted
}

// TendermintBlockToBlockIdentifier converts a tendermint result block to a rosetta block identifier
func TendermintBlockToBlockIdentifier(block *tmcoretypes.ResultBlock) *types.BlockIdentifier {
	return &types.BlockIdentifier{
		Index: block.Block.Height,
		Hash:  block.Block.Hash().String(),
	}
}
