package v046_test

import (
	"math/rand"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	"github.com/cosmos/cosmos-sdk/x/auth/apptestutils"
	"github.com/cosmos/cosmos-sdk/x/auth/keeper"

	"github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

// TestMigrateMapAccAddressToAccNumberKey test cases for state migration of map to accAddr to accNum
func TestMigrateMapAccAddressToAccNumberKey(t *testing.T) {
	var (
		accountKeeper keeper.AccountKeeper
	)

	app, err := simtestutil.Setup(
		apptestutils.AppConfig,
		&accountKeeper,
	)
	require.NoError(t, err)

	// new base account
	senderPrivKey := secp256k1.GenPrivKey()
	randAccNumber := uint64(rand.Intn(100000-10000) + 10000)
	acc := types.NewBaseAccount(senderPrivKey.PubKey().Address().Bytes(), senderPrivKey.PubKey(), randAccNumber, 0)

	ctx := app.BaseApp.NewContext(false, tmproto.Header{Time: time.Now()})

	// migrator
	m := keeper.NewMigrator(accountKeeper, app.GRPCQueryRouter())
	// set the account to store with map acc addr to acc number
	require.NoError(t, m.V45_SetAccount(ctx, acc))

	testCases := []struct {
		name        string
		doMigration bool
		accNum      uint64
	}{
		{
			name:        "without state migration",
			doMigration: false,
			accNum:      acc.AccountNumber,
		},
		{
			name:        "with state migration",
			doMigration: true,
			accNum:      acc.AccountNumber,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.doMigration {
				require.NoError(t, m.Migrate2to3(ctx))
			}

			//  get the account address by acc id
			accAddr := accountKeeper.GetAccountAddressByID(ctx, tc.accNum)

			if tc.doMigration {
				require.Equal(t, accAddr, acc.Address)
			} else {
				require.Equal(t, len(accAddr), 0)
			}
		})
	}
}
