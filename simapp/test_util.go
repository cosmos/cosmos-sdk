package simapp

import (
	"io"

	"github.com/tendermint/tendermint/libs/log"

	bam "github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	dbm "github.com/tendermint/tendermint/libs/db"
)

// NewSimAppUNSAFE is used for debugging purposes only.
//
// NOTE: to not use this function with non-test code
func NewSimAppUNSAFE(logger log.Logger, db dbm.DB, traceStore io.Writer, loadLatest bool,
	invCheckPeriod uint, baseAppOptions ...func(*bam.BaseApp),
) (gapp *SimApp, keyMain, keyStaking *sdk.KVStoreKey, stakingKeeper staking.Keeper) {

	gapp = NewSimApp(logger, db, traceStore, loadLatest, invCheckPeriod, baseAppOptions...)
	return gapp, gapp.keyMain, gapp.keyStaking, gapp.stakingKeeper
}
