package rosetta

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gogo/protobuf/jsonpb"

	"github.com/coinbase/rosetta-sdk-go/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// opsToMsgsAndFees converts rosetta operations to sdk.Msg and fees represented as sdk.Coins
func opsToMsgsAndFees(interfaceRegistry jsonpb.AnyResolver, ops []*types.Operation) ([]sdk.Msg, sdk.Coins, error) {
	var feeAmnt []*types.Amount
	var newOps []*types.Operation
	var msgType string
	// find the fee operation and put it aside
	for _, op := range ops {
		switch op.Type {
		case OperationFee:
			amount := op.Amount
			feeAmnt = append(feeAmnt, amount)
		default:
			// check if operation matches the one already used
			// as, at the moment, we only support operations
			// that represent a single cosmos-sdk message
			switch {
			// if msgType was not set then set it
			case msgType == "":
				msgType = op.Type
			// if msgType does not match op.Type then it means we're trying to send multiple messages in a single tx
			case msgType != op.Type:
				return nil, nil, fmt.Errorf("only single message operations are supported: %s - %s", msgType, op.Type)
			}
			// append operation to new ops list
			newOps = append(newOps, op)
		}
	}
	// convert all operations, except fee op to sdk.Msgs
	msgs, err := opsToMsgs(interfaceRegistry, newOps)
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

func opsToMsgs(interfaceRegistry jsonpb.AnyResolver, ops []*types.Operation) ([]sdk.Msg, error) {
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
