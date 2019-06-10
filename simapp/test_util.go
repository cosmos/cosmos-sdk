package simapp

import (
	"fmt"
	"io"

	"github.com/tendermint/tendermint/libs/log"

	bam "github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/staking"
	cmn "github.com/tendermint/tendermint/libs/common"
	dbm "github.com/tendermint/tendermint/libs/db"
)

const (
	mainStore         = "main"
	accountStore      = "account"
	stakingStore      = "staking"
	slashingStore     = "slashing"
	mintStore         = "mint"
	distributionStore = "distribution"
	supplyStore       = "supply"
	paramStore        = "params"
	govStore          = "gov"
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

func retrieveSimLog(storeName string, appA, appB *SimApp, kvA, kvB cmn.KVPair) (log string) {

	log = fmt.Sprintf("store A %X => %X\n"+
		"store B %X => %X", kvA.Key, kvA.Value, kvB.Key, kvB.Value)

	if len(kvA.Value) == 0 && len(kvB.Value) == 0 {
		return
	}

	if storeName == accountStore {
		var accA auth.Account
		var accB auth.Account
		appA.cdc.MustUnmarshalBinaryBare(kvA.Value, &accA)
		appB.cdc.MustUnmarshalBinaryBare(kvB.Value, &accB)
		log = fmt.Sprintf("%v\n%v", accA, accB)
	}

	return log
}
