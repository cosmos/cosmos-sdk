package conversion

import (
	"fmt"
	"time"

	"github.com/coinbase/rosetta-sdk-go/types"
	tmcoretypes "github.com/tendermint/tendermint/rpc/core/types"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/server/rosetta"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TimeToMilliseconds converts time to milliseconds timestamp
func TimeToMilliseconds(t time.Time) int64 {
	return t.UnixNano() / (int64(time.Millisecond) / int64(time.Nanosecond))
}

// SdkCoinsToRosettaAmounts converts []sdk.Coin to rosetta amounts
// availableCoins keeps track of current available coins vs the coins
// owned by an address. This is required to support historical balances
// as rosetta expects them to be set to 0, if an address does not own them
func SdkCoinsToRosettaAmounts(ownedCoins []sdk.Coin, availableCoins sdk.Coins) []*types.Amount {
	amounts := make([]*types.Amount, len(availableCoins))
	ownedCoinsMap := make(map[string]sdk.Int, len(availableCoins))

	for _, ownedCoin := range ownedCoins {
		ownedCoinsMap[ownedCoin.Denom] = ownedCoin.Amount
	}

	for i, coin := range availableCoins {
		value, owned := ownedCoinsMap[coin.Denom]
		if !owned {
			amounts[i] = &types.Amount{
				Value: sdk.NewInt(0).String(),
				Currency: &types.Currency{
					Symbol: coin.Denom,
				},
			}
			continue
		}
		amounts[i] = &types.Amount{
			Value: value.String(),
			Currency: &types.Currency{
				Symbol: coin.Denom,
			},
		}
	}

	return amounts
}

// SdkTxsWithHashToRosettaTxs converts sdk transactions wrapped with their hash to rosetta transactions
func SdkTxsWithHashToRosettaTxs(txs []*rosetta.SdkTxWithHash) []*types.Transaction {
	converted := make([]*types.Transaction, len(txs))
	for i, tx := range txs {
		converted[i] = SdkTxWithHashToRosettaTx(tx)
	}

	return converted
}

func SdkTxWithHashToRosettaTx(tx *rosetta.SdkTxWithHash) *types.Transaction {
	hasError := tx.Code != 0
	return &types.Transaction{
		TransactionIdentifier: &types.TransactionIdentifier{Hash: tx.HexHash},
		Operations:            SdkTxToOperations(tx.Tx, true, hasError),
		Metadata: map[string]interface{}{
			rosetta.Log: tx.Log,
		},
	}
}

// SdkTxToOperations converts an sdk.Tx to rosetta operations
func SdkTxToOperations(tx sdk.Tx, withStatus, hasError bool) []*types.Operation {
	var operations []*types.Operation

	msgOps := sdkMsgsToRosettaOperations(tx.GetMsgs(), withStatus, hasError)
	operations = append(operations, msgOps...)

	feeTx := tx.(sdk.FeeTx)
	feeOps := sdkFeeTxToOperations(feeTx, withStatus, len(msgOps))
	operations = append(operations, feeOps...)

	return operations
}

// sdkFeeTxToOperations converts sdk.FeeTx to rosetta operations
func sdkFeeTxToOperations(feeTx sdk.FeeTx, withStatus bool, previousOps int) []*types.Operation {
	feeCoins := feeTx.GetFee()
	var ops []*types.Operation
	if feeCoins != nil {
		var feeOps = rosettaFeeOperationsFromCoins(feeCoins, feeTx.FeePayer().String(), withStatus, previousOps)
		ops = append(ops, feeOps...)
	}

	return ops
}

// rosettaFeeOperationsFromCoins returns the list of rosetta fee operations given sdk coins
func rosettaFeeOperationsFromCoins(coins sdk.Coins, account string, withStatus bool, previousOps int) []*types.Operation {
	feeOps := make([]*types.Operation, 0)
	var status string
	if withStatus {
		status = rosetta.StatusSuccess
	}

	for i, coin := range coins {
		op := &types.Operation{
			OperationIdentifier: &types.OperationIdentifier{
				Index: int64(previousOps + i),
			},
			Type:   rosetta.OperationFee,
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

// sdkMsgsToRosettaOperations converts sdk messages to rosetta operations
func sdkMsgsToRosettaOperations(msgs []sdk.Msg, withStatus bool, hasError bool) []*types.Operation {
	var operations []*types.Operation
	for _, msg := range msgs {
		if rosettaMsg, ok := msg.(rosetta.Msg); ok {
			operations = append(operations, rosettaMsg.ToOperations(withStatus, hasError)...)
		}
	}

	return operations
}

// TMTxsToRosettaTxsIdentifiers converts a tendermint raw transaction into a rosetta tx identifier
func TMTxsToRosettaTxsIdentifiers(txs []tmtypes.Tx) []*types.TransactionIdentifier {
	converted := make([]*types.TransactionIdentifier, len(txs))
	for i, tx := range txs {
		converted[i] = &types.TransactionIdentifier{Hash: fmt.Sprintf("%x", tx.Hash())}
	}

	return converted
}

// TMBlockToRosettaBlockIdentifier converts a tendermint result block to a rosetta block identifier
func TMBlockToRosettaBlockIdentifier(block *tmcoretypes.ResultBlock) *types.BlockIdentifier {
	return &types.BlockIdentifier{
		Index: block.Block.Height,
		Hash:  block.Block.Hash().String(),
	}
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

// TMStatusToRosettaSyncStatus converts a tendermint status to rosetta sync status
func TMStatusToRosettaSyncStatus(status *tmcoretypes.ResultStatus) *types.SyncStatus {
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

// TMBlockToRosettaParentBlockIdentifier returns the parent block identifier from the last block
func TMBlockToRosettaParentBlockIdentifier(block *tmcoretypes.ResultBlock) *types.BlockIdentifier {
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
