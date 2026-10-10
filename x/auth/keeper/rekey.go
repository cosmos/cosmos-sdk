package keeper

import (
	"context"

	"cosmossdk.io/collections"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

// applyRekey replaces acc's pubkey with newPk and saves the account. It records
// the replaced pubkey in PubKeyHistory under the account's next rotation index
// and keeps RekeyIndex in sync: the pair for the old key is removed, and a pair
// for the new key is added only if the new key's natural address differs from
// the account address.
//
// The caller must already have validated newPk (types.ValidateRekeyPubKey) and
// its proof of possession. History is keyed by rotation index rather than block
// height, so an account may rotate more than once per block and a zero-height
// export cannot cause a collision.
func (ak AccountKeeper) applyRekey(ctx context.Context, acc sdk.AccountI, newPk cryptotypes.PubKey) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	addr := acc.GetAddress()

	index, err := ak.nextRotationIndex(ctx, addr)
	if err != nil {
		return err
	}
	histKey := collections.Join(addr, index)

	newKeyAddr := sdk.AccAddress(newPk.Address())
	newKeyAddrStr, err := ak.addressCodec.BytesToString(newKeyAddr)
	if err != nil {
		return err
	}

	entry := types.PubKeyHistoryEntry{
		ReplacedAtHeight: sdkCtx.BlockHeight(),
		ReplacedAtTime:   sdkCtx.BlockTime(),
		NewKeyAddress:    newKeyAddrStr,
	}

	if oldPk := acc.GetPubKey(); oldPk != nil {
		entry.PubKey, err = codectypes.NewAnyWithValue(oldPk)
		if err != nil {
			return err
		}
		if err := ak.RekeyIndex.Remove(ctx, collections.Join(sdk.AccAddress(oldPk.Address()), addr)); err != nil {
			return err
		}
	}

	if err := ak.PubKeyHistory.Set(ctx, histKey, entry); err != nil {
		return err
	}
	if !newKeyAddr.Equals(addr) {
		if err := ak.RekeyIndex.Set(ctx, collections.Join(newKeyAddr, addr)); err != nil {
			return err
		}
	}

	if err := acc.SetPubKey(newPk); err != nil {
		return err
	}
	ak.SetAccount(ctx, acc)
	return nil
}

// nextRotationIndex returns one more than the highest rotation index stored
// for addr, or 0 if the account has no history.
func (ak AccountKeeper) nextRotationIndex(ctx context.Context, addr sdk.AccAddress) (uint64, error) {
	rng := collections.NewPrefixedPairRange[sdk.AccAddress, uint64](addr).Descending()
	iter, err := ak.PubKeyHistory.Iterate(ctx, rng)
	if err != nil {
		return 0, err
	}
	defer iter.Close()
	if !iter.Valid() {
		return 0, nil
	}
	key, err := iter.Key()
	if err != nil {
		return 0, err
	}
	return key.K2() + 1, nil
}

// isRekeyed reports whether pk is set, is not a module credential, and does not
// hash to addr, meaning the account holding it must appear in RekeyIndex.
func isRekeyed(addr sdk.AccAddress, pk cryptotypes.PubKey) bool {
	if pk == nil {
		return false
	}
	if _, ok := pk.(*types.ModuleCredential); ok {
		return false
	}
	return !sdk.AccAddress(pk.Address()).Equals(addr)
}
