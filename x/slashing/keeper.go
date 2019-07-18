package slashing

import (
	"fmt"
	"time"

	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/slashing/types"
)

// Keeper of the slashing store
type Keeper struct {
	storeKey   sdk.StoreKey
	cdc        *codec.Codec
	sk         types.StakingKeeper
	paramspace params.Subspace

	// codespace
	codespace sdk.CodespaceType
}

// NewKeeper creates a slashing keeper
func NewKeeper(cdc *codec.Codec, key sdk.StoreKey, sk types.StakingKeeper, paramspace params.Subspace, codespace sdk.CodespaceType) Keeper {
	keeper := Keeper{
		storeKey:   key,
		cdc:        cdc,
		sk:         sk,
		paramspace: paramspace.WithKeyTable(ParamKeyTable()),
		codespace:  codespace,
	}
	return keeper
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// handle a validator signing two blocks at the same height
// power: power of the double-signing validator at the height of infraction
func (k Keeper) HandleDoubleSign(ctx sdk.Context, addr crypto.Address, infractionHeight int64, timestamp time.Time, power int64) {
	logger := k.Logger(ctx)

	// calculate the age of the evidence
	time := ctx.BlockHeader().Time
	age := time.Sub(timestamp)

	// fetch the validator public key
	consAddr := sdk.ConsAddress(addr)
	pubkey, err := k.getPubkey(ctx, addr)
	if err != nil {
		// Ignore evidence that cannot be handled.
		// NOTE:
		// We used to panic with:
		// `panic(fmt.Sprintf("Validator consensus-address %v not found", consAddr))`,
		// but this couples the expectations of the app to both Tendermint and
		// the simulator.  Both are expected to provide the full range of
		// allowable but none of the disallowed evidence types.  Instead of
		// getting this coordination right, it is easier to relax the
		// constraints and ignore evidence that cannot be handled.
		return
	}

	// Reject evidence if the double-sign is too old
	if age > k.MaxEvidenceAge(ctx) {
		logger.Info(fmt.Sprintf("Ignored double sign from %s at height %d, age of %d past max age of %d",
			sdk.ConsAddress(pubkey.Address()), infractionHeight, age, k.MaxEvidenceAge(ctx)))
		return
	}

	// Get validator and signing info
	validator := k.sk.ValidatorByConsAddr(ctx, consAddr)
	if validator == nil || validator.IsUnbonded() {
		// Defensive.
		// Simulation doesn't take unbonding periods into account, and
		// Tendermint might break this assumption at some point.
		return
	}

	// fetch the validator signing info
	signInfo, found := k.getValidatorSigningInfo(ctx, consAddr)
	if !found {
		panic(fmt.Sprintf("Expected signing info for validator %s but not found", consAddr))
	}

	// validator is already tombstoned
	if signInfo.Tombstoned {
		logger.Info(fmt.Sprintf("Ignored double sign from %s at height %d, validator already tombstoned", sdk.ConsAddress(pubkey.Address()), infractionHeight))
		return
	}

	// double sign confirmed
	logger.Info(fmt.Sprintf("Confirmed double sign from %s at height %d, age of %d", sdk.ConsAddress(pubkey.Address()), infractionHeight, age))

	// We need to retrieve the stake distribution which signed the block, so we subtract ValidatorUpdateDelay from the evidence height.
	// Note that this *can* result in a negative "distributionHeight", up to -ValidatorUpdateDelay,
	// i.e. at the end of the pre-genesis block (none) = at the beginning of the genesis block.
	// That's fine since this is just used to filter unbonding delegations & redelegations.
	distributionHeight := infractionHeight - sdk.ValidatorUpdateDelay

	// get the percentage slash penalty fraction
	fraction := k.SlashFractionDoubleSign(ctx)

	// Slash validator
	// `power` is the int64 power of the validator as provided to/by
	// Tendermint. This value is validator.Tokens as sent to Tendermint via
	// ABCI, and now received as evidence.
	// The fraction is passed in to separately to slash unbonding and rebonding delegations.
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSlash,
			sdk.NewAttribute(types.AttributeKeyAddress, consAddr.String()),
			sdk.NewAttribute(types.AttributeKeyPower, fmt.Sprintf("%d", power)),
			sdk.NewAttribute(types.AttributeKeyReason, types.AttributeValueDoubleSign),
		),
	)
	k.sk.Slash(ctx, consAddr, distributionHeight, power, fraction)

	// Jail validator if not already jailed
	// begin unbonding validator if not already unbonding (tombstone)
	if !validator.IsJailed() {
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeSlash,
				sdk.NewAttribute(types.AttributeKeyJailed, consAddr.String()),
			),
		)
		k.sk.Jail(ctx, consAddr)
	}

	// Set tombstoned to be true
	signInfo.Tombstoned = true

	// Set jailed until to be forever (max time)
	signInfo.JailedUntil = types.DoubleSignJailEndTime

	// Set validator signing info
	k.SetValidatorSigningInfo(ctx, consAddr, signInfo)
}

// handle a validator signature, must be called once per validator per block
// TODO refactor to take in a consensus address, additionally should maybe just take in the pubkey too
func (k Keeper) HandleValidatorSignature(ctx sdk.Context, addr crypto.Address, power int64, signed bool) {
	logger := k.Logger(ctx)
	height := ctx.BlockHeight()
	consAddr := sdk.ConsAddress(addr)
	pubkey, err := k.getPubkey(ctx, addr)
	if err != nil {
		panic(fmt.Sprintf("Validator consensus-address %s not found", consAddr))
	}

	// fetch signing info
	signInfo, found := k.getValidatorSigningInfo(ctx, consAddr)
	if !found {
		panic(fmt.Sprintf("Expected signing info for validator %s but not found", consAddr))
	}

	// this is a relative index, so it counts blocks the validator *should* have signed
	// will use the 0-value default signing info if not present, except for start height
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
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeLiveness,
				sdk.NewAttribute(types.AttributeKeyAddress, consAddr.String()),
				sdk.NewAttribute(types.AttributeKeyMissedBlocks, fmt.Sprintf("%d", signInfo.MissedBlocksCounter)),
				sdk.NewAttribute(types.AttributeKeyHeight, fmt.Sprintf("%d", height)),
			),
		)

		logger.Info(
			fmt.Sprintf("Absent validator %s (%s) at height %d, %d missed, threshold %d", consAddr, pubkey, height, signInfo.MissedBlocksCounter, k.MinSignedPerWindow(ctx)))
	}

	minHeight := signInfo.StartHeight + k.SignedBlocksWindow(ctx)
	maxMissed := k.SignedBlocksWindow(ctx) - k.MinSignedPerWindow(ctx)

	// if we are past the minimum height and the validator has missed too many blocks, punish them
	if height > minHeight && signInfo.MissedBlocksCounter > maxMissed {
		validator := k.sk.ValidatorByConsAddr(ctx, consAddr)
		if validator != nil && !validator.IsJailed() {

			// Downtime confirmed: slash and jail the validator
			logger.Info(fmt.Sprintf("Validator %s past min height of %d and below signed blocks threshold of %d",
				consAddr, minHeight, k.MinSignedPerWindow(ctx)))

			// We need to retrieve the stake distribution which signed the block, so we subtract ValidatorUpdateDelay from the evidence height,
			// and subtract an additional 1 since this is the LastCommit.
			// Note that this *can* result in a negative "distributionHeight" up to -ValidatorUpdateDelay-1,
			// i.e. at the end of the pre-genesis block (none) = at the beginning of the genesis block.
			// That's fine since this is just used to filter unbonding delegations & redelegations.
			distributionHeight := height - sdk.ValidatorUpdateDelay - 1

			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					types.EventTypeSlash,
					sdk.NewAttribute(types.AttributeKeyAddress, consAddr.String()),
					sdk.NewAttribute(types.AttributeKeyPower, fmt.Sprintf("%d", power)),
					sdk.NewAttribute(types.AttributeKeyReason, types.AttributeValueMissingSignature),
					sdk.NewAttribute(types.AttributeKeyJailed, consAddr.String()),
				),
			)
			k.sk.Slash(ctx, consAddr, distributionHeight, power, k.SlashFractionDowntime(ctx))
			k.sk.Jail(ctx, consAddr)

			signInfo.JailedUntil = ctx.BlockHeader().Time.Add(k.DowntimeJailDuration(ctx))

			// We need to reset the counter & array so that the validator won't be immediately slashed for downtime upon rebonding.
			signInfo.MissedBlocksCounter = 0
			signInfo.IndexOffset = 0
			k.clearValidatorMissedBlockBitArray(ctx, consAddr)
		} else {
			// Validator was (a) not found or (b) already jailed, don't slash
			logger.Info(
				fmt.Sprintf("Validator %s would have been slashed for downtime, but was either not found in store or already jailed", consAddr),
			)
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
	err := k.cdc.UnmarshalBinaryLengthPrefixed(store.Get(types.GetAddrPubkeyRelationKey(address)), &pubkey)
	if err != nil {
		return nil, fmt.Errorf("address %s not found", sdk.ConsAddress(address))
	}
	return pubkey, nil
}

func (k Keeper) setAddrPubkeyRelation(ctx sdk.Context, addr crypto.Address, pubkey crypto.PubKey) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(pubkey)
	store.Set(types.GetAddrPubkeyRelationKey(addr), bz)
}

func (k Keeper) deleteAddrPubkeyRelation(ctx sdk.Context, addr crypto.Address) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetAddrPubkeyRelationKey(addr))
}
