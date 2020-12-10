package cosmos

import (
	"fmt"
	"github.com/coinbase/rosetta-sdk-go/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"strconv"
	"strings"
)

func operationsToSdkMsgs(ops []*types.Operation) ([]sdk.Msg, string, sdk.Coins, error) {
	var feeAmnt []*types.Amount
	var newOps []*types.Operation
	if len(ops)%2 == 0 {
		msgs, signAddr, err := convertOpsToMsgs(ops)
		return msgs, signAddr, nil, err
	}

	if len(ops)%2 == 1 {
		for _, op := range ops {
			switch op.Type {
			case OperationFee:
				amount := op.Amount
				feeAmnt = append(feeAmnt, amount)
			default:
				newOps = append(newOps, op)
			}
		}
	}
	msgs, signAddr, err := convertOpsToMsgs(newOps)
	if err != nil {
		return nil, "", nil, err
	}

	return msgs, signAddr, amountsToCoins(feeAmnt), nil
}

func convertOpsToMsgs(ops []*types.Operation) ([]sdk.Msg, string, error) {
	var msgs []sdk.Msg
	var signAddr string
	var sendOps []*types.Operation
	var delOps []*types.Operation
	for _, op := range ops {
		switch op.Type {
		case opBankSend:
			sendOps = append(sendOps, op)
		case opDelegate:
			delOps = append(delOps, op)
		}
	}
	if len(sendOps) == 2 {
		sendMsg, err := operationsToSdkBankMsgSend(sendOps)
		if err != nil {
			return nil, "", err
		}
		msgs = append(msgs, sendMsg)
		signAddr = sendMsg.FromAddress.String()
	}

	if len(delOps) == 2 {
		delMsg, err := operationsToSdkStakingMsgDelegate(delOps)
		if err != nil {
			return nil, "", err
		}
		msgs = append(msgs, delMsg)
		signAddr = delMsg.DelegatorAddress.String()
	}

	return msgs, signAddr, nil
}

// operationsToSdkBankMsgSend extracts the from and to addresses from a list of operations.
// We assume that it comes formated in the correct way. And that the balance of the sender is the same
// as the receiver operations.
func operationsToSdkBankMsgSend(ops []*types.Operation) (bank.MsgSend, error) {
	var (
		from, to sdk.AccAddress
		sendAmt  sdk.Coin
		err      error
	)

	for _, op := range ops {
		if strings.HasPrefix(op.Amount.Value, "-") {
			from, err = sdk.AccAddressFromBech32(op.Account.Address)
			if err != nil {
				return bank.MsgSend{}, err
			}
			continue
		}

		to, err = sdk.AccAddressFromBech32(op.Account.Address)
		if err != nil {
			return bank.MsgSend{}, err
		}

		amount, err := strconv.ParseInt(op.Amount.Value, 10, 64)
		if err != nil {
			return bank.MsgSend{}, fmt.Errorf("invalid amount")
		}

		sendAmt = sdk.NewCoin(op.Amount.Currency.Symbol, sdk.NewInt(amount))

	}

	return bank.MsgSend{FromAddress: from, ToAddress: to, Amount: sdk.NewCoins(sendAmt)}, nil
}

func operationsToSdkStakingMsgDelegate(ops []*types.Operation) (stakingtypes.MsgDelegate, error) {
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
				return stakingtypes.MsgDelegate{}, err
			}
			continue
		}

		valAddr, err = sdk.ValAddressFromBech32(op.Account.Address)
		if err != nil {
			return stakingtypes.MsgDelegate{}, err
		}

		amount, err := strconv.ParseInt(op.Amount.Value, 10, 64)
		if err != nil {
			return stakingtypes.MsgDelegate{}, fmt.Errorf("invalid amount")
		}

		sendAmt = sdk.NewCoin(op.Amount.Currency.Symbol, sdk.NewInt(amount))

	}

	return stakingtypes.NewMsgDelegate(delAddr, valAddr, sendAmt), nil
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

type payloadReqMeta struct {
	ChainID       string
	Sequence      uint64
	AccountNumber uint64
	Gas           uint64
	Memo          string
}

// GetMetadataFromPayloadReq obtains the metadata from the request to /construction/payloads endpoint.
func GetMetadataFromPayloadReq(metadata map[string]interface{}) (*payloadReqMeta, error) {
	chainID, ok := metadata[ChainIDKey].(string)
	if !ok {
		return nil, fmt.Errorf("chain_id metadata was not provided")
	}

	sequence, ok := metadata[SequenceKey]
	if !ok {
		return nil, fmt.Errorf("sequence metadata was not provided")
	}

	seqNum, ok := sequence.(float64)
	if !ok {
		return nil, fmt.Errorf("invalid sequence value")
	}

	accountNum, ok := metadata[AccountNumberKey]
	if !ok {
		return nil, fmt.Errorf("account_number metadata was not provided")
	}
	accNum, ok := accountNum.(float64)
	if !ok {
		return nil, fmt.Errorf("invalid account_number value")
	}

	gasNum, ok := metadata[GasKey]
	if !ok {
		return nil, fmt.Errorf("gas metadata was not provided")
	}
	gasF64, ok := gasNum.(float64)
	if !ok {
		return nil, fmt.Errorf("invalid gas value")
	}

	memo, ok := metadata["memo"]
	if !ok {
		memo = ""
	}
	memoStr, ok := memo.(string)
	if !ok {
		return nil, fmt.Errorf("invalid account_number value")
	}

	return &payloadReqMeta{
		ChainID:       chainID,
		Sequence:      uint64(seqNum),
		AccountNumber: uint64(accNum),
		Gas:           uint64(gasF64),
		Memo:          memoStr,
	}, nil
}
