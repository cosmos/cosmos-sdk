package conversion

import (
	"strconv"
	"strings"

	types2 "github.com/cosmos/cosmos-sdk/codec/types"

	"github.com/coinbase/rosetta-sdk-go/types"

	"github.com/cosmos/cosmos-sdk/server/rosetta"
	sdk "github.com/cosmos/cosmos-sdk/types"
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
	msgs, signAddr, err := ConvertOpsToMsgs(ir, newOps)
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
		if opName == rosetta.OperationFee {
			continue
		}

		msgType, err := ir.Resolve("/" + opName) // Types are registered as /proto-name in the interface registry.
		if err != nil {
			return nil, "", err
		}

		if rosettaMsg, ok := msgType.(rosetta.Msg); ok {
			m, fromAddr, err := rosettaMsg.FromOperations(operations)
			if err != nil {
				return nil, "", err
			}
			msgs = append(msgs, m)
			signAddr = fromAddr
		}
	}

	return msgs, signAddr, nil
}
