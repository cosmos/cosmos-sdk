package query

import (
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	auth "github.com/cosmos/cosmos-sdk/x/auth/types"
	staking "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"
	dbm "github.com/tendermint/tm-db"
	"testing"
)

func TestPagination(t *testing.T) {

	keyAcc := sdk.NewKVStoreKey(auth.StoreKey)
	keyStaking := sdk.NewKVStoreKey(staking.StoreKey)

	db := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(db)

	ms.MountStoreWithDB(keyAcc, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyStaking, sdk.StoreTypeIAVL, db)

	err := ms.LoadLatestVersion()
	require.Nil(t, err)

	cdc, _ := simapp.MakeCodecs()

}
