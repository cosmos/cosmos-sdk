package slashing

import (
	"encoding/binary"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/stake"
	crypto "github.com/tendermint/go-crypto"
)

const (
	// MaxEvidenceAge - Max age for evidence - 21 days (3 weeks)
	// TODO Should this be a governance parameter or just modifiable with SoftwareUpgradeProposals?
	// MaxEvidenceAge = 60 * 60 * 24 * 7 * 3
	// TODO Temporarily set to 2 minutes for testnets.
	MaxEvidenceAge int64 = 60 * 2

	// SignedBlocksWindow - sliding window for downtime slashing
	// TODO Governance parameter?
	// TODO Temporarily set to 100 blocks for testnets
	SignedBlocksWindow int64 = 100

	// Downtime slashing threshold - 50%
	// TODO Governance parameter?
	MinSignedPerWindow int64 = SignedBlocksWindow / 2

	// Downtime unbond duration - 1 day
	// TODO Governance parameter?
	DowntimeUnbondDuration int64 = 86400
)

var (
	// SlashFractionDoubleSign - currently 5%
	// TODO Governance parameter?
	SlashFractionDoubleSign = sdk.NewRat(1).Quo(sdk.NewRat(20))

	// SlashFractionDowntime - currently 0
	// TODO Governance parameter?
	SlashFractionDowntime = sdk.ZeroRat()
)

// Keeper of the slashing store
type Keeper struct {
	storeKey    sdk.StoreKey
	cdc         *wire.Codec
	stakeKeeper stake.Keeper

	// codespace
	codespace sdk.CodespaceType
}

// NewKeeper creates a slashing keeper
func NewKeeper(cdc *wire.Codec, key sdk.StoreKey, sk stake.Keeper, codespace sdk.CodespaceType) Keeper {
	keeper := Keeper{
		storeKey:    key,
		cdc:         cdc,
		stakeKeeper: sk,
		codespace:   codespace,
	}
	return keeper
}

// handle a validator signing two blocks at the same height
func (k Keeper) handleDoubleSign(ctx sdk.Context, height int64, timestamp int64, pubkey crypto.PubKey) {
	logger := ctx.Logger().With("module", "x/slashing")
	age := ctx.BlockHeader().Time - timestamp
	if age > MaxEvidenceAge {
		logger.Info(fmt.Sprintf("Ignored double sign from %v at height %d, age of %d past max age of %d", pubkey.Address(), height, age, MaxEvidenceAge))
		return
	}
	logger.Info(fmt.Sprintf("Confirmed double sign from %v at height %d, age of %d less than max age of %d", pubkey.Address(), height, age, MaxEvidenceAge))
	validator := k.stakeKeeper.ValidatorByPubKey(ctx, pubkey)
	if validator == nil {
		panic(fmt.Errorf("Attempted to slash nonexistent validator with address %s", pubkey.Address()))
	}
	validator.Slash(ctx, height, SlashFractionDoubleSign)
	logger.Info(fmt.Sprintf("Slashed validator %s by fraction %v for double-sign at height %d", pubkey.Address(), SlashFractionDoubleSign, height))
}

// handle a validator signature, must be called once per validator per block
func (k Keeper) handleValidatorSignature(ctx sdk.Context, pubkey crypto.PubKey, signed bool) {
	logger := ctx.Logger().With("module", "x/slashing")
	height := ctx.BlockHeight()
	if !signed {
		logger.Info(fmt.Sprintf("Absent validator %v at height %d", pubkey.Address(), height))
	}
	index := height % SignedBlocksWindow
	address := pubkey.Address()
	signInfo, _ := k.getValidatorSigningInfo(ctx, address)
	previous := k.getValidatorSigningBitArray(ctx, address, index)
	if previous && !signed {
		k.setValidatorSigningBitArray(ctx, address, index, false)
		signInfo.SignedBlocksCounter--
		k.setValidatorSigningInfo(ctx, address, signInfo)
	} else if !previous && signed {
		k.setValidatorSigningBitArray(ctx, address, index, true)
		signInfo.SignedBlocksCounter++
		k.setValidatorSigningInfo(ctx, address, signInfo)
	}
	minHeight := signInfo.StartHeight + SignedBlocksWindow
	if height > minHeight && signInfo.SignedBlocksCounter < MinSignedPerWindow {
		validator := k.stakeKeeper.ValidatorByPubKey(ctx, pubkey)
		if validator == nil {
			panic(fmt.Errorf("Attempted to slash nonexistent validator with address %s", pubkey.Address()))
		}
		validator.Slash(ctx, height, SlashFractionDowntime)
		validator.ForceUnbond(ctx, DowntimeUnbondDuration)
		logger.Info(fmt.Sprintf("Slashed validator %s by fraction %v and unbonded for downtime at height %d, cannot rebond for %ds",
			address, SlashFractionDowntime, height, DowntimeUnbondDuration))
	}
}

func (k Keeper) getValidatorSigningInfo(ctx sdk.Context, address sdk.Address) (info validatorSigningInfo, found bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(validatorSigningInfoKey(address))
	if bz == nil {
		found = false
	} else {
		k.cdc.MustUnmarshalBinary(bz, &info)
		found = true
	}
	return
}

func (k Keeper) setValidatorSigningInfo(ctx sdk.Context, address sdk.Address, info validatorSigningInfo) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinary(info)
	store.Set(validatorSigningInfoKey(address), bz)
}

func (k Keeper) getValidatorSigningBitArray(ctx sdk.Context, address sdk.Address, index int64) (signed bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(validatorSigningBitArrayKey(address, index))
	if bz == nil {
		// lazy: treat empty key as unsigned
		signed = false
	} else {
		k.cdc.MustUnmarshalBinary(bz, &signed)
	}
	return
}

func (k Keeper) setValidatorSigningBitArray(ctx sdk.Context, address sdk.Address, index int64, signed bool) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinary(signed)
	store.Set(validatorSigningBitArrayKey(address, index), bz)
}

type validatorSigningInfo struct {
	StartHeight         int64 `json:"start_height"`
	SignedBlocksCounter int64 `json:"signed_blocks_counter"`
}

func validatorSigningInfoKey(v sdk.Address) []byte {
	return append([]byte{0x01}, v.Bytes()...)
}

func validatorSigningBitArrayKey(v sdk.Address, i int64) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(i))
	return append([]byte{0x02}, append(v.Bytes(), b...)...)
}
