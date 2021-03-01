package rosetta

import (
	"github.com/coinbase/rosetta-sdk-go/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// sdkCoinsToRosettaAmounts converts []sdk.Coin to rosetta amounts
func sdkCoinsToRosettaAmounts(ownedCoins []sdk.Coin, availableCoins sdk.Coins) []*types.Amount {
	amounts := make([]*types.Amount, len(availableCoins))
	ownedCoinsMap := make(map[string]sdk.Int, len(availableCoins))

	for _, ownedCoin := range ownedCoins {
		ownedCoinsMap[ownedCoin.Denom] = ownedCoin.Amount
	}

	for i, coin := range availableCoins {
		value, owned := ownedCoinsMap[coin.Denom]
		if !owned {
			amounts[i] = &types.Amount{
				Value: sdk.NewInt(0).String(),
				Currency: &types.Currency{
					Symbol: coin.Denom,
				},
			}
			continue
		}
		amounts[i] = &types.Amount{
			Value: value.String(),
			Currency: &types.Currency{
				Symbol: coin.Denom,
			},
		}
	}

	return amounts
}

// sdkTxToOperations converts an sdk.Tx to rosetta operations
func sdkTxToOperations(tx sdk.Tx, withStatus, hasError bool) []*types.Operation {
	var operations []*types.Operation

	msgOps := sdkMsgsToRosettaOperations(tx.GetMsgs(), withStatus, hasError)
	operations = append(operations, msgOps...)

	feeTx := tx.(sdk.FeeTx)
	feeOps := sdkFeeTxToOperations(feeTx, withStatus, len(msgOps))
	operations = append(operations, feeOps...)

	return operations
}

// sdkFeeTxToOperations converts sdk.FeeTx to rosetta operations
func sdkFeeTxToOperations(feeTx sdk.FeeTx, withStatus bool, previousOps int) []*types.Operation {
	feeCoins := feeTx.GetFee()
	var ops []*types.Operation
	if feeCoins != nil {
		var feeOps = rosettaFeeOperationsFromCoins(feeCoins, feeTx.FeePayer().String(), withStatus, previousOps)
		ops = append(ops, feeOps...)
	}

	return ops
}

// rosettaFeeOperationsFromCoins returns the list of rosetta fee operations given sdk coins
func rosettaFeeOperationsFromCoins(coins sdk.Coins, account string, withStatus bool, previousOps int) []*types.Operation {
	feeOps := make([]*types.Operation, 0)
	var status string
	if withStatus {
		status = StatusTxSuccess
	}

	for i, coin := range coins {
		op := &types.Operation{
			OperationIdentifier: &types.OperationIdentifier{
				Index: int64(previousOps + i),
			},
			Type:   OperationFee,
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

// sdkMsgsToRosettaOperations converts sdk messages to rosetta operations
func sdkMsgsToRosettaOperations(msgs []sdk.Msg, withStatus bool, hasError bool) []*types.Operation {
	var operations []*types.Operation
	for _, msg := range msgs {
		if rosettaMsg, ok := msg.(Msg); ok {
			operations = append(operations, rosettaMsg.ToOperations(withStatus, hasError)...)
		}
	}

	return operations
}
