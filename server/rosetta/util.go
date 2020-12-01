package rosetta

import (
	"context"
	"fmt"
	"github.com/cosmos/cosmos-sdk/x/auth"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"strconv"
	"strings"

	"github.com/cosmos/cosmos-sdk/x/bank"

	"github.com/cosmos/cosmos-sdk/server/rosetta/client/tendermint"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/coinbase/rosetta-sdk-go/types"
)

const (
	zerox = "0x"
)

// HexPrefix ensures that string representation of hex starts with 0x.
func HexPrefix(hex string) string {
	if !strings.HasPrefix(hex, zerox) {
		return zerox + hex
	}
	return hex
}

// getTxByHash calls
func (l launchpad) getTxByHash(ctx context.Context, hash string) (*types.Transaction, *types.Error) {
	txQuery, err := l.cosmos.GetTx(ctx, hash)
	if err != nil {
		return nil, ErrNodeConnection
	}
	if txQuery.Tx == nil {
		return nil, ErrInvalidTxHash
	}
	tx := cosmosTxToRosettaTx(txQuery)

	return tx, nil
}

func toBlockIdentifier(result tendermint.BlockResponse) (*types.BlockIdentifier, error) {
	if result.BlockID.Hash == "" {
		return nil, nil
	}
	height, err := strconv.ParseUint(result.Block.Header.Height, 10, 64)
	if err != nil {
		return nil, err
	}
	return &types.BlockIdentifier{
		Index: int64(height),
		Hash:  result.BlockID.Hash,
	}, nil
}

func toTransactions(txs []sdk.TxResponse) (transactions []*types.Transaction, err error) {
	for _, tx := range txs {
		transactions = append(transactions, cosmosTxToRosettaTx(tx))
	}
	return
}

// tendermintTxToRosettaTx converts a Tendermint api TxResponseResult to a Transaction
// in the type expected by Rosetta.
func tendermintTxToRosettaTx(res tendermint.TxResponse) *types.Transaction {
	return &types.Transaction{
		TransactionIdentifier: &types.TransactionIdentifier{
			Hash: res.Hash,
		},
		Operations: nil, // TODO difficult to get the operations from the mempool (maybe not worth it due to block times).
	}
}

// SdkTxToOperations converts an sdk.Tx to rosetta operations
func SdkTxToOperations(tx auth.StdTx, withStatus, hasError bool) []*types.Operation {
	var operations []*types.Operation

	feeCoins := tx.Fee.Amount
	var feeOps = rosettaFeeOperationsFromCoins(feeCoins, tx.FeePayer().String(), withStatus)
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
		status = StatusSuccess
	}

	for i, coin := range coins {
		op := &types.Operation{
			OperationIdentifier: &types.OperationIdentifier{
				Index: int64(i),
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

func OperationsToSdkMsg(ops []*types.Operation) ([]sdk.Msg, string, sdk.Coins, error) {
	var feeAmnt []*types.Amount
	var newOps []*types.Operation
	if len(ops)%2 == 0 {
		msgs, signAddr, err := ConvertOpsToMsgs(ops)
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
	msgs, signAddr, err := ConvertOpsToMsgs(ops)
	if err != nil {
		return nil, "", nil, err
	}

	return msgs, signAddr, AmountsToCoins(feeAmnt), nil
}

func ConvertOpsToMsgs(ops []*types.Operation) ([]sdk.Msg, string, error) {
	var msgs []sdk.Msg
	var signAddr string
	var sendOps []*types.Operation
	var delOps []*types.Operation
	for _, op := range ops {
		switch op.Type {
		case OperationMsgSend:
			sendOps = append(sendOps, op)
		case OperationDelegate:
			delOps = append(delOps, op)
		}
	}
	if len(sendOps) == 2 {
		sendMsg, err := OperationsToSdkBankMsgSend(sendOps)
		if err != nil {
			return nil, "", err
		}
		msgs = append(msgs, sendMsg)
		signAddr = sendMsg.FromAddress.String()
	}

	if len(delOps) == 2 {
		delMsg, err := OperationsToSdkStakingMsgDelegate(delOps)
		if err != nil {
			return nil, "", err
		}
		msgs = append(msgs, delMsg)
		signAddr = delMsg.DelegatorAddress.String()
	}

	return msgs, signAddr, nil
}

// AmountsToCoins converts rosetta amounts to sdk coins
func AmountsToCoins(amounts []*types.Amount) sdk.Coins {
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

// OperationsToSdkBankMsgSend extracts the from and to addresses from a list of operations.
// We assume that it comes formated in the correct way. And that the balance of the sender is the same
// as the receiver operations.
func OperationsToSdkBankMsgSend(ops []*types.Operation) (bank.MsgSend, error) {
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

	return bank.NewMsgSend(from, to, sdk.NewCoins(sendAmt)), nil
}

// cosmosTxToRosettaTx converts a Cosmos api TxQuery to a Transaction
// in the type expected by Rosetta.
func cosmosTxToRosettaTx(tx sdk.TxResponse) *types.Transaction {
	hasError := tx.Code > 0
	return &types.Transaction{
		TransactionIdentifier: &types.TransactionIdentifier{
			Hash: tx.TxHash,
		},
		Operations: sdkMsgsToRosettaOperations(tx.Tx.GetMsgs(), true, hasError, 0),
	}
}

// sdkMsgsToRosettaOperations converts sdk messages to rosetta operations
func sdkMsgsToRosettaOperations(msgs []sdk.Msg, withStatus bool, hasError bool, feeLen int) []*types.Operation {
	var operations []*types.Operation
	var status string
	for i, msg := range msgs {
		switch msg.Type() { // nolint
		case OperationMsgSend:
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
					status = StatusSuccess
					if hasError {
						status = StatusReverted
					}
				}
				return &types.Operation{
					OperationIdentifier: &types.OperationIdentifier{
						Index: int64(index),
					},
					Type:   OperationMsgSend,
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
		case OperationDelegate:
			newMsg := msg.(*stakingtypes.MsgDelegate)
			delAddr := newMsg.DelegatorAddress
			valAddr := newMsg.ValidatorAddress
			coin := newMsg.Amount
			delOp := func(account, amount string, index int) *types.Operation {
				if withStatus {
					status = StatusSuccess
					if hasError {
						status = StatusReverted
					}
				}
				return &types.Operation{
					OperationIdentifier: &types.OperationIdentifier{
						Index: int64(index),
					},
					Type:   OperationDelegate,
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

func OperationsToSdkStakingMsgDelegate(ops []*types.Operation) (stakingtypes.MsgDelegate, error) {
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

// getTransferTxDataFromOperations extracts the from and to addresses from a list of operations.
// We assume that it comes formated in the correct way. And that the balance of the sender is the same
// as the receiver operations.
func getTransferTxDataFromOperations(ops []*types.Operation) (*TransferTxData, error) {
	var (
		transferData = &TransferTxData{}
		err          error
	)

	for _, op := range ops {
		if strings.HasPrefix(op.Amount.Value, "-") {
			transferData.From, err = sdk.AccAddressFromBech32(op.Account.Address)
			if err != nil {
				return nil, err
			}
		} else {
			transferData.To, err = sdk.AccAddressFromBech32(op.Account.Address)
			if err != nil {
				return nil, err
			}

			amount, err := strconv.ParseInt(op.Amount.Value, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid amount")
			}

			transferData.Amount = sdk.NewCoin(op.Amount.Currency.Symbol, sdk.NewInt(amount))
		}
	}

	return transferData, nil
}
