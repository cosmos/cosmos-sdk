package slashing

import (
	"fmt"
	"time"

	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	stake "github.com/cosmos/cosmos-sdk/x/stake/types"
	abci "github.com/tendermint/tendermint/abci/types"
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
func (k Keeper) handleDoubleSign(ctx sdk.Context, addr crypto.Address, infractionHeight int64, timestamp time.Time, power int64) {
	logger := ctx.Logger().With("module", "x/slashing")
	time := ctx.BlockHeader().Time
	age := time.Sub(timestamp)
	consAddr := sdk.ConsAddress(addr)
	pubkey, err := k.getPubkey(ctx, addr)
	if err != nil {
		panic(fmt.Sprintf("Validator consensus-address %v not found", consAddr))
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
	k.validatorSet.Slash(ctx, consAddr, distributionHeight, power, revisedFraction)

	// Jail validator if not already jailed
	validator := k.validatorSet.ValidatorByConsAddr(ctx, consAddr)
	if !validator.GetJailed() {
		k.validatorSet.Jail(ctx, consAddr)
	}

	// Set or updated validator jail duration
	signInfo, found := k.getValidatorSigningInfo(ctx, consAddr)
	if !found {
		panic(fmt.Sprintf("Expected signing info for validator %s but not found", consAddr))
	}
	signInfo.JailedUntil = time.Add(k.DoubleSignUnbondDuration(ctx))
	k.setValidatorSigningInfo(ctx, consAddr, signInfo)
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
		// If this validator has never been seen before, construct a new SigningInfo with the correct start height
		signInfo = NewValidatorSigningInfo(height, 0, time.Unix(0, 0), 0)
	}
	index := signInfo.IndexOffset % k.SignedBlocksWindow(ctx)
	signInfo.IndexOffset++

	// Update signed block bit array & counter
	// This counter just tracks the sum of the bit array
	// That way we avoid needing to read/write the whole array each time
	previous := k.getValidatorSigningBitArray(ctx, consAddr, index)
	if previous == signed {
		// Array value at this index has not changed, no need to update counter
	} else if previous && !signed {
		// Array value has changed from signed to unsigned, decrement counter
		k.setValidatorSigningBitArray(ctx, consAddr, index, false)
		signInfo.SignedBlocksCounter--
	} else if !previous && signed {
		// Array value has changed from unsigned to signed, increment counter
		k.setValidatorSigningBitArray(ctx, consAddr, index, true)
		signInfo.SignedBlocksCounter++
	}

	if !signed {
		logger.Info(fmt.Sprintf("Absent validator %s at height %d, %d signed, threshold %d", addr, height, signInfo.SignedBlocksCounter, k.MinSignedPerWindow(ctx)))
	}
	minHeight := signInfo.StartHeight + k.SignedBlocksWindow(ctx)
	if height > minHeight && signInfo.SignedBlocksCounter < k.MinSignedPerWindow(ctx) {
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
		} else {
			// Validator was (a) not found or (b) already jailed, don't slash
			logger.Info(fmt.Sprintf("Validator %s would have been slashed for downtime, but was either not found in store or already jailed",
				pubkey.Address()))
		}
	}

	// Set the updated signing info
	k.setValidatorSigningInfo(ctx, consAddr, signInfo)
}

// AddValidators adds the validators to the keepers validator addr to pubkey mapping.
func (k Keeper) AddValidators(ctx sdk.Context, vals []abci.ValidatorUpdate) {
	for i := 0; i < len(vals); i++ {
		val := vals[i]
		pubkey, err := tmtypes.PB2TM.PubKey(val.PubKey)
		if err != nil {
			panic(err)
		}
		k.addPubkey(ctx, pubkey)
	}
}

// TODO: Make a method to remove the pubkey from the map when a validator is unbonded.
func (k Keeper) addPubkey(ctx sdk.Context, pubkey crypto.PubKey) {
	addr := pubkey.Address()
	k.setAddrPubkeyRelation(ctx, addr, pubkey)
}

func (k Keeper) getPubkey(ctx sdk.Context, address crypto.Address) (crypto.PubKey, error) {
	store := ctx.KVStore(k.storeKey)
	var pubkey crypto.PubKey
	err := k.cdc.UnmarshalBinary(store.Get(getAddrPubkeyRelationKey(address)), &pubkey)
	if err != nil {
		return nil, fmt.Errorf("address %v not found", address)
	}
	return pubkey, nil
}

func (k Keeper) setAddrPubkeyRelation(ctx sdk.Context, addr crypto.Address, pubkey crypto.PubKey) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinary(pubkey)
	store.Set(getAddrPubkeyRelationKey(addr), bz)
}

func (k Keeper) deleteAddrPubkeyRelation(ctx sdk.Context, addr crypto.Address) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(getAddrPubkeyRelationKey(addr))
}
