package blockstm

import (
	"bytes"

	"cosmossdk.io/collections"
	storetypes "cosmossdk.io/store/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	// Keep these prefixes in sync with x/auth and x/bank collection keys.
	// authAccountNumberSeqPrefix and bankBalancesStoreKeyPrefix are both
	// NewPrefix(2) intentionally — they reference different stores (acc vs bank).
	authAccountStorePrefix     = collections.NewPrefix(1)
	authAccountNumberSeqPrefix = collections.NewPrefix(2)
	bankBalancesStoreKeyPrefix = collections.NewPrefix(2)
)

func EstimateBank(tx sdk.Tx, coinDenom string) (Locations, bool) {
	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return nil, false
	}
	feePayer := sdk.AccAddress(feeTx.FeePayer())
	// balance key
	balanceKey, err := collections.EncodeKeyWithPrefix(
		bankBalancesStoreKeyPrefix,
		collections.PairKeyCodec(sdk.AccAddressKey, collections.StringKey),
		collections.Join(feePayer, coinDenom),
	)
	if err != nil {
		return nil, false
	}
	return Locations{balanceKey}, true
}

func EstimateAuth(tx sdk.Tx, authKVStore storetypes.KVStore) (Locations, bool) {
	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return nil, false
	}
	feePayer := sdk.AccAddress(feeTx.FeePayer())
	// account key
	accKey, err := collections.EncodeKeyWithPrefix(
		authAccountStorePrefix,
		sdk.AccAddressKey,
		feePayer,
	)
	if err != nil {
		return nil, false
	}
	authEstimate := Locations{accKey}
	if authKVStore != nil {
		if !authKVStore.Has(accKey) {
			globalAccountNumberKey := authAccountNumberSeqPrefix.Bytes()
			authEstimate = append(authEstimate, globalAccountNumberKey)
			// Sort so MVMemory sees deterministic key ordering.
			if bytes.Compare(authEstimate[0], authEstimate[1]) > 0 {
				authEstimate[0], authEstimate[1] = authEstimate[1], authEstimate[0]
			}
		}
	}
	return authEstimate, true
}
