package keeper_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"
)

func TestGetLastTokenizeShareRecordId(t *testing.T) {
	_, app, ctx := createTestInput(t)
	lastTokenizeShareRecordID := app.StakingKeeper.GetLastTokenizeShareRecordID(ctx)
	require.Equal(t, lastTokenizeShareRecordID, uint64(0))
	app.StakingKeeper.SetLastTokenizeShareRecordID(ctx, 100)
	lastTokenizeShareRecordID = app.StakingKeeper.GetLastTokenizeShareRecordID(ctx)
	require.Equal(t, lastTokenizeShareRecordID, uint64(100))
}

func TestGetTokenizeShareRecord(t *testing.T) {
	_, app, ctx := createTestInput(t)
	addrDels, _ := generateAddresses(app, ctx, 2)
	owner1, owner2 := addrDels[0], addrDels[1]

	tokenizeShareRecord1 := types.TokenizeShareRecord{
		Id:            0,
		Owner:         owner1.String(),
		ModuleAccount: "test-module-account-1",
		Validator:     "test-validator",
	}
	tokenizeShareRecord2 := types.TokenizeShareRecord{
		Id:            1,
		Owner:         owner2.String(),
		ModuleAccount: "test-module-account-2",
		Validator:     "test-validator",
	}
	tokenizeShareRecord3 := types.TokenizeShareRecord{
		Id:            2,
		Owner:         owner1.String(),
		ModuleAccount: "test-module-account-3",
		Validator:     "test-validator",
	}
	err := app.StakingKeeper.AddTokenizeShareRecord(ctx, tokenizeShareRecord1)
	require.NoError(t, err)
	err = app.StakingKeeper.AddTokenizeShareRecord(ctx, tokenizeShareRecord2)
	require.NoError(t, err)
	err = app.StakingKeeper.AddTokenizeShareRecord(ctx, tokenizeShareRecord3)
	require.NoError(t, err)

	tokenizeShareRecord, err := app.StakingKeeper.GetTokenizeShareRecord(ctx, 2)
	require.NoError(t, err)
	require.Equal(t, tokenizeShareRecord, tokenizeShareRecord3)

	tokenizeShareRecord, err = app.StakingKeeper.GetTokenizeShareRecordByDenom(ctx, tokenizeShareRecord2.GetShareTokenDenom())
	require.NoError(t, err)
	require.Equal(t, tokenizeShareRecord, tokenizeShareRecord2)

	tokenizeShareRecords := app.StakingKeeper.GetAllTokenizeShareRecords(ctx)
	require.Equal(t, len(tokenizeShareRecords), 3)

	tokenizeShareRecords = app.StakingKeeper.GetTokenizeShareRecordsByOwner(ctx, owner1)
	require.Equal(t, len(tokenizeShareRecords), 2)

	tokenizeShareRecords = app.StakingKeeper.GetTokenizeShareRecordsByOwner(ctx, owner2)
	require.Equal(t, len(tokenizeShareRecords), 1)
}
