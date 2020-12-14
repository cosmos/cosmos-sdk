package conversion

import (
	"fmt"
	"strconv"
	"strings"

	types2 "github.com/cosmos/cosmos-sdk/codec/types"

	"github.com/coinbase/rosetta-sdk-go/types"

	"github.com/cosmos/cosmos-sdk/server/rosetta"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// RosettaOperationsToSdkMsg converts rosetta operations to sdk.Msg and coins
func RosettaOperationsToSdkMsg(ir types2.InterfaceRegistry, ops []*types.Operation) ([]sdk.Msg, string, sdk.Coins, error) {
	var feeAmnt []*types.Amount
	var newOps []*types.Operation
	if len(ops)%2 == 0 {
		msgs, signAddr, err := ConvertOpsToMsgs(ops)
		return msgs, signAddr, nil, err
	}

	if len(ops)%2 == 1 {
		for _, op := range ops {
			switch op.Type {
			case rosetta.OperationFee:
				amount := op.Amount
				feeAmnt = append(feeAmnt, amount)
			default:
				newOps = append(newOps, op)
			}
		}
	}
	msgs, signAddr, err := ConvertOpsToMsgs(newOps)
	if err != nil {
		return nil, "", nil, err
	}

	return msgs, signAddr, RosettaAmountsToCoins(feeAmnt), nil
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

func ConvertOpsToMsgs(ops []*types.Operation) ([]sdk.Msg, string, error) {
	var msgs []sdk.Msg
	var signAddr string
	var sendOps []*types.Operation
	var delOps []*types.Operation
	for _, op := range ops {
		switch op.Type {
		case "cosmos.bank.v1beta1.MsgSend": // TODO temporary proto Message Name.
			sendOps = append(sendOps, op)
		case rosetta.OperationDelegate:
			delOps = append(delOps, op)
		}
	}
	if len(sendOps) == 2 {
		sendMsg, err := RosettaOperationsToSdkBankMsgSend(sendOps)
		if err != nil {
			return nil, "", err
		}
		msgs = append(msgs, sendMsg)
		signAddr = sendMsg.FromAddress
	}

	if len(delOps) == 2 {
		delMsg, err := RosettaOperationsToSdkStakingMsgDelegate(delOps)
		if err != nil {
			return nil, "", err
		}
		msgs = append(msgs, delMsg)
		signAddr = delMsg.DelegatorAddress
	}

	return msgs, signAddr, nil
}

func RosettaOperationsToSdkStakingMsgDelegate(ops []*types.Operation) (*stakingtypes.MsgDelegate, error) {
	var (
		delAddr sdk.AccAddress
		valAddr sdk.ValAddress
		sendAmt sdk.Coin
		err     error
	)

	for _, op := range ops {
		if strings.HasPrefix(op.Amount.Value, "-") {
			delAddr, err = sdk.AccAddressFromBech32(op.Account.Address)
			if err != nil {
				return nil, err
			}
			continue
		}

		valAddr, err = sdk.ValAddressFromBech32(op.Account.Address)
		if err != nil {
			return nil, err
		}

		amount, err := strconv.ParseInt(op.Amount.Value, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid amount")
		}

		sendAmt = sdk.NewCoin(op.Amount.Currency.Symbol, sdk.NewInt(amount))

	}

	return stakingtypes.NewMsgDelegate(delAddr, valAddr, sendAmt), nil
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
