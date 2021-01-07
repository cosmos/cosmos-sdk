package rosetta

import (
	"strconv"
	"strings"

	"github.com/gogo/protobuf/jsonpb"

	"github.com/coinbase/rosetta-sdk-go/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// operationsToSdkMsgs converts rosetta operations to sdk.Msg and coins
func operationsToSdkMsgs(interfaceRegistry jsonpb.AnyResolver, ops []*types.Operation) ([]sdk.Msg, sdk.Coins, error) {
	var feeAmnt []*types.Amount
	var newOps []*types.Operation
	// find the fee operation and put it aside
	for _, op := range ops {
		switch op.Type {
		case OperationFee:
			amount := op.Amount
			feeAmnt = append(feeAmnt, amount)
		default:
			newOps = append(newOps, op)
		}
	}
	// convert all operations, except fee op to sdk.Msgs
	msgs, err := ConvertOpsToMsgs(interfaceRegistry, newOps)
	if err != nil {
		return nil, nil, err
	}

	return msgs, amountsToCoins(feeAmnt), nil
}

// amountsToCoins converts rosetta amounts to sdk coins
func amountsToCoins(amounts []*types.Amount) sdk.Coins {
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

func ConvertOpsToMsgs(interfaceRegistry jsonpb.AnyResolver, ops []*types.Operation) ([]sdk.Msg, error) {
	var msgs []sdk.Msg
	var operationsByType = make(map[string][]*types.Operation)
	for _, op := range ops {
		operationsByType[op.Type] = append(operationsByType[op.Type], op)
	}

	for opName, operations := range operationsByType {
		if opName == OperationFee {
			continue
		}

		msgType, err := interfaceRegistry.Resolve("/" + opName) // Types are registered as /proto-name in the interface registry.
		if err != nil {
			return nil, err
		}

		if rosettaMsg, ok := msgType.(Msg); ok {
			m, err := rosettaMsg.FromOperations(operations)
			if err != nil {
				return nil, err
			}
			msgs = append(msgs, m)
		}
	}

	return msgs, nil
}
