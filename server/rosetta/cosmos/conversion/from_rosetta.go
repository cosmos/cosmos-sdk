package conversion

import (
	"fmt"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/cosmos/cosmos-sdk/server/rosetta"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"strconv"
	"strings"
)

// RosettaOperationsToSdkMsg converts rosetta operations to sdk.Msg and coins
func RosettaOperationsToSdkMsg(ops []*types.Operation) (sdk.Msg, sdk.Coins, error) {
	var feeAmnt []*types.Amount
	var sendOps []*types.Operation
	if len(ops) == 2 {
		sendMsg, err := RosettaOperationsToSdkBankMsgSend(ops)
		return sendMsg, nil, err
	}

	if len(ops) == 3 {
		for _, op := range ops {
			if op.Type == rosetta.OperationFee {
				amount := op.Amount
				feeAmnt = append(feeAmnt, amount)
			}
			if op.Type == rosetta.OperationSend {
				sendOps = append(sendOps, op)
			}
		}
	}
	sendMsg, err := RosettaOperationsToSdkBankMsgSend(sendOps)
	if err != nil {
		return nil, nil, err
	}

	return sendMsg, RosettaAmountsToCoins(feeAmnt), nil
}

// RosettaAmountsToCoins converts rosetta amounts to sdk coins
func RosettaAmountsToCoins(amounts []*types.Amount) sdk.Coins {
	var feeCoins sdk.Coins

	for _, amount := range amounts {
		absValue := strings.Trim(amount.Value, "-")
		value, err := strconv.ParseInt(absValue, 10, 64)
		if err != nil {
			return nil
		}
		coin := sdk.NewCoin(amount.Currency.Symbol, sdk.NewInt(value))
		feeCoins = append(feeCoins, coin)
	}

	return feeCoins
}

// RosettaOperationsToSdkBankMsgSend extracts the from and to addresses from a list of operations.
// We assume that it comes formated in the correct way. And that the balance of the sender is the same
// as the receiver operations.
func RosettaOperationsToSdkBankMsgSend(ops []*types.Operation) (*banktypes.MsgSend, error) {
	var (
		from, to sdk.AccAddress
		sendAmt  sdk.Coin
		err      error
	)

	for _, op := range ops {
		if strings.HasPrefix(op.Amount.Value, "-") {
			from, err = sdk.AccAddressFromBech32(op.Account.Address)
			if err != nil {
				return nil, err
			}
			continue
		}

		to, err = sdk.AccAddressFromBech32(op.Account.Address)
		if err != nil {
			return nil, err
		}

		amount, err := strconv.ParseInt(op.Amount.Value, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid amount")
		}

		sendAmt = sdk.NewCoin(op.Amount.Currency.Symbol, sdk.NewInt(amount))

	}

	return banktypes.NewMsgSend(from, to, sdk.NewCoins(sendAmt)), nil
}
