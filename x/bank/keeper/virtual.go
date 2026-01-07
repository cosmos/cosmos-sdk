package keeper

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/hex"
	"fmt"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
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
	k.emitSendCoinsEvents(ctx, fromAddr, toAddr, amt)
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
	k.emitSendCoinsEvents(ctx, fromAddr, toAddr, amt)
	return nil
}

func (k BaseSendKeeper) addVirtualCoins(ctx context.Context, addr sdk.AccAddress, amt sdk.Coins) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := sdkCtx.ObjectStore(k.objStoreKey)

	// bytes containing account address followed by the txn index
	key := make([]byte, len(addr)+8)
	copy(key, addr)
	binary.BigEndian.PutUint64(key[len(addr):], uint64(sdkCtx.TxIndex()))

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

	key := make([]byte, len(addr)+8)
	copy(key, addr)
	binary.BigEndian.PutUint64(key[len(addr):], uint64(sdkCtx.TxIndex()))

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

	it := store.Iterator(nil, nil)
	defer it.Close()
	for ; it.Valid(); it.Next() {
		if len(it.Key()) <= 8 {
			return fmt.Errorf("unexpected key length: %s", hex.EncodeToString(it.Key()))
		}

		addr := it.Key()[:len(it.Key())-8]
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
