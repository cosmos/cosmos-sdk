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
		msgs, signAddr, err := ConvertOpsToMsgs(ir, ops)
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
	msgs, signAddr, err := ConvertOpsToMsgs(nil, newOps)
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

func ConvertOpsToMsgs(ir types2.InterfaceRegistry, ops []*types.Operation) ([]sdk.Msg, string, error) {
	var msgs []sdk.Msg
	var signAddr string
	var operationsByType = make(map[string][]*types.Operation)
	for _, op := range ops {
		operationsByType[op.Type] = append(operationsByType[op.Type], op)
	}

	for opName, operations := range operationsByType {
		if len(operations) == 2 {
			if opName == "cosmos.bank.v1beta1.MsgSend" {
				sendMsg, fromAddr, err := RosettaOperationsToSdkBankMsgSend(operations)
				if err != nil {
					return nil, "", err
				}
				msgs = append(msgs, sendMsg)
				signAddr = fromAddr
			} else if opName == "cosmos.staking.v1beta1.MsgDelegate" {
				delMsg, fromAddr, err := RosettaOperationsToSdkStakingMsgDelegate(operations)
				if err != nil {
					return nil, "", err
				}
				msgs = append(msgs, delMsg)
				signAddr = fromAddr
			}
		}
	}

	return msgs, signAddr, nil
}

func RosettaOperationsToSdkStakingMsgDelegate(ops []*types.Operation) (sdk.Msg, string, error) {
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
				return nil, "", err
			}
			continue
		}

		valAddr, err = sdk.ValAddressFromBech32(op.Account.Address)
		if err != nil {
			return nil, "", err
		}

		amount, err := strconv.ParseInt(op.Amount.Value, 10, 64)
		if err != nil {
			return nil, "", fmt.Errorf("invalid amount")
		}

		sendAmt = sdk.NewCoin(op.Amount.Currency.Symbol, sdk.NewInt(amount))

	}

	return stakingtypes.NewMsgDelegate(delAddr, valAddr, sendAmt), delAddr.String(), nil
}

// RosettaOperationsToSdkBankMsgSend extracts the from and to addresses from a list of operations.
// We assume that it comes formated in the correct way. And that the balance of the sender is the same
// as the receiver operations.
func RosettaOperationsToSdkBankMsgSend(ops []*types.Operation) (sdk.Msg, string, error) {
	var (
		from, to sdk.AccAddress
		sendAmt  sdk.Coin
		err      error
	)

	for _, op := range ops {
		if strings.HasPrefix(op.Amount.Value, "-") {
			from, err = sdk.AccAddressFromBech32(op.Account.Address)
			if err != nil {
				return nil, "", err
			}
			continue
		}

		to, err = sdk.AccAddressFromBech32(op.Account.Address)
		if err != nil {
			return nil, "", err
		}

		amount, err := strconv.ParseInt(op.Amount.Value, 10, 64)
		if err != nil {
			return nil, "", fmt.Errorf("invalid amount")
		}

		sendAmt = sdk.NewCoin(op.Amount.Currency.Symbol, sdk.NewInt(amount))

	}

	return banktypes.NewMsgSend(from, to, sdk.NewCoins(sendAmt)), from.String(), nil
}
