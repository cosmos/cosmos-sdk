package oracle

import (
	"bytes"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Handler handles payload after it passes voting process
type Handler func(ctx sdk.Context, p Payload) sdk.Error

func (keeper Keeper) update(ctx sdk.Context, val sdk.Validator, valset sdk.ValidatorSet, p Payload, info Info) Info {
	info.Power = info.Power.Add(val.GetPower())

	// Return if the voted power is not bigger than required power
	totalPower := valset.TotalPower(ctx)
	requiredPower := totalPower.Mul(keeper.supermaj)
	if !info.Power.GT(requiredPower) {
		return info
	}

	// Check if the validators hash has been changed during the vote process
	// and recalculate voted power
	hash := ctx.BlockHeader().ValidatorsHash
	if !bytes.Equal(hash, info.Hash) {
		info.Power = sdk.ZeroRat()
		info.Hash = hash
		prefix := GetSignPrefix(p, keeper.cdc)
		store := ctx.KVStore(keeper.key)
		iter := sdk.KVStorePrefixIterator(store, prefix)
		for ; iter.Valid(); iter.Next() {
			if valset.Validator(ctx, iter.Value()) != nil {
				store.Delete(iter.Key())
				continue
			}
			info.Power = info.Power.Add(val.GetPower())
		}
		if !info.Power.GT(totalPower.Mul(keeper.supermaj)) {
			return info
		}
	}

	info.Status = Processed
	return info
}

// Handle is used by other modules to handle Msg
func (keeper Keeper) Handle(h Handler, ctx sdk.Context, o Msg, codespace sdk.CodespaceType) sdk.Result {
	valset := keeper.valset

	signer := o.Signer
	payload := o.Payload

	// Check the oracle is not in process
	info := keeper.Info(ctx, payload)
	if info.Status != Pending {
		return ErrAlreadyProcessed(codespace).Result()
	}

	// Check if it is reporting timeout
	now := ctx.BlockHeight()
	if now > info.LastSigned+keeper.timeout {
		info = Info{Status: Timeout}
		keeper.setInfo(ctx, payload, info)
		keeper.clearSigns(ctx, payload)
		return sdk.Result{}
	}
	info.LastSigned = ctx.BlockHeight()

	// Check the signer is a validater
	val := valset.Validator(ctx, signer)
	if val == nil {
		return ErrNotValidator(codespace, signer).Result()
	}

	// Check double signing
	if keeper.signed(ctx, payload, signer) {
		return ErrAlreadySigned(codespace).Result()
	}

	keeper.sign(ctx, payload, signer)

	info = keeper.update(ctx, val, valset, payload, info)
	if info.Status == Processed {
		info = Info{Status: Processed}
	}

	keeper.setInfo(ctx, payload, info)

	if info.Status == Processed {
		keeper.clearSigns(ctx, payload)
		cctx, write := ctx.CacheContext()
		err := h(cctx, payload)
		if err != nil {
			return sdk.Result{
				Code: sdk.ABCICodeOK,
				Log:  err.ABCILog(),
			}
		}
		write()

	}

	return sdk.Result{}
}
