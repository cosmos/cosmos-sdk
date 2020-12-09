package cosmos

import (
	"github.com/cosmos/cosmos-sdk/server/rosetta"
	"time"

	"github.com/coinbase/rosetta-sdk-go/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func timeToMilliseconds(t time.Time) int64 {
	return t.UnixNano() / (int64(time.Millisecond) / int64(time.Nanosecond))
}

// sdkTxToOperations converts an sdk.Tx to rosetta operations
func sdkTxToOperations(tx sdk.Tx, withStatus, hasError bool) []*types.Operation {
	var operations []*types.Operation

	feeTx := tx.(auth.StdTx)
	feeCoins := feeTx.Fee.Amount
	var feeOps = rosettaFeeOperationsFromCoins(feeCoins, feeTx.GetSigners()[0].String(), withStatus)
	operations = append(operations, feeOps...)

	msgOps := sdkMsgsToRosettaOperations(tx.GetMsgs(), withStatus, hasError, len(feeCoins))
	operations = append(operations, msgOps...)

	return operations
}

// rosettaFeeOperationsFromCoins returns the list of rosetta fee operations given sdk coins
func rosettaFeeOperationsFromCoins(coins sdk.Coins, account string, withStatus bool) []*types.Operation {
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
			Type:   opFee,
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
func sdkMsgsToRosettaOperations(msgs []sdk.Msg, withStatus bool, hasError bool, feeLen int) []*types.Operation {
	var operations []*types.Operation
	var status string
	for i, msg := range msgs {
		switch msg.Type() { // nolint

		case opBankSend:
			newMsg := msg.(bank.MsgSend)
			fromAddress := newMsg.FromAddress
			toAddress := newMsg.ToAddress
			amounts := newMsg.Amount
			if len(amounts) == 0 {
				continue
			}
			coin := amounts[0]
			sendOp := func(account, amount string, index int) *types.Operation {
				if withStatus {
					status = rosetta.StatusSuccess
					if hasError {
						status = rosetta.StatusReverted
					}
				}
				return &types.Operation{
					OperationIdentifier: &types.OperationIdentifier{
						Index: int64(index),
					},
					Type:   opBankSend,
					Status: status,
					Account: &types.AccountIdentifier{
						Address: account,
					},
					Amount: &types.Amount{
						Value: amount,
						Currency: &types.Currency{
							Symbol: coin.Denom,
						},
					},
				}
			}
			operations = append(operations,
				sendOp(fromAddress.String(), "-"+coin.Amount.String(), feeLen+i),
				sendOp(toAddress.String(), coin.Amount.String(), feeLen+i+1),
			)

		case opDelegate:
			newMsg := msg.(stakingtypes.MsgDelegate)
			delAddr := newMsg.DelegatorAddress
			valAddr := newMsg.ValidatorAddress
			coin := newMsg.Amount
			delOp := func(account, amount string, index int) *types.Operation {
				if withStatus {
					status = rosetta.StatusSuccess
					if hasError {
						status = rosetta.StatusReverted
					}
				}
				return &types.Operation{
					OperationIdentifier: &types.OperationIdentifier{
						Index: int64(index),
					},
					Type:   opDelegate,
					Status: status,
					Account: &types.AccountIdentifier{
						Address: account,
					},
					Amount: &types.Amount{
						Value: amount,
						Currency: &types.Currency{
							Symbol: coin.Denom,
						},
					},
				}
			}
			operations = append(operations,
				delOp(delAddr.String(), "-"+coin.Amount.String(), feeLen+i),
				delOp(valAddr.String(), coin.Amount.String(), feeLen+i+1),
			)
		}
	}

	return operations
}
