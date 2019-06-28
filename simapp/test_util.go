package simapp

import (
	"bytes"
	"fmt"
	"io"

	cmn "github.com/tendermint/tendermint/libs/common"

	"github.com/cosmos/cosmos-sdk/codec"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"

	bam "github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/staking"
)

// DONTCOVER

// NewSimAppUNSAFE is used for debugging purposes only.
//
// NOTE: to not use this function with non-test code
func NewSimAppUNSAFE(logger log.Logger, db dbm.DB, traceStore io.Writer, loadLatest bool,
	invCheckPeriod uint, baseAppOptions ...func(*bam.BaseApp),
) (gapp *SimApp, keyMain, keyStaking *sdk.KVStoreKey, stakingKeeper staking.Keeper) {

	gapp = NewSimApp(logger, db, traceStore, loadLatest, invCheckPeriod, baseAppOptions...)
	return gapp, gapp.keyMain, gapp.keyStaking, gapp.stakingKeeper
}

// getSimulationLog unmarshals the KVPair's Value to the corresponding type based on the
// each's module store key and the prefix bytes of the KVPair's key.
// nolint: deadcode unused
func getSimulationLog(storeName string, cdcA, cdcB *codec.Codec, kvA, kvB cmn.KVPair) (log string) {
	log = fmt.Sprintf("store A %s => %X\nstore B %s => %X\n", storeName, storeName, kvB.Key, kvB.Value)

	if len(kvA.Value) == 0 && len(kvB.Value) == 0 {
		return
	}

	switch storeName {
	case auth.StoreKey:
		return decodeAccountStore(cdcA, cdcB, kvA, kvB)
	default:
		return storeName
	}
}

// nolint: deadcode unused
func decodeAccountStore(cdcA, cdcB *codec.Codec, kvA, kvB cmn.KVPair) string {
	switch {
	case bytes.Equal(kvA.Key[:1], auth.AddressStoreKeyPrefix):
		var accA, accB auth.Account
		cdcA.MustUnmarshalBinaryBare(kvA.Value, &accA)
		cdcB.MustUnmarshalBinaryBare(kvB.Value, &accB)
		return fmt.Sprintf("%v\n%v", accA, accB)
	case bytes.Equal(kvA.Key, auth.GlobalAccountNumberKey):
		var globalAccNumberA, globalAccNumberB uint64
		cdcA.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &globalAccNumberA)
		cdcB.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &globalAccNumberB)
		return fmt.Sprintf("GlobalAccNumberA: %d\nGlobalAccNumberB: %d", globalAccNumberA, globalAccNumberB)
	}
	return fmt.Sprintf("\nstore %s", kvA.Key)
}
