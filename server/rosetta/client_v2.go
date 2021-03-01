package rosetta

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"

	rosettatypes "github.com/coinbase/rosetta-sdk-go/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/gogo/protobuf/proto"
	crgerrs "github.com/tendermint/cosmos-rosetta-gateway/errors"
	abci "github.com/tendermint/tendermint/abci/types"
	tmtypes "github.com/tendermint/tendermint/types"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

const mockBeginEndBlockTxLength = sha256.Size + 1
const endBlockHashStart = 0x0
const beginBlockHashStart = 0x1
const burnerAddressIdentifier = "burner"

func (c *Client) blockTxs(ctx context.Context, height *int64) ([]*rosettatypes.Transaction, error) {
	// get block info
	blockInfo, err := c.clientCtx.Client.Block(ctx, height)
	if err != nil {
		return nil, err
	}
	// get block events
	blockResults, err := c.clientCtx.Client.BlockResults(ctx, height)
	if err != nil {
		return nil, err
	}

	if len(blockResults.TxsResults) != len(blockInfo.Block.Txs) {
		// wtf?
		panic("block results transactions do now match block transactions")
	}
	// process begin and end block txs
	beginBlockTx := &rosettatypes.Transaction{
		TransactionIdentifier: &rosettatypes.TransactionIdentifier{Hash: beginBlockTxHash(blockInfo.BlockID.Hash)},
		Operations: normalizeOperationIndexes(
			nil,
			eventsToBalanceOperations(StatusSuccess, blockResults.BeginBlockEvents),
		),
	}
	endBlockTx := &rosettatypes.Transaction{
		TransactionIdentifier: &rosettatypes.TransactionIdentifier{Hash: endBlockTxHash(blockInfo.BlockID.Hash)},
		Operations: normalizeOperationIndexes(
			nil,
			eventsToBalanceOperations(StatusSuccess, blockResults.EndBlockEvents),
		),
	}

	deliverTx := make([]*rosettatypes.Transaction, len(blockInfo.Block.Txs))
	// process normal txs
	for i, tx := range blockInfo.Block.Txs {
		rosTx, err := toRosTx(c.clientCtx.TxConfig.TxDecoder(), tx, blockResults.TxsResults[i])
		if err != nil {
			return nil, err
		}
		deliverTx[i] = rosTx
	}

	finalTxs := make([]*rosettatypes.Transaction, 0, 2+len(deliverTx))
	finalTxs = append(finalTxs, beginBlockTx)
	finalTxs = append(finalTxs, deliverTx...)
	finalTxs = append(finalTxs, endBlockTx)

	return finalTxs, nil
}

func endBlockTxHash(hash []byte) string {
	return fmt.Sprintf("%x%x", endBlockHashStart, hash)
}

func beginBlockTxHash(hash []byte) string {
	return fmt.Sprintf("%x%x", beginBlockHashStart, hash)
}

func eventsToBalanceOperations(status string, events []abci.Event) []*rosettatypes.Operation {
	var ops []*rosettatypes.Operation

	for _, e := range events {
		balanceOps, ok := eventToBalanceOperations(status, e)
		if !ok {
			continue
		}
		ops = append(ops, balanceOps...)
	}

	return ops
}

// eventToBalanceOperations converts an event to a rosetta balance operation
// it will panic if the event is malformed because it might mean the sdk spec
// has changed and rosetta needs to reflect those changes too
func eventToBalanceOperations(status string, event abci.Event) (operations []*rosettatypes.Operation, isBalanceEvent bool) {

	var (
		accountIdentifier string
		coinChange        sdk.Coins
		isSub             bool
	)

	switch event.Type {
	default:
		return nil, false
	case banktypes.EventTypeCoinSpent:
		spender, err := sdk.AccAddressFromBech32((string)(event.Attributes[0].Value))
		if err != nil {
			panic(err)
		}
		coins, err := sdk.ParseCoinsNormalized((string)(event.Attributes[1].Value))
		if err != nil {
			panic(err)
		}

		isSub = true
		coinChange = coins
		accountIdentifier = spender.String()

	case banktypes.EventTypeCoinReceived:
		receiver, err := sdk.AccAddressFromBech32((string)(event.Attributes[0].Value))
		if err != nil {
			panic(err)
		}
		coins, err := sdk.ParseCoinsNormalized((string)(event.Attributes[1].Value))
		if err != nil {
			panic(err)
		}

		isSub = false
		coinChange = coins
		accountIdentifier = receiver.String()
	}

	operations = make([]*rosettatypes.Operation, len(coinChange))

	for i, coin := range coinChange {

		value := coin.Amount.String()
		// in case the event is a subtract balance one the rewrite value with
		// the negative coin identifier
		if isSub {
			value = "-" + value
		}

		op := &rosettatypes.Operation{
			Type:    event.Type,
			Status:  status,
			Account: &rosettatypes.AccountIdentifier{Address: accountIdentifier},
			Amount: &rosettatypes.Amount{
				Value: value,
				Currency: &rosettatypes.Currency{
					Symbol:   coin.Denom,
					Decimals: 0,
				},
			},
		}

		operations[i] = op
	}
	return operations, true
}

func toRosTx(txDecoder sdk.TxDecoder, rawTx tmtypes.Tx, txResult *abci.ResponseDeliverTx) (*rosettatypes.Transaction, error) {
	// decode tx
	tx, err := txDecoder(rawTx)
	if err != nil {
		return nil, crgerrs.WrapError(crgerrs.ErrCodec, err.Error())
	}
	// get initial status, as per sdk design, if one msg fails
	// the whole TX will be considered failing, so we can't have
	// 1 msg being success and 1 msg being reverted
	status := StatusSuccess
	if txResult.Code != abci.CodeTypeOK {
		status = StatusReverted
	}
	// get operations from msgs
	msgs := tx.GetMsgs()
	var rawTxOps []*rosettatypes.Operation
	for _, msg := range msgs {
		ops, err := opsFromMsg(status, msg)
		if err != nil {
			return nil, err
		}
		rawTxOps = append(rawTxOps, ops...)
	}

	// now get balance events from response deliver tx
	balanceOps := eventsToBalanceOperations(status, txResult.Events)

	// now normalize indexes
	totalOps := normalizeOperationIndexes(rawTxOps, balanceOps)

	return &rosettatypes.Transaction{
		TransactionIdentifier: &rosettatypes.TransactionIdentifier{Hash: fmt.Sprintf("%x", rawTx.Hash())},
		Operations:            totalOps,
	}, nil
}

// opsFromMsg will create an operation for each msg signer
// with the message proto name as type, and the raw fields
// as metadata
func opsFromMsg(status string, msg sdk.Msg) ([]*rosettatypes.Operation, error) {
	opName := proto.MessageName(msg)
	// in case proto does not recognize the message name
	// then we should try to cast it to service msg, to
	// check if it was wrapped or not, in case the cast
	// from sdk.ServiceMsg to sdk.Msg fails, then a
	// codec error is returned
	if opName == "" {
		unwrappedMsg, ok := msg.(sdk.ServiceMsg)
		if !ok {
			return nil, crgerrs.WrapError(crgerrs.ErrCodec, fmt.Sprintf("unrecognized message type: %T", msg))
		}

		msg, ok = unwrappedMsg.Request.(sdk.Msg)
		if !ok {
			return nil, crgerrs.WrapError(
				crgerrs.ErrCodec,
				fmt.Sprintf("unable to cast %T to sdk.Msg, method: %s", unwrappedMsg, unwrappedMsg.MethodName),
			)
		}

		opName = proto.MessageName(msg)
		if opName == "" {
			return nil, crgerrs.WrapError(crgerrs.ErrCodec, fmt.Sprintf("unrecognized message type: %T", msg))
		}
	}

	meta, err := msgToMetadata(msg)
	if err != nil {
		return nil, err
	}

	ops := make([]*rosettatypes.Operation, len(msg.GetSigners()))
	for i, signer := range msg.GetSigners() {
		op := &rosettatypes.Operation{
			Type:     opName,
			Status:   status,
			Account:  &rosettatypes.AccountIdentifier{Address: signer.String()},
			Metadata: meta,
		}

		ops[i] = op
	}

	return ops, nil
}

// msgToMetadata converts an sdk.Msg to map[string]interface{}
func msgToMetadata(msg sdk.Msg) (map[string]interface{}, error) {
	b, err := json.Marshal(msg)
	if err != nil {
		return nil, crgerrs.WrapError(crgerrs.ErrCodec, err.Error())
	}
	m := make(map[string]interface{})
	err = json.Unmarshal(b, &m)
	if err != nil {
		return nil, crgerrs.WrapError(crgerrs.ErrCodec, err.Error())
	}
	return m, nil
}

// normalizeOperationIndexes adds the indexes to operations adhering to specific rules:
// operations related to messages will be always before than the balance ones
func normalizeOperationIndexes(msgOps []*rosettatypes.Operation, balanceOps []*rosettatypes.Operation) (finalOps []*rosettatypes.Operation) {
	lenMsgOps := len(msgOps)
	lenBalanceOps := len(balanceOps)
	finalOps = make([]*rosettatypes.Operation, 0, lenMsgOps+lenBalanceOps)

	var currentIndex int64
	// add indexes to msg ops
	for _, op := range msgOps {
		op.OperationIdentifier = &rosettatypes.OperationIdentifier{
			Index: currentIndex,
		}

		finalOps = append(finalOps, op)
		currentIndex++
	}

	// add indexes to balance ops
	for _, op := range balanceOps {
		op.OperationIdentifier = &rosettatypes.OperationIdentifier{
			Index: currentIndex,
		}

		finalOps = append(finalOps, op)
		currentIndex++
	}

	return finalOps
}
