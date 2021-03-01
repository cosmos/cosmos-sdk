package rosetta

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/gogo/protobuf/proto"
	crgerrs "github.com/tendermint/cosmos-rosetta-gateway/errors"
	crgtypes "github.com/tendermint/cosmos-rosetta-gateway/types"
	abci "github.com/tendermint/tendermint/abci/types"
	tmcoretypes "github.com/tendermint/tendermint/rpc/core/types"
	tmtypes "github.com/tendermint/tendermint/types"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/gogo/protobuf/jsonpb"

	rosettatypes "github.com/coinbase/rosetta-sdk-go/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ----------------------- sdk to rosetta ------------------------

// sdkTxToRosettaTx converts a tendermint raw transaction and its result (if provided) to a rosetta transaction
// using the sdk transaction decoder
func sdkTxToRosettaTx(txDecoder sdk.TxDecoder, rawTx tmtypes.Tx, txResult *abci.ResponseDeliverTx) (
	*rosettatypes.Transaction,
	error) {
	// decode tx
	tx, err := txDecoder(rawTx)
	if err != nil {
		return nil, crgerrs.WrapError(crgerrs.ErrCodec, err.Error())
	}
	// get initial status, as per sdk design, if one msg fails
	// the whole TX will be considered failing, so we can't have
	// 1 msg being success and 1 msg being reverted
	status := StatusTxSuccess
	switch txResult {
	// if nil, we're probably checking an unconfirmed tx
	case nil:
		status = StatusTxUnconfirmed
	// set the status
	default:
		if txResult.Code != abci.CodeTypeOK {
			status = StatusTxReverted
		}
	}
	// get operations from msgs
	msgs := tx.GetMsgs()
	var rawTxOps []*rosettatypes.Operation
	for _, msg := range msgs {
		ops, err := sdkMsgToRosettaOperation(status, msg)
		if err != nil {
			return nil, err
		}
		rawTxOps = append(rawTxOps, ops...)
	}

	// now get balance events from response deliver tx
	var balanceOps []*rosettatypes.Operation
	// tx result might be nil, in case we're querying an unconfirmed tx from the mempool
	if txResult != nil {
		balanceOps = sdkEventsToBalanceOperations(status, txResult.Events)
	}

	// now normalize indexes
	totalOps := normalizeOperationIndexes(rawTxOps, balanceOps)

	return &rosettatypes.Transaction{
		TransactionIdentifier: &rosettatypes.TransactionIdentifier{Hash: fmt.Sprintf("%x", rawTx.Hash())},
		Operations:            totalOps,
	}, nil
}

// sdkEventsToBalanceOperations takes a slice of tendermint events and converts them to
// balance operations
func sdkEventsToBalanceOperations(status string, events []abci.Event) []*rosettatypes.Operation {
	var ops []*rosettatypes.Operation

	for _, e := range events {
		balanceOps, ok := sdkEventToBalanceOperations(status, e)
		if !ok {
			continue
		}
		ops = append(ops, balanceOps...)
	}

	return ops
}

// sdkEventToBalanceOperations converts an event to a rosetta balance operation
// it will panic if the event is malformed because it might mean the sdk spec
// has changed and rosetta needs to reflect those changes too.
// The balance operations are multiple, one for each denom.
func sdkEventToBalanceOperations(status string, event abci.Event) (operations []*rosettatypes.Operation, isBalanceEvent bool) {

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

	// rosetta does not have the concept of burning coins, so we need to mock
	// the burn as a send to an address that cannot be resolved to anything
	case banktypes.EventTypeCoinBurn:
		coins, err := sdk.ParseCoinsNormalized((string)(event.Attributes[1].Value))
		if err != nil {
			panic(err)
		}

		coinChange = coins
		accountIdentifier = burnerAddressIdentifier
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

// sdkMsgToRosettaOperation will create an operation for each msg signer
// with the message proto name as type, and the raw fields
// as metadata
func sdkMsgToRosettaOperation(status string, msg sdk.Msg) ([]*rosettatypes.Operation, error) {
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

	meta, err := sdkMsgToRosettaMetadata(msg)
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

// sdkMsgToRosettaMetadata converts an sdk.Msg to map[string]interface{}
func sdkMsgToRosettaMetadata(msg sdk.Msg) (map[string]interface{}, error) {
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

// ---------------------------- end ------------------------------

// opsToMsgsAndFees converts rosetta operations to sdk.Msg and fees represented as sdk.Coins
func opsToMsgsAndFees(interfaceRegistry jsonpb.AnyResolver, ops []*rosettatypes.Operation) ([]sdk.Msg, sdk.Coins, error) {
	var feeAmnt []*rosettatypes.Amount
	var newOps []*rosettatypes.Operation
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
func amountsToCoins(amounts []*rosettatypes.Amount) sdk.Coins {
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

func opsToMsgs(interfaceRegistry jsonpb.AnyResolver, ops []*rosettatypes.Operation) ([]sdk.Msg, error) {
	var msgs []sdk.Msg
	var operationsByType = make(map[string][]*rosettatypes.Operation)
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

// ------------------- from tendermint to rosetta ------------------

// tmTxsToRosettaTxsIdentifiers converts a tendermint raw transactions into an array of rosetta tx identifiers
func tmTxsToRosettaTxsIdentifiers(txs []tmtypes.Tx) []*rosettatypes.TransactionIdentifier {
	converted := make([]*rosettatypes.TransactionIdentifier, len(txs))
	for i, tx := range txs {
		converted[i] = tmTxToRosettaTxIdentifier(tx)
	}

	return converted
}

// tmTxToRosettaTxIdentifier converts a tendermint raw transaction into a rosetta tx identifier
func tmTxToRosettaTxIdentifier(tx tmtypes.Tx) *rosettatypes.TransactionIdentifier {
	return &rosettatypes.TransactionIdentifier{Hash: fmt.Sprintf("%x", tx.Hash())}
}

// tmPeersToRosettaPeers converts tendermint peers to rosetta ones
func tmPeersToRosettaPeers(peers []tmcoretypes.Peer) []*rosettatypes.Peer {
	converted := make([]*rosettatypes.Peer, len(peers))

	for i, peer := range peers {
		converted[i] = &rosettatypes.Peer{
			PeerID: peer.NodeInfo.Moniker,
			Metadata: map[string]interface{}{
				"addr": peer.NodeInfo.ListenAddr,
			},
		}
	}

	return converted
}

// tmStatusToRosettaSyncStatus converts a tendermint status to rosetta sync status
func tmStatusToRosettaSyncStatus(status *tmcoretypes.ResultStatus) *rosettatypes.SyncStatus {
	// determine sync status
	var stage = StatusPeerSynced
	if status.SyncInfo.CatchingUp {
		stage = StatusPeerSyncing
	}

	return &rosettatypes.SyncStatus{
		CurrentIndex: status.SyncInfo.LatestBlockHeight,
		TargetIndex:  nil, // sync info does not allow us to get target height
		Stage:        &stage,
	}
}

// tmBlockToRosettaBlockIdentifier converts a tendermint result block to a rosetta block identifier
func tmBlockToRosettaBlockIdentifier(block *tmcoretypes.ResultBlock) *rosettatypes.BlockIdentifier {
	return &rosettatypes.BlockIdentifier{
		Index: block.Block.Height,
		Hash:  block.Block.Hash().String(),
	}
}

// tmBlockToRosettaParentBlockIdentifier returns the parent block identifier from the last block
func tmBlockToRosettaParentBlockIdentifier(block *tmcoretypes.ResultBlock) *rosettatypes.BlockIdentifier {
	if block.Block.Height == 1 {
		return &rosettatypes.BlockIdentifier{
			Index: 1,
			Hash:  fmt.Sprintf("%X", block.BlockID.Hash.Bytes()),
		}
	}

	return &rosettatypes.BlockIdentifier{
		Index: block.Block.Height - 1,
		Hash:  fmt.Sprintf("%X", block.Block.LastBlockID.Hash.Bytes()),
	}
}

// tmResultBlockToRosettaBlockResponse converts a tendermint result block to block response
func tmResultBlockToRosettaBlockResponse(block *tmcoretypes.ResultBlock) crgtypes.BlockResponse {
	return crgtypes.BlockResponse{
		Block:                tmBlockToRosettaBlockIdentifier(block),
		ParentBlock:          tmBlockToRosettaParentBlockIdentifier(block),
		MillisecondTimestamp: timeToMilliseconds(block.Block.Time),
		TxCount:              int64(len(block.Block.Txs)),
	}
}

// --------------------------- end ---------------------------------

// ----------------------- raw ros utils ---------------------------

// endBlockTxHash produces a mock endblock hash that rosetta can query
// for endblock operations, it also serves the purpose of representing
// part of the state changes happening at endblock level (balance ones)
func endBlockTxHash(hash []byte) string {
	return fmt.Sprintf("%x%x", endBlockHashStart, hash)
}

// beginBlockTxHash produces a mock beginblock hash that rosetta can query
// for beginblock operations, it also serves the purpose of representing
// part of the state changes happening at beginblock level (balance ones)
func beginBlockTxHash(hash []byte) string {
	return fmt.Sprintf("%x%x", beginBlockHashStart, hash)
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

// --------------------------- end ---------------------------------
