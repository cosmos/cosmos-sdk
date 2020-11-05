package conversion

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/server/rosetta"
	sdk "github.com/cosmos/cosmos-sdk/types"
	tmcoretypes "github.com/tendermint/tendermint/rpc/core/types"
	tmtypes "github.com/tendermint/tendermint/types"
)

// TmTxToSdkTx converts a tendermint transaction to a cosmos one
func TmTxToSdkTx(decode sdk.TxDecoder, tx tmtypes.Tx) (sdk.Tx, error) {
	return decode(tx)
}

// TmTxsToSdkTxs converts multiple tendermint transactions to cosmos ones
func TmTxsToSdkTxs(decode sdk.TxDecoder, txs []tmtypes.Tx) ([]sdk.Tx, error) {
	converted := make([]sdk.Tx, len(txs))
	for i, tx := range txs {
		sdkTx, err := decode(tx)
		if err != nil {
			return nil, err
		}
		converted[i] = sdkTx
	}
	return converted, nil
}

// TmResultTxsToSdkTxs converts tendermint result txs to cosmos sdk.Tx
func TmResultTxsToSdkTxs(decode sdk.TxDecoder, txs []*tmcoretypes.ResultTx) ([]*rosetta.SdkTxWithHash, error) {
	converted := make([]*rosetta.SdkTxWithHash, len(txs))
	for i, tx := range txs {
		sdkTx, err := decode(tx.Tx)
		if err != nil {
			return nil, err
		}
		converted[i] = &rosetta.SdkTxWithHash{
			HexHash: fmt.Sprintf("%X", tx.Tx.Hash()),
			Tx:      sdkTx,
		}
	}
	return converted, nil
}
