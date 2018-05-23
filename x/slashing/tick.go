package slashing

import (
	"bytes"
	"encoding/binary"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/abci/types"
	crypto "github.com/tendermint/go-crypto"
)

func NewBeginBlocker(sk Keeper) sdk.BeginBlocker {
	return func(ctx sdk.Context, req abci.RequestBeginBlock) abci.ResponseBeginBlock {
		heightBytes := make([]byte, 8)
		binary.LittleEndian.PutUint64(heightBytes, uint64(req.Header.Height))
		tags := sdk.NewTags("height", heightBytes)
		for _, evidence := range req.ByzantineValidators {
			var pk crypto.PubKey
			sk.cdc.MustUnmarshalBinary(evidence.PubKey, &pk)
			switch {
			case bytes.Compare(evidence.Type, []byte("doubleSign")) == 0:
				sk.handleDoubleSign(ctx, evidence.Height, evidence.Time, pk)
			default:
				ctx.Logger().With("module", "x/slashing").Error(fmt.Sprintf("Ignored unknown evidence type: %s", string(evidence.Type)))
			}
		}
		absent := make(map[string]bool)
		for _, pubkey := range req.AbsentValidators {
			var pk crypto.PubKey
			sk.cdc.MustUnmarshalBinary(pubkey, &pk)
			absent[string(pk.Bytes())] = true
		}
		sk.stakeKeeper.IterateValidatorsBonded(ctx, func(_ int64, validator sdk.Validator) (stop bool) {
			pubkey := validator.GetPubKey()
			sk.handleValidatorSignature(ctx, pubkey, !absent[string(pubkey.Bytes())])
			return false
		})
		// TODO Add some more tags so clients can track slashing
		return abci.ResponseBeginBlock{
			Tags: tags.ToKVPairs(),
		}
	}
}
