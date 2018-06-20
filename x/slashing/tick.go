package slashing

import (
	"encoding/binary"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/abci/types"
	tmtypes "github.com/tendermint/tendermint/types"
)

// slashing begin block functionality
func BeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock, sk Keeper) (tags sdk.Tags) {
	// Tag the height
	heightBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(heightBytes, uint64(req.Header.Height))
	tags = sdk.NewTags("height", heightBytes)

	// Deal with any equivocation evidence
	for _, evidence := range req.ByzantineValidators {
		pk, err := tmtypes.PB2TM.PubKey(evidence.Validator.PubKey)
		if err != nil {
			panic(err)
		}
		switch string(evidence.Type) {
		case tmtypes.ABCIEvidenceTypeDuplicateVote:
			sk.handleDoubleSign(ctx, evidence.Height, evidence.Time, pk)
		default:
			ctx.Logger().With("module", "x/slashing").Error(fmt.Sprintf("ignored unknown evidence type: %s", string(evidence.Type)))
		}
	}

	// Iterate over all the validators  which *should* have signed this block
	for _, validator := range req.Validators {
		present := validator.SignedLastBlock
		pubkey, err := tmtypes.PB2TM.PubKey(validator.Validator.PubKey)
		if err != nil {
			panic(err)
		}
		sk.handleValidatorSignature(ctx, pubkey, present)
	}

	return
}
