package conversion

import (
	"github.com/coinbase/rosetta-sdk-go/types"

	"github.com/cosmos/cosmos-sdk/server/rosetta"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetFeeOpFromCoins
func GetFeeOpFromCoins(coins sdk.Coins, account string, withStatus bool) []*types.Operation {
	feeOps := make([]*types.Operation, 0)
	var status string
	if withStatus {
		status = rosetta.StatusSuccess
	}
	for i, coin := range coins {
		op := &types.Operation{
			OperationIdentifier: &types.OperationIdentifier{
				Index: int64(i),
			},
			Type:   rosetta.OperationFee,
			Status: status,
			Account: &types.AccountIdentifier{
				Address: account,
			},
			Amount: &types.Amount{
				Value: "-" + coin.Amount.String(),
				Currency: &types.Currency{
					Symbol: coin.Denom,
				},
			},
		}
		feeOps = append(feeOps, op)
	}
	return feeOps
}
