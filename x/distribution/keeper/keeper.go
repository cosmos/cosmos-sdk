package keeper

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/types/address"
	"github.com/cosmos/cosmos-sdk/types/kv"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// Keeper of the distribution store
type Keeper struct {
	storeKey      sdk.StoreKey
	cdc           codec.BinaryCodec
	paramSpace    paramtypes.Subspace
	authKeeper    types.AccountKeeper
	bankKeeper    types.BankKeeper
	stakingKeeper types.StakingKeeper

	blockedAddrs map[string]bool

	feeCollectorName string // name of the FeeCollector ModuleAccount
}

// NewKeeper creates a new distribution Keeper instance
func NewKeeper(
	cdc codec.BinaryCodec, key sdk.StoreKey, paramSpace paramtypes.Subspace,
	ak types.AccountKeeper, bk types.BankKeeper, sk types.StakingKeeper,
	feeCollectorName string, blockedAddrs map[string]bool,
) Keeper {
	// ensure distribution module account is set
	if addr := ak.GetModuleAddress(types.ModuleName); addr == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.ModuleName))
	}

	// set KeyTable if it has not already been set
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		storeKey:         key,
		cdc:              cdc,
		paramSpace:       paramSpace,
		authKeeper:       ak,
		bankKeeper:       bk,
		stakingKeeper:    sk,
		feeCollectorName: feeCollectorName,
		blockedAddrs:     blockedAddrs,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

// SetWithdrawAddr sets a new address that will receive the rewards upon withdrawal
func (k Keeper) SetWithdrawAddr(ctx sdk.Context, delegatorAddr sdk.AccAddress, withdrawAddr sdk.AccAddress) error {
	if k.blockedAddrs[withdrawAddr.String()] {
		return sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "%s is not allowed to receive external funds", withdrawAddr)
	}

	if !k.GetWithdrawAddrEnabled(ctx) {
		return types.ErrSetWithdrawAddrDisabled
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSetWithdrawAddress,
			sdk.NewAttribute(types.AttributeKeyWithdrawAddress, withdrawAddr.String()),
		),
	)

	k.SetDelegatorWithdrawAddr(ctx, delegatorAddr, withdrawAddr)
	return nil
}

// withdraw rewards from a delegation
func (k Keeper) WithdrawDelegationRewards(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) (sdk.Coins, error) {
	val := k.stakingKeeper.Validator(ctx, valAddr)
	if val == nil {
		return nil, types.ErrNoValidatorDistInfo
	}

	del := k.stakingKeeper.Delegation(ctx, delAddr, valAddr)
	if del == nil {
		return nil, types.ErrEmptyDelegationDistInfo
	}

	// withdraw rewards
	rewards, err := k.withdrawDelegationRewards(ctx, val, del)
	if err != nil {
		return nil, err
	}

	// reinitialize the delegation
	k.initializeDelegation(ctx, valAddr, delAddr)
	return rewards, nil
}

// withdraw validator commission
func (k Keeper) WithdrawValidatorCommission(ctx sdk.Context, valAddr sdk.ValAddress) (sdk.Coins, error) {
	// fetch validator accumulated commission
	accumCommission := k.GetValidatorAccumulatedCommission(ctx, valAddr)
	if accumCommission.Commission.IsZero() {
		return nil, types.ErrNoValidatorCommission
	}

	commission, remainder := accumCommission.Commission.TruncateDecimal()
	k.SetValidatorAccumulatedCommission(ctx, valAddr, types.ValidatorAccumulatedCommission{Commission: remainder}) // leave remainder to withdraw later

	// update outstanding
	outstanding := k.GetValidatorOutstandingRewards(ctx, valAddr).Rewards
	k.SetValidatorOutstandingRewards(ctx, valAddr, types.ValidatorOutstandingRewards{Rewards: outstanding.Sub(sdk.NewDecCoinsFromCoins(commission...))})

	if !commission.IsZero() {
		accAddr := sdk.AccAddress(valAddr)
		withdrawAddr := k.GetDelegatorWithdrawAddr(ctx, accAddr)
		err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, withdrawAddr, commission)
		if err != nil {
			return nil, err
		}
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeWithdrawCommission,
			sdk.NewAttribute(sdk.AttributeKeyAmount, commission.String()),
		),
	)

	return commission, nil
}

// GetTotalRewards returns the total amount of fee distribution rewards held in the store
func (k Keeper) GetTotalRewards(ctx sdk.Context) (totalRewards sdk.DecCoins) {
	k.IterateValidatorOutstandingRewards(ctx,
		func(_ sdk.ValAddress, rewards types.ValidatorOutstandingRewards) (stop bool) {
			totalRewards = totalRewards.Add(rewards.Rewards...)
			return false
		},
	)

	return totalRewards
}

// FundCommunityPool allows an account to directly fund the community fund pool.
// The amount is first added to the distribution module account and then directly
// added to the pool. An error is returned if the amount cannot be sent to the
// module account.
func (k Keeper) FundCommunityPool(ctx sdk.Context, amount sdk.Coins, sender sdk.AccAddress) error {
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, sender, types.ModuleName, amount); err != nil {
		return err
	}

	feePool := k.GetFeePool(ctx)
	feePool.CommunityPool = feePool.CommunityPool.Add(sdk.NewDecCoinsFromCoins(amount...)...)
	k.SetFeePool(ctx, feePool)

	return nil
}

// SaveGrant method grants the provided authorization to the grantee on the granter's account
// with the provided expiration time. If there is an existing authorization grant for the
// same `sdk.Msg` type, this grant overwrites that.
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

// DeleteGrant revokes any authorization for the provided message type granted to the grantee
// by the granter.
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

func (k Keeper) PerformRestake(ctx sdk.Context, delegator sdk.AccAddress, validator sdk.ValAddress) error {
	coins, err := k.WithdrawDelegationRewards(ctx, delegator, validator)
	if err != nil {
		return err
	}

	baseDenom := "uscrt"
	//if baseDenom == "" {
	//	baseDenom = sdk.DefaultBondDenom
	//}
	coinsToRedelegate := coins.AmountOf(baseDenom)

	println(coinsToRedelegate.String())

	val := k.stakingKeeper.Validator(ctx, validator)

	_, err = k.stakingKeeper.DoDelegate(ctx, delegator, coinsToRedelegate, 1, val, true)
	if err != nil {
		return err
	}

	return nil
}

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

//// grantStoreKey - return authorization store key
//// Items are stored with the following key: values
////
//// - 0x01<granterAddressLen (1 Byte)><granterAddress_Bytes><granteeAddressLen (1 Byte)><granteeAddress_Bytes><msgType_Bytes>: Grant
//func grantStoreKey(grantee sdk.AccAddress, granter sdk.AccAddress, msgType string) []byte {
//	m := conv.UnsafeStrToBytes(msgType)
//	granter = address.MustLengthPrefix(granter)
//	grantee = address.MustLengthPrefix(grantee)
//
//	l := 1 + len(grantee) + len(granter) + len(m)
//	key := make([]byte, l)
//	copy(key, GrantKey)
//	copy(key[1:], granter)
//	copy(key[1+len(granter):], grantee)
//	copy(key[l-len(m):], m)
//	//	fmt.Println(">>>> len", l, key)
//	return key
//}
//
//// addressesFromGrantStoreKey - split granter & grantee address from the authorization key
//func addressesFromGrantStoreKey(key []byte) (granterAddr, granteeAddr sdk.AccAddress) {
//	// key is of format:
//	// 0x01<granterAddressLen (1 Byte)><granterAddress_Bytes><granteeAddressLen (1 Byte)><granteeAddress_Bytes><msgType_Bytes>
//	kv.AssertKeyAtLeastLength(key, 2)
//	granterAddrLen := key[1] // remove prefix key
//	kv.AssertKeyAtLeastLength(key, int(3+granterAddrLen))
//	granterAddr = sdk.AccAddress(key[2 : 2+granterAddrLen])
//	granteeAddrLen := int(key[2+granterAddrLen])
//	kv.AssertKeyAtLeastLength(key, 4+int(granterAddrLen+byte(granteeAddrLen)))
//	granteeAddr = sdk.AccAddress(key[3+granterAddrLen : 3+granterAddrLen+byte(granteeAddrLen)])
//
//	return granterAddr, granteeAddr
//}
//
//// firstAddressFromGrantStoreKey parses the first address only
//func firstAddressFromGrantStoreKey(key []byte) sdk.AccAddress {
//	addrLen := key[0]
//	return sdk.AccAddress(key[1 : 1+addrLen])
//}

//func (k Keeper) PerformAllRestakes(ctx )
