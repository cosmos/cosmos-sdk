package staking

import (
	crypto "github.com/tendermint/go-crypto"

	sdk "github.com/cosmos/cosmos-sdk/types"
	wire "github.com/cosmos/cosmos-sdk/wire"
)

type StakingMapper struct {
	key sdk.StoreKey
	cdc *wire.Codec
}

func NewMapper(key sdk.StoreKey) StakingMapper {
	cdc := wire.NewCodec()
	return StakingMapper{
		key: key,
		cdc: cdc,
	}
}

func (sm StakingMapper) getBondInfo(ctx sdk.Context, addr sdk.Address) bondInfo {
	store := ctx.KVStore(sm.key)
	bz := store.Get(addr)
	if bz == nil {
		return bondInfo{}
	}
	var bi bondInfo
	err := sm.cdc.UnmarshalBinary(bz, &bi)
	if err != nil {
		panic(err)
	}
	return bi
}

func (sm StakingMapper) setBondInfo(ctx sdk.Context, addr sdk.Address, bi bondInfo) {
	store := ctx.KVStore(sm.key)
	bz, err := sm.cdc.MarshalBinary(bi)
	if err != nil {
		panic(err)
	}
	store.Set(addr, bz)
}

func (sm StakingMapper) deleteBondInfo(ctx sdk.Context, addr sdk.Address) {
	store := ctx.KVStore(sm.key)
	store.Delete(addr)
}

func (sm StakingMapper) Bond(ctx sdk.Context, addr sdk.Address, pubKey crypto.PubKey, power int64) (int64, sdk.Error) {
	bi := sm.getBondInfo(ctx, addr)
	if bi.isEmpty() {
		bi = bondInfo{
			PubKey: pubKey,
			Power:  power,
		}
		sm.setBondInfo(ctx, addr, bi)
		return bi.Power, nil
	}

	newPower := bi.Power + power
	newBi := bondInfo{
		PubKey: bi.PubKey,
		Power:  newPower,
	}
	sm.setBondInfo(ctx, addr, newBi)

	return newBi.Power, nil
}

func (sm StakingMapper) Unbond(ctx sdk.Context, addr sdk.Address) (crypto.PubKey, int64, sdk.Error) {
	bi := sm.getBondInfo(ctx, addr)
	if bi.isEmpty() {
		return crypto.PubKey{}, 0, ErrInvalidUnbond()
	}
	sm.deleteBondInfo(ctx, addr)
	return bi.PubKey, bi.Power, nil
}

type bondInfo struct {
	PubKey crypto.PubKey
	Power  int64
}

func (bi bondInfo) isEmpty() bool {
	if bi == (bondInfo{}) {
		return true
	}
	return false
}
