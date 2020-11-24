package client

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/server/rosetta"
	sdk "github.com/cosmos/cosmos-sdk/types"
	tmcoretypes "github.com/tendermint/tendermint/rpc/core/types"
)

// tmResultTxsToSdkTxsWithHash converts tendermint result txs to cosmos sdk.Tx
func tmResultTxsToSdkTxsWithHash(decode sdk.TxDecoder, txs []*tmcoretypes.ResultTx) ([]*rosetta.SdkTxWithHash, error) {
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
