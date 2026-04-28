package keeper

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/hex"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

// SendCoinsFromAccountToModuleVirtual sends coins from account to a virtual module account.
func (k BaseSendKeeper) SendCoinsFromAccountToModuleVirtual(
	ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins,
) error {
	recipientAcc := k.ak.GetModuleAccount(ctx, recipientModule)
	if recipientAcc == nil {
		panic(errorsmod.Wrapf(sdkerrors.ErrUnknownAddress, "module account %s does not exist", recipientModule))
	}

	return k.SendCoinsToVirtual(ctx, senderAddr, recipientAcc.GetAddress(), amt)
}

// SendCoinsFromModuleToAccountVirtual sends coins from account to a virtual module account.
func (k BaseSendKeeper) SendCoinsFromModuleToAccountVirtual(
	ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins,
) error {
	senderAddr := k.ak.GetModuleAddress(senderModule)
	if senderAddr == nil {
		panic(errorsmod.Wrapf(sdkerrors.ErrUnknownAddress, "module account %s does not exist", senderModule))
	}

	if k.BlockedAddr(recipientAddr) {
		return errorsmod.Wrapf(sdkerrors.ErrUnauthorized, "%s is not allowed to receive funds", recipientAddr)
	}

	return k.SendCoinsFromVirtual(ctx, senderAddr, recipientAddr, amt)
}

// MintCoinsVirtual creates new coins and buffers them in virtual state.
// The balance is credited to module account virtually, and supply change is buffered.
// Both are settled at end of block to avoid BlockSTM conflicts.
func (k BaseKeeper) MintCoinsVirtual(ctx context.Context, moduleName string, amounts sdk.Coins) error {
	err := k.mintCoinsRestrictionFn(ctx, amounts)
	if err != nil {
		k.logger.Error(fmt.Sprintf("Module %q attempted to mint coins %s it doesn't have permission for, error %v", moduleName, amounts, err))
		return err
	}
	acc := k.ak.GetModuleAccount(ctx, moduleName)
	if acc == nil {
		panic(errorsmod.Wrapf(sdkerrors.ErrUnknownAddress, "module account %s does not exist", moduleName))
	}

	if !acc.HasPermission(authtypes.Minter) {
		panic(errorsmod.Wrapf(sdkerrors.ErrUnauthorized, "module account %s does not have permissions to mint tokens", moduleName))
	}

	if !amounts.IsValid() {
		return errorsmod.Wrap(sdkerrors.ErrInvalidCoins, amounts.String())
	}

	k.addVirtualCoins(ctx, acc.GetAddress(), amounts)

	for _, amount := range amounts {
		k.addVirtualSupply(ctx, amount.GetDenom(), amount.Amount)
	}

	k.logger.Debug("minted coins from module account (virtual)", "amount", amounts.String(), "from", moduleName)

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx.EventManager().EmitEvent(
		types.NewCoinMintEvent(acc.GetAddress(), amounts),
	)

	return nil
}

// BurnCoinsVirtual burns coins by deducting from virtual balance and buffering supply change.
// Use this when burning coins that were minted via MintCoinsVirtual.
func (k BaseKeeper) BurnCoinsVirtual(ctx context.Context, moduleName string, amounts sdk.Coins) error {
	acc := k.ak.GetModuleAccount(ctx, moduleName)
	if acc == nil {
		panic(errorsmod.Wrapf(sdkerrors.ErrUnknownAddress, "module account %s does not exist", moduleName))
	}

	if !acc.HasPermission(authtypes.Burner) {
		panic(errorsmod.Wrapf(sdkerrors.ErrUnauthorized, "module account %s does not have permissions to burn tokens", moduleName))
	}
	if !amounts.IsValid() {
		return errorsmod.Wrap(sdkerrors.ErrInvalidCoins, amounts.String())
	}

	err := k.subVirtualCoins(ctx, acc.GetAddress(), amounts)
	if err != nil {
		return err
	}

	for _, amount := range amounts {
		k.subVirtualSupply(ctx, amount.GetDenom(), amount.Amount)
	}

	k.logger.Debug("burned tokens from module account (from virtual)", "amount", amounts.String(), "from", moduleName)

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx.EventManager().EmitEvent(
		types.NewCoinBurnEvent(acc.GetAddress(), amounts),
	)

	return nil
}

// SendCoinsToVirtual accumulate the recipient's coins in a per-transaction transient state,
// which are sumed up and added to the real account at the end of block.
// Events are emitted the same as normal send.
func (k BaseSendKeeper) SendCoinsToVirtual(ctx context.Context, fromAddr, toAddr sdk.AccAddress, amt sdk.Coins) error {
	var err error
	err = k.subUnlockedCoins(ctx, fromAddr, amt)
	if err != nil {
		return err
	}

	toAddr, err = k.sendRestriction.apply(ctx, fromAddr, toAddr, amt)
	if err != nil {
		return err
	}

	k.addVirtualCoins(ctx, toAddr, amt)
	if err := k.emitSendCoinsEvents(ctx, fromAddr, toAddr, amt); err != nil {
		return err
	}
	return nil
}

// SendCoinsFromVirtual deduct coins from virtual from account and send to recipient account.
func (k BaseSendKeeper) SendCoinsFromVirtual(ctx context.Context, fromAddr, toAddr sdk.AccAddress, amt sdk.Coins) error {
	var err error
	err = k.subVirtualCoins(ctx, fromAddr, amt)
	if err != nil {
		return err
	}

	toAddr, err = k.sendRestriction.apply(ctx, fromAddr, toAddr, amt)
	if err != nil {
		return err
	}

	err = k.addCoins(ctx, toAddr, amt)
	if err != nil {
		return err
	}

	k.ensureAccountCreated(ctx, toAddr)
	if err := k.emitSendCoinsEvents(ctx, fromAddr, toAddr, amt); err != nil {
		return err
	}
	return nil
}

func (k BaseSendKeeper) addVirtualCoins(ctx context.Context, addr sdk.AccAddress, amt sdk.Coins) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := sdkCtx.ObjectStore(k.objStoreKey)

	// key: VirtualBalancePrefix + address + txIndex (8 bytes)
	key := make([]byte, len(types.VirtualBalancePrefix)+len(addr)+8)
	copy(key, types.VirtualBalancePrefix)
	copy(key[len(types.VirtualBalancePrefix):], addr)
	binary.BigEndian.PutUint64(key[len(types.VirtualBalancePrefix)+len(addr):], uint64(sdkCtx.TxIndex()))

	var coins sdk.Coins
	value := store.Get(key)
	if value != nil {
		coins = value.(sdk.Coins)
	}
	coins = coins.Add(amt...)
	store.Set(key, coins)
}

func (k BaseSendKeeper) subVirtualCoins(ctx context.Context, addr sdk.AccAddress, amt sdk.Coins) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := sdkCtx.ObjectStore(k.objStoreKey)

	// key: VirtualBalancePrefix + address + txIndex (8 bytes)
	key := make([]byte, len(types.VirtualBalancePrefix)+len(addr)+8)
	copy(key, types.VirtualBalancePrefix)
	copy(key[len(types.VirtualBalancePrefix):], addr)
	binary.BigEndian.PutUint64(key[len(types.VirtualBalancePrefix)+len(addr):], uint64(sdkCtx.TxIndex()))

	value := store.Get(key)
	if value == nil {
		return errorsmod.Wrapf(
			sdkerrors.ErrInsufficientFunds,
			"spendable balance 0 is smaller than %s",
			amt,
		)
	}
	spendable := value.(sdk.Coins)
	balance, hasNeg := spendable.SafeSub(amt...)
	if hasNeg {
		return errorsmod.Wrapf(
			sdkerrors.ErrInsufficientFunds,
			"spendable balance %s is smaller than %s",
			spendable, amt,
		)
	}
	if balance.IsZero() {
		store.Delete(key)
	} else {
		store.Set(key, balance)
	}

	return nil
}

// CreditVirtualAccounts sum up the transient coins and add them to the real account,
// should be called at end blocker.
func (k BaseSendKeeper) CreditVirtualAccounts(ctx context.Context) error {
	// No-op if we're not using the objStore to accumulate to module accounts
	if k.objStoreKey == nil {
		return nil
	}
	store := sdk.UnwrapSDKContext(ctx).ObjectStore(k.objStoreKey)

	var toAddr sdk.AccAddress
	sum := sdk.NewMapCoins(nil)
	flushCurrentAddr := func() error {
		if len(sum) == 0 {
			// nothing to flush
			return nil
		}

		if err := k.addCoins(ctx, toAddr, sum.ToCoins()); err != nil {
			return err
		}
		clear(sum)

		k.ensureAccountCreated(ctx, toAddr)
		return nil
	}

	// Calculate end key for VirtualBalancePrefix range
	endKey := make([]byte, len(types.VirtualBalancePrefix))
	copy(endKey, types.VirtualBalancePrefix)
	endKey[len(endKey)-1]++

	it := store.Iterator(types.VirtualBalancePrefix, endKey)
	defer it.Close()
	for ; it.Valid(); it.Next() {
		key := it.Key()
		// key format: VirtualBalancePrefix + address + txIndex (8 bytes)
		minKeyLen := len(types.VirtualBalancePrefix) + 8
		if len(key) <= minKeyLen {
			return fmt.Errorf("unexpected key length: %s", hex.EncodeToString(key))
		}

		addr := key[len(types.VirtualBalancePrefix) : len(key)-8]
		if !bytes.Equal(toAddr, addr) {
			if err := flushCurrentAddr(); err != nil {
				return err
			}
			toAddr = addr
		}

		sum.Add(it.Value().(sdk.Coins)...)
	}

	return flushCurrentAddr()
}

func (k BaseKeeper) addVirtualSupply(ctx context.Context, denom string, amount math.Int) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := sdkCtx.ObjectStore(k.objStoreKey)

	// key: VirtualSupplyPrefix + denom + txIndex (8 bytes)
	denomBytes := []byte(denom)
	key := make([]byte, len(types.VirtualSupplyPrefix)+len(denomBytes)+8)
	copy(key, types.VirtualSupplyPrefix)
	copy(key[len(types.VirtualSupplyPrefix):], denomBytes)
	binary.BigEndian.PutUint64(key[len(types.VirtualSupplyPrefix)+len(denomBytes):], uint64(sdkCtx.TxIndex()))

	var current math.Int
	value := store.Get(key)
	if value != nil {
		current = value.(math.Int)
	} else {
		current = math.ZeroInt()
	}
	current = current.Add(amount)
	store.Set(key, current)
}

func (k BaseKeeper) subVirtualSupply(ctx context.Context, denom string, amount math.Int) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := sdkCtx.ObjectStore(k.objStoreKey)

	// key: VirtualSupplyPrefix + denom + txIndex (8 bytes)
	denomBytes := []byte(denom)
	key := make([]byte, len(types.VirtualSupplyPrefix)+len(denomBytes)+8)
	copy(key, types.VirtualSupplyPrefix)
	copy(key[len(types.VirtualSupplyPrefix):], denomBytes)
	binary.BigEndian.PutUint64(key[len(types.VirtualSupplyPrefix)+len(denomBytes):], uint64(sdkCtx.TxIndex()))

	var current math.Int
	value := store.Get(key)
	if value != nil {
		current = value.(math.Int)
	} else {
		current = math.ZeroInt()
	}
	current = current.Sub(amount)
	store.Set(key, current)
}

// SettleVirtualSupply aggregates virtual supply changes and applies them to the real Supply.
// Should be called at end blocker.
func (k BaseKeeper) SettleVirtualSupply(ctx context.Context) error {
	if k.objStoreKey == nil {
		return nil
	}
	store := sdk.UnwrapSDKContext(ctx).ObjectStore(k.objStoreKey)

	supplyChanges := make(map[string]math.Int)

	// Calculate end key for VirtualSupplyPrefix range
	endKey := make([]byte, len(types.VirtualSupplyPrefix))
	copy(endKey, types.VirtualSupplyPrefix)
	endKey[len(endKey)-1]++

	it := store.Iterator(types.VirtualSupplyPrefix, endKey)
	defer it.Close()
	for ; it.Valid(); it.Next() {
		key := it.Key()
		// key format: VirtualSupplyPrefix + denom + txIndex (8 bytes)
		minKeyLen := len(types.VirtualSupplyPrefix) + 8
		if len(key) <= minKeyLen {
			return fmt.Errorf("unexpected supply key length: %s", hex.EncodeToString(key))
		}

		denom := string(key[len(types.VirtualSupplyPrefix) : len(key)-8])
		amount := it.Value().(math.Int)

		if existing, ok := supplyChanges[denom]; ok {
			supplyChanges[denom] = existing.Add(amount)
		} else {
			supplyChanges[denom] = amount
		}
	}

	for denom, change := range supplyChanges {
		if change.IsZero() {
			continue
		}
		// Use k.Supply.Get directly to avoid double-counting virtual supply
		amt, err := k.Supply.Get(ctx, denom)
		if err != nil {
			amt = math.ZeroInt()
		}
		k.setSupply(ctx, sdk.NewCoin(denom, amt.Add(change)))
	}

	return nil
}
