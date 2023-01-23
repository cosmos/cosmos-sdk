package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/kv"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
)

func (k Keeper) GetRestakeValidatorsForDelegator(ctx sdk.Context, delegator sdk.AccAddress) (validators []string) {

	delegatorPrefix := keyPrefixFromDelegator(delegator)
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, delegatorPrefix)

	defer iter.Close()
	for ; iter.Valid(); iter.Next() {

		key := iter.Key()
		//k.Logger(ctx).Info(fmt.Sprintf("from iter - %s", hex.EncodeToString(key)))
		_, validator := addressesFromRestakeKeyStore(key)
		//k.Logger(ctx).Info(fmt.Sprintf("from iter - %s %s - %s", delegator, validator, hex.EncodeToString(key)))
		validators = append(validators, validator.String())
	}
	return
}

// SaveAutoRestakeEntry method does stuff
func (k Keeper) SaveAutoRestakeEntry(ctx sdk.Context, delegator sdk.AccAddress, validator sdk.ValAddress) error {
	store := ctx.KVStore(k.storeKey)

	delegation := k.stakingKeeper.Delegation(ctx, delegator, validator)
	valInfo := k.stakingKeeper.Validator(ctx, validator)

	currentStake := valInfo.TokensFromShares(delegation.GetShares())

	if k.GetMinimumRestakeThreshold(ctx).GT(currentStake) {
		return types.ErrNotEnoughStakeForAuto
	}

	skey := autoRestakeKey(delegator, validator)

	store.Set(skey, []byte("k"))
	return nil
}

// DeleteAutoRestakeEntry does more stuff
func (k Keeper) DeleteAutoRestakeEntry(ctx sdk.Context, delegator sdk.AccAddress, validator sdk.ValAddress) error {
	store := ctx.KVStore(k.storeKey)
	skey := autoRestakeKey(delegator, validator)
	found := store.Has(skey)
	if !found {
		return sdkerrors.ErrNotFound.Wrap("authorization not found")
	}
	store.Delete(skey)
	return nil
}

// PerformRestake does the thing it's meant to do
func (k Keeper) PerformRestake(ctx sdk.Context, delegator sdk.AccAddress, validator sdk.ValAddress) error {
	coins, err := k.WithdrawDelegationRewards(ctx, delegator, validator)
	if err != nil {
		return err
	}

	baseDenom := k.stakingKeeper.BondDenom(ctx)

	coinsToRedelegate := coins.AmountOf(baseDenom)

	val := k.stakingKeeper.Validator(ctx, validator)

	_, err = k.stakingKeeper.DoDelegate(ctx, delegator, coinsToRedelegate, 1, val, true)
	if err != nil {
		return err
	}

	return nil
}

// autoRestakeKey returns the byte array that we use in the store
func autoRestakeKey(delegator sdk.AccAddress, validator sdk.ValAddress) []byte {
	// key is of format:
	// 0xF0<granterAddressLen (1 Byte)><granterAddress_Bytes><granteeAddressLen (1 Byte)><granteeAddress_Bytes>
	delegator = address.MustLengthPrefix(delegator)
	validator = address.MustLengthPrefix(validator)

	//fmt.Println("saving key: ", hex.EncodeToString(delegator), hex.EncodeToString(validator))

	////	l := 1 + len(grantee) + len(granter) + len(m)
	////	key := make([]byte, l)
	////	copy(key, GrantKey)
	////	copy(key[1:], granter)
	////	copy(key[1+len(granter):], grantee)
	////	copy(key[l-len(m):], m)
	////	//	fmt.Println(">>>> len", l, key)
	////	return key

	l := 1 + len(delegator) + len(validator)
	key := make([]byte, l)
	copy(key, types.AutoRestakeEntryPrefix)
	copy(key[1:], delegator)
	copy(key[1+len(delegator):], validator)

	return key
}

func addressesFromRestakeKeyStore(key []byte) (delegatorAddr sdk.AccAddress, validatorAddr sdk.ValAddress) {
	// key is of format:
	// 0xF0<granterAddressLen (1 Byte)><granterAddress_Bytes><granteeAddressLen (1 Byte)><granteeAddress_Bytes>
	kv.AssertKeyAtLeastLength(key, 2)

	delAddrLen := key[1] // remove prefix key
	kv.AssertKeyAtLeastLength(key, int(3+delAddrLen))
	valAddrLen := int(key[2+delAddrLen])
	kv.AssertKeyAtLeastLength(key, 3+int(delAddrLen+byte(valAddrLen)))

	// lol go code sucks
	delegatorAddr = sdk.AccAddress(key[2 : 2+delAddrLen])
	validatorAddr = sdk.ValAddress(key[3+delAddrLen : 3+delAddrLen+byte(valAddrLen)])

	return delegatorAddr, validatorAddr
}

// delegatorAddressFromRestakeKeyStore parses the delegator address only - will be useful for iterating by delegator
// (probably)
func delegatorAddressFromRestakeKeyStore(key []byte) sdk.AccAddress {
	addrLen := key[0]
	return sdk.AccAddress(key[1 : 1+addrLen])
}

// delegatorAddressFromRestakeKeyStore parses the delegator address only - will be useful for iterating by delegator
// (probably)
func keyPrefixFromDelegator(delegator sdk.AccAddress) []byte {

	// this is a stupid function name
	delegator = address.MustLengthPrefix(delegator)

	l := 1 + len(delegator)
	key := make([]byte, l)
	copy(key, types.AutoRestakeEntryPrefix)
	copy(key[1:], delegator)

	return key
}
