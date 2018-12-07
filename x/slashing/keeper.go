package slashing

import (
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	stake "github.com/cosmos/cosmos-sdk/x/stake/types"
	"github.com/tendermint/tendermint/crypto"
)

// Keeper of the slashing store
type Keeper struct {
	storeKey     sdk.StoreKey
	cdc          *codec.Codec
	validatorSet sdk.ValidatorSet
	paramspace   params.Subspace

	// codespace
	codespace sdk.CodespaceType
}

// NewKeeper creates a slashing keeper
func NewKeeper(cdc *codec.Codec, key sdk.StoreKey, vs sdk.ValidatorSet, paramspace params.Subspace, codespace sdk.CodespaceType) Keeper {
	keeper := Keeper{
		storeKey:     key,
		cdc:          cdc,
		validatorSet: vs,
		paramspace:   paramspace.WithTypeTable(ParamTypeTable()),
		codespace:    codespace,
	}
	return keeper
}

// handle a validator signing two blocks at the same height
// power: power of the double-signing validator at the height of infraction
func (k Keeper) handleDoubleSign(ctx sdk.Context, addr crypto.Address, infractionHeight int64, timestamp time.Time, power int64) {
	logger := ctx.Logger().With("module", "x/slashing")
	time := ctx.BlockHeader().Time
	age := time.Sub(timestamp)
	consAddr := sdk.ConsAddress(addr)
	pubkey, err := k.getPubkey(ctx, addr)
	if err != nil {
		panic(fmt.Sprintf("Validator consensus-address %v not found", consAddr))
	}

	// Get validator.
	validator := k.validatorSet.ValidatorByConsAddr(ctx, consAddr)
	if validator == nil || validator.GetStatus() == sdk.Unbonded {
		// Defensive.
		// Simulation doesn't take unbonding periods into account, and
		// Tendermint might break this assumption at some point.
		return
	}

	// Double sign too old
	maxEvidenceAge := k.MaxEvidenceAge(ctx)
	if age > maxEvidenceAge {
		logger.Info(fmt.Sprintf("Ignored double sign from %s at height %d, age of %d past max age of %d", pubkey.Address(), infractionHeight, age, maxEvidenceAge))
		return
	}

	// Double sign confirmed
	logger.Info(fmt.Sprintf("Confirmed double sign from %s at height %d, age of %d less than max age of %d", pubkey.Address(), infractionHeight, age, maxEvidenceAge))

	// We need to retrieve the stake distribution which signed the block, so we subtract ValidatorUpdateDelay from the evidence height.
	// Note that this *can* result in a negative "distributionHeight", up to -ValidatorUpdateDelay,
	// i.e. at the end of the pre-genesis block (none) = at the beginning of the genesis block.
	// That's fine since this is just used to filter unbonding delegations & redelegations.
	distributionHeight := infractionHeight - stake.ValidatorUpdateDelay

	// Cap the amount slashed to the penalty for the worst infraction
	// within the slashing period when this infraction was committed
	fraction := k.SlashFractionDoubleSign(ctx)
	revisedFraction := k.capBySlashingPeriod(ctx, consAddr, fraction, distributionHeight)
	logger.Info(fmt.Sprintf("Fraction slashed capped by slashing period from %v to %v", fraction, revisedFraction))

	// Slash validator
	// `power` is the int64 power of the validator as provided to/by
	// Tendermint. This value is validator.Tokens as sent to Tendermint via
	// ABCI, and now received as evidence.
	// The revisedFraction (which is the new fraction to be slashed) is passed
	// in separately to separately slash unbonding and rebonding delegations.
	k.validatorSet.Slash(ctx, consAddr, distributionHeight, power, revisedFraction)

	// Jail validator if not already jailed
	if !validator.GetJailed() {
		k.validatorSet.Jail(ctx, consAddr)
	}

	// Set or updated validator jail duration
	signInfo, found := k.getValidatorSigningInfo(ctx, consAddr)
	if !found {
		panic(fmt.Sprintf("Expected signing info for validator %s but not found", consAddr))
	}
	signInfo.JailedUntil = time.Add(k.DoubleSignUnbondDuration(ctx))
	k.SetValidatorSigningInfo(ctx, consAddr, signInfo)
}

// handle a validator signature, must be called once per validator per block
// TODO refactor to take in a consensus address, additionally should maybe just take in the pubkey too
func (k Keeper) handleValidatorSignature(ctx sdk.Context, addr crypto.Address, power int64, signed bool) {
	logger := ctx.Logger().With("module", "x/slashing")
	height := ctx.BlockHeight()
	consAddr := sdk.ConsAddress(addr)
	pubkey, err := k.getPubkey(ctx, addr)
	if err != nil {
		panic(fmt.Sprintf("Validator consensus-address %v not found", consAddr))
	}
	// Local index, so counts blocks validator *should* have signed
	// Will use the 0-value default signing info if not present, except for start height
	signInfo, found := k.getValidatorSigningInfo(ctx, consAddr)
	if !found {
		panic(fmt.Sprintf("Expected signing info for validator %s but not found", consAddr))
	}
	index := signInfo.IndexOffset % k.SignedBlocksWindow(ctx)
	signInfo.IndexOffset++

	// Update signed block bit array & counter
	// This counter just tracks the sum of the bit array
	// That way we avoid needing to read/write the whole array each time
	previous := k.getValidatorMissedBlockBitArray(ctx, consAddr, index)
	missed := !signed
	switch {
	case !previous && missed:
		// Array value has changed from not missed to missed, increment counter
		k.setValidatorMissedBlockBitArray(ctx, consAddr, index, true)
		signInfo.MissedBlocksCounter++
	case previous && !missed:
		// Array value has changed from missed to not missed, decrement counter
		k.setValidatorMissedBlockBitArray(ctx, consAddr, index, false)
		signInfo.MissedBlocksCounter--
	default:
		// Array value at this index has not changed, no need to update counter
	}

	if missed {
		logger.Info(fmt.Sprintf("Absent validator %s at height %d, %d missed, threshold %d", addr, height, signInfo.MissedBlocksCounter, k.MinSignedPerWindow(ctx)))
	}
	minHeight := signInfo.StartHeight + k.SignedBlocksWindow(ctx)
	maxMissed := k.SignedBlocksWindow(ctx) - k.MinSignedPerWindow(ctx)
	if height > minHeight && signInfo.MissedBlocksCounter > maxMissed {
		validator := k.validatorSet.ValidatorByConsAddr(ctx, consAddr)
		if validator != nil && !validator.GetJailed() {
			// Downtime confirmed: slash and jail the validator
			logger.Info(fmt.Sprintf("Validator %s past min height of %d and below signed blocks threshold of %d",
				pubkey.Address(), minHeight, k.MinSignedPerWindow(ctx)))
			// We need to retrieve the stake distribution which signed the block, so we subtract ValidatorUpdateDelay from the evidence height,
			// and subtract an additional 1 since this is the LastCommit.
			// Note that this *can* result in a negative "distributionHeight" up to -ValidatorUpdateDelay-1,
			// i.e. at the end of the pre-genesis block (none) = at the beginning of the genesis block.
			// That's fine since this is just used to filter unbonding delegations & redelegations.
			distributionHeight := height - stake.ValidatorUpdateDelay - 1
			k.validatorSet.Slash(ctx, consAddr, distributionHeight, power, k.SlashFractionDowntime(ctx))
			k.validatorSet.Jail(ctx, consAddr)
			signInfo.JailedUntil = ctx.BlockHeader().Time.Add(k.DowntimeUnbondDuration(ctx))
			// We need to reset the counter & array so that the validator won't be immediately slashed for downtime upon rebonding.
			signInfo.MissedBlocksCounter = 0
			signInfo.IndexOffset = 0
			k.clearValidatorMissedBlockBitArray(ctx, consAddr)
		} else {
			// Validator was (a) not found or (b) already jailed, don't slash
			logger.Info(fmt.Sprintf("Validator %s would have been slashed for downtime, but was either not found in store or already jailed",
				pubkey.Address()))
		}
	}

	// Set the updated signing info
	k.SetValidatorSigningInfo(ctx, consAddr, signInfo)
}

func (k Keeper) addPubkey(ctx sdk.Context, pubkey crypto.PubKey) {
	addr := pubkey.Address()
	k.setAddrPubkeyRelation(ctx, addr, pubkey)
}

func (k Keeper) getPubkey(ctx sdk.Context, address crypto.Address) (crypto.PubKey, error) {
	store := ctx.KVStore(k.storeKey)
	var pubkey crypto.PubKey
	err := k.cdc.UnmarshalBinaryLengthPrefixed(store.Get(getAddrPubkeyRelationKey(address)), &pubkey)
	if err != nil {
		return nil, fmt.Errorf("address %v not found", address)
	}
	return pubkey, nil
}

func (k Keeper) setAddrPubkeyRelation(ctx sdk.Context, addr crypto.Address, pubkey crypto.PubKey) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(pubkey)
	store.Set(getAddrPubkeyRelationKey(addr), bz)
}

func (k Keeper) deleteAddrPubkeyRelation(ctx sdk.Context, addr crypto.Address) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(getAddrPubkeyRelationKey(addr))
}
