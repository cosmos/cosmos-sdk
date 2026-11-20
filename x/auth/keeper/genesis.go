package keeper

import (
	"cosmossdk.io/collections"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

// InitGenesis - Init store state from genesis data
//
// CONTRACT: old coins from the FeeCollectionKeeper need to be transferred through
// a genesis port script to the new fee collector account
func (ak AccountKeeper) InitGenesis(ctx sdk.Context, data types.GenesisState) {
	if err := ak.Params.Set(ctx, data.Params); err != nil {
		panic(err)
	}

	accounts, err := types.UnpackAccounts(data.Accounts)
	if err != nil {
		panic(err)
	}
	accounts = types.SanitizeGenesisAccounts(accounts)

	// Set the accounts.
	for _, acc := range accounts {
		ak.SetAccount(ctx, acc)
	}

	// Set the pubkey rotation history. Each account's entries are in rotation
	// order and get rotation indices 0..n-1.
	for _, h := range data.PubKeyHistory {
		addr, err := ak.addressCodec.StringToBytes(h.Address)
		if err != nil {
			panic(err)
		}
		for i, entry := range h.Entries {
			if err := ak.PubKeyHistory.Set(ctx, collections.Join(sdk.AccAddress(addr), uint64(i)), entry); err != nil {
				panic(err)
			}
		}
	}

	// Rebuild the rekey index from the accounts' current pubkeys. It is not
	// part of genesis.
	for _, acc := range accounts {
		addr, pk := acc.GetAddress(), acc.GetPubKey()
		if !isRekeyed(addr, pk) {
			continue
		}
		if err := ak.RekeyIndex.Set(ctx, collections.Join(sdk.AccAddress(pk.Address()), addr)); err != nil {
			panic(err)
		}
	}

	ak.GetModuleAccount(ctx, types.FeeCollectorName)
}

// ExportGenesis returns a GenesisState for a given context and keeper
func (ak AccountKeeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	params := ak.GetParams(ctx)

	var genAccounts types.GenesisAccounts
	ak.IterateAccounts(ctx, func(account sdk.AccountI) bool {
		genAccount := account.(types.GenesisAccount)
		genAccounts = append(genAccounts, genAccount)
		return false
	})

	genState := types.NewGenesisState(params, genAccounts)

	// History is walked in key order, so entries of one account are adjacent
	// and sorted by rotation index.
	err := ak.PubKeyHistory.Walk(ctx, nil, func(key collections.Pair[sdk.AccAddress, uint64], entry types.PubKeyHistoryEntry) (bool, error) {
		addr, err := ak.addressCodec.BytesToString(key.K1())
		if err != nil {
			return true, err
		}
		n := len(genState.PubKeyHistory)
		if n == 0 || genState.PubKeyHistory[n-1].Address != addr {
			genState.PubKeyHistory = append(genState.PubKeyHistory, types.GenesisPubKeyHistory{Address: addr})
			n++
		}
		genState.PubKeyHistory[n-1].Entries = append(genState.PubKeyHistory[n-1].Entries, entry)
		return false, nil
	})
	if err != nil {
		panic(err)
	}

	return genState
}
