package conversion

import (
	"fmt"
	"time"

	"github.com/coinbase/rosetta-sdk-go/types"
	tmcoretypes "github.com/tendermint/tendermint/rpc/core/types"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/server/rosetta"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

// TimeToMilliseconds converts time to milliseconds timestamp
func TimeToMilliseconds(t time.Time) int64 {
	return t.UnixNano() / (int64(time.Millisecond) / int64(time.Nanosecond))
}

// CoinsToBalance converts sdk.Coins to rosetta.Amounts
func CoinsToBalance(coins []sdk.Coin) []*types.Amount {
	amounts := make([]*types.Amount, len(coins))

	for i, coin := range coins {
		amounts[i] = &types.Amount{
			Value: coin.Amount.String(),
			Currency: &types.Currency{
				Symbol: coin.Denom,
			},
		}
	}

	return amounts
}

// ResultTxSearchToTransaction converts tendermint search transactions to rosetta ones
func ResultTxSearchToTransaction(txs []*rosetta.SdkTxWithHash) []*types.Transaction {
	converted := make([]*types.Transaction, len(txs))
	for i, tx := range txs {
		//hasError := tx.Code > 0
		converted[i] = &types.Transaction{
			TransactionIdentifier: &types.TransactionIdentifier{Hash: tx.HexHash},
			Operations:            SdkTxToOperations(tx.Tx, false, false),
			Metadata:              nil,
		}
	}
	return converted
}

// SdkTxResponseToOperations converts a tx response to operations
func SdkTxToOperations(tx sdk.Tx, hasError, withoutStatus bool) []*types.Operation {
	return toOperations(tx.GetMsgs(), hasError, withoutStatus)
}

// TendermintTxsToTxIdentifiers converts a tendermint raw transaction into a rosetta tx identifier
func TendermintTxsToTxIdentifiers(txs []tmtypes.Tx) []*types.TransactionIdentifier {
	converted := make([]*types.TransactionIdentifier, len(txs))
	for i, tx := range txs {
		converted[i] = &types.TransactionIdentifier{Hash: fmt.Sprintf("%x", tx.Hash())} // TODO hash is sha256, so we hex it?
	}
	return converted
}

// TendermintBlockToBlockIdentifier converts a tendermint result block to a rosetta block identifier
func TendermintBlockToBlockIdentifier(block *tmcoretypes.ResultBlock) *types.BlockIdentifier {
	return &types.BlockIdentifier{
		Index: block.Block.Height,
		Hash:  block.Block.Hash().String(),
	}
}

func toOperations(msgs []sdk.Msg, hasError bool, withoutStatus bool) []*types.Operation {
	var operations []*types.Operation
	for i, msg := range msgs {
		switch msg.Type() { // nolint
		case "send":
			newMsg := msg.(*banktypes.MsgSend)
			fromAddress := newMsg.FromAddress
			toAddress := newMsg.ToAddress
			amounts := newMsg.Amount
			if len(amounts) == 0 {
				continue
			}
			coin := amounts[0]
			sendOp := func(account, amount string, index int) *types.Operation {
				status := rosetta.StatusSuccess
				if hasError {
					status = rosetta.StatusReverted
				}
				if withoutStatus {
					status = ""
				}
				return &types.Operation{
					OperationIdentifier: &types.OperationIdentifier{
						Index: int64(index),
					},
					Type:   rosetta.OperationMsgSend,
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
				sendOp(fromAddress, "-"+coin.Amount.String(), i),
				sendOp(toAddress, coin.Amount.String(), i+1),
			)
		}
	}
	return operations
}

// TmPeersToRosettaPeers converts tendermint peers to rosetta ones
func TmPeersToRosettaPeers(peers []tmcoretypes.Peer) []*types.Peer {
	converted := make([]*types.Peer, len(peers))
	for i, peer := range peers {
		converted[i] = &types.Peer{
			PeerID: peer.NodeInfo.Moniker,
			Metadata: map[string]interface{}{
				"addr": peer.NodeInfo.ListenAddr,
			},
		}
	}
	return converted
}

// TendermintStatusToSync converts a tendermint status to rosetta sync status
func TendermintStatusToSync(status *tmcoretypes.ResultStatus) *types.SyncStatus {
	// determine sync status
	var stage = rosetta.StageSynced
	if status.SyncInfo.CatchingUp {
		stage = rosetta.StageSyncing
	}

	return &types.SyncStatus{
		CurrentIndex: status.SyncInfo.LatestBlockHeight,
		TargetIndex:  nil, // sync info does not allow us to get target height
		Stage:        &stage,
	}
}

// ParentBlockIdentifierFromLastBlock returns the parent block identifier from the last block
func ParentBlockIdentifierFromLastBlock(block *tmcoretypes.ResultBlock) *types.BlockIdentifier {
	if block.Block.Height == 1 {
		return &types.BlockIdentifier{
			Index: 1,
			Hash:  fmt.Sprintf("%X", block.BlockID.Hash.Bytes()),
		}
	}
	return &types.BlockIdentifier{
		Index: block.Block.Height - 1,
		Hash:  fmt.Sprintf("%X", block.Block.LastBlockID.Hash.Bytes()),
	}
}
