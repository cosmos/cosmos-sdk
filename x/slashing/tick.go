package slashing

import (
	"encoding/binary"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/abci/types"
	crypto "github.com/tendermint/go-crypto"
	tmtypes "github.com/tendermint/tendermint/types"
)

func NewBeginBlocker(sk Keeper) sdk.BeginBlocker {
	return func(ctx sdk.Context, req abci.RequestBeginBlock) abci.ResponseBeginBlock {
		// Tag the height
		heightBytes := make([]byte, 8)
		binary.LittleEndian.PutUint64(heightBytes, uint64(req.Header.Height))
		tags := sdk.NewTags("height", heightBytes)

		// Deal with any equivocation evidence
		for _, evidence := range req.ByzantineValidators {
			var pk crypto.PubKey
			sk.cdc.MustUnmarshalBinary(evidence.PubKey, &pk)
			switch string(evidence.Type) {
			case tmtypes.DUPLICATE_VOTE:
				sk.handleDoubleSign(ctx, evidence.Height, evidence.Time, pk)
			default:
				ctx.Logger().With("module", "x/slashing").Error(fmt.Sprintf("Ignored unknown evidence type: %s", string(evidence.Type)))
			}
		}

		// Figure out which validators were absent
		absent := make(map[crypto.PubKey]struct{})
		for _, pubkey := range req.AbsentValidators {
			var pk crypto.PubKey
			sk.cdc.MustUnmarshalBinary(pubkey, &pk)
			absent[pk] = struct{}{}
		}

		// Iterate over all the validators which *should* have signed this block
		sk.stakeKeeper.IterateValidatorsBonded(ctx, func(_ int64, validator sdk.Validator) (stop bool) {
			pubkey := validator.GetPubKey()
			present := true
			if _, ok := absent[pubkey]; ok {
				present = false
			}
			sk.handleValidatorSignature(ctx, pubkey, present)
			return false
		})

		// Return the begin block response
		// TODO Return something composable, so other modules can also have BeginBlockers
		// TODO Add some more tags so clients can track slashing events
		return abci.ResponseBeginBlock{
			Tags: tags.ToKVPairs(),
		}
	}
}
