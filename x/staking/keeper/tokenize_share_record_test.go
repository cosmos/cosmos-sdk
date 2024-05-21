package keeper_test

import (
	"strconv"
	"testing"
	"time"

	"cosmossdk.io/math"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"
)

func (suite *KeeperTestSuite) TestGetLastTokenizeShareRecordId() {
	ctx, keeper := suite.ctx, suite.stakingKeeper
	lastTokenizeShareRecordID := keeper.GetLastTokenizeShareRecordID(ctx)
	suite.Equal(lastTokenizeShareRecordID, uint64(0))
	keeper.SetLastTokenizeShareRecordID(ctx, 100)
	lastTokenizeShareRecordID = keeper.GetLastTokenizeShareRecordID(ctx)
	suite.Equal(lastTokenizeShareRecordID, uint64(100))
}

func (suite *KeeperTestSuite) TestGetTokenizeShareRecord() {
	ctx, keeper := suite.ctx, suite.stakingKeeper
	addrs := simtestutil.CreateIncrementalAccounts(2)

	owner1, owner2 := addrs[0], addrs[1]
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
	err := keeper.AddTokenizeShareRecord(ctx, tokenizeShareRecord1)
	suite.NoError(err)
	err = keeper.AddTokenizeShareRecord(ctx, tokenizeShareRecord2)
	suite.NoError(err)
	err = keeper.AddTokenizeShareRecord(ctx, tokenizeShareRecord3)
	suite.NoError(err)

	tokenizeShareRecord, err := keeper.GetTokenizeShareRecord(ctx, 2)
	suite.NoError(err)
	suite.Equal(tokenizeShareRecord, tokenizeShareRecord3)

	tokenizeShareRecord, err = keeper.GetTokenizeShareRecordByDenom(ctx, tokenizeShareRecord2.GetShareTokenDenom())
	suite.NoError(err)
	suite.Equal(tokenizeShareRecord, tokenizeShareRecord2)

	tokenizeShareRecords := keeper.GetAllTokenizeShareRecords(ctx)
	suite.Equal(len(tokenizeShareRecords), 3)

	tokenizeShareRecords = keeper.GetTokenizeShareRecordsByOwner(ctx, owner1)
	suite.Equal(len(tokenizeShareRecords), 2)

	tokenizeShareRecords = keeper.GetTokenizeShareRecordsByOwner(ctx, owner2)
	suite.Equal(len(tokenizeShareRecords), 1)
}

func (suite *KeeperTestSuite) TestTokenizeRedelegationShares() {
	srcDelAddr := sdk.AccAddress([]byte("SrcDelAddr"))
	lsmModuleAddr := sdk.AccAddress([]byte("lsmAddr"))
	validatorDstAddress := sdk.ValAddress([]byte("ValDstAddr"))

	reds := []types.Redelegation{
		{
			DelegatorAddress:    srcDelAddr.String(),
			ValidatorSrcAddress: sdk.ValAddress([]byte("ValSrcAddr")).String(),
			ValidatorDstAddress: validatorDstAddress.String(),
			Entries: []types.RedelegationEntry{
				{
					SharesDst: sdk.NewDecFromInt(math.NewInt(5)),
				},
			},
		},
		{
			DelegatorAddress:    srcDelAddr.String(),
			ValidatorSrcAddress: sdk.ValAddress([]byte("ValSrcAddr1")).String(),
			ValidatorDstAddress: validatorDstAddress.String(),
			Entries: []types.RedelegationEntry{
				{
					SharesDst: sdk.NewDecFromInt(math.NewInt(5)),
				},
			},
		},
		{
			DelegatorAddress:    srcDelAddr.String(),
			ValidatorSrcAddress: sdk.ValAddress([]byte("ValSrcAddr2")).String(),
			ValidatorDstAddress: validatorDstAddress.String(),
			Entries: []types.RedelegationEntry{
				{
					SharesDst: sdk.NewDecFromInt(math.NewInt(2)),
				},
			},
		},
		{
			DelegatorAddress:    srcDelAddr.String(),
			ValidatorSrcAddress: sdk.ValAddress([]byte("ValSrcAddr3")).String(),
			ValidatorDstAddress: validatorDstAddress.String(),
			Entries: []types.RedelegationEntry{
				{
					SharesDst: sdk.NewDecFromInt(math.NewInt(5)),
				},
				{
					SharesDst: sdk.NewDecFromInt(math.NewInt(1)),
				},
				{
					SharesDst: sdk.NewDecFromInt(math.NewInt(2)),
				},
			},
		},
	}

	testCases := []struct {
		name                   string
		amountTokenized        int64
		delegatorShares        int64
		redelegations          []types.Redelegation
		expDelegatorRedsShares sdk.Dec
		expTokenizedRedsShares sdk.Dec
		expErr                 bool
	}{
		{
			name:                   "expect to fail when the tokenized amount is greater than the total delegator shares",
			amountTokenized:        20,
			delegatorShares:        10,
			redelegations:          reds[0:2], // total of 10 shares
			expDelegatorRedsShares: sdk.NewDecFromInt(math.NewInt(10)),
			expTokenizedRedsShares: sdk.ZeroDec(),
			expErr:                 true,
		},
		{
			name:                   "expect to fail when the redelegated shares are greater than the total delegator shares",
			amountTokenized:        10,
			delegatorShares:        10,
			redelegations:          reds[0:3], // total of 12 shares
			expDelegatorRedsShares: sdk.NewDecFromInt(math.NewInt(12)),
			expTokenizedRedsShares: sdk.ZeroDec(),
			expErr:                 true,
		},
		{
			name:                   "expect to tokenize all the shares from the redelegations",
			amountTokenized:        20,
			delegatorShares:        20,
			redelegations:          reds, // total of 20 shares
			expDelegatorRedsShares: sdk.ZeroDec(),
			expTokenizedRedsShares: sdk.NewDecFromInt(math.NewInt(20)),
			expErr:                 false,
		},
		{
			name:                   "expect to tokenize half of the shares from the redelegations",
			amountTokenized:        30,
			delegatorShares:        40,
			redelegations:          reds, // total of 20 shares
			expDelegatorRedsShares: sdk.NewDecFromInt(math.NewInt(10)),
			expTokenizedRedsShares: sdk.NewDecFromInt(math.NewInt(10)),
			expErr:                 false,
		},
		{
			name:                   "expect to tokenize 3/5 of the shares from the redelegations",
			amountTokenized:        6,
			delegatorShares:        10,
			redelegations:          reds[0:2], // total of 10 shares
			expDelegatorRedsShares: sdk.NewDecFromInt(math.NewInt(4)),
			expTokenizedRedsShares: sdk.NewDecFromInt(math.NewInt(6)),
			expErr:                 false,
		},
		{
			name:                   "expect to tokenize 5/12 of the shares from the redelegations",
			amountTokenized:        10,
			delegatorShares:        17,
			redelegations:          reds[0:3], // total of 12 shares
			expDelegatorRedsShares: sdk.NewDecFromInt(math.NewInt(7)),
			expTokenizedRedsShares: sdk.NewDecFromInt(math.NewInt(5)),
			expErr:                 false,
		},
	}

	ctx, stakingKeeper := suite.ctx, suite.stakingKeeper
	stakingKeeper.SetValidator(ctx, types.Validator{
		OperatorAddress: validatorDstAddress.String(),
	})

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			cCtx, _ := ctx.CacheContext()
			for _, red := range tc.redelegations {
				stakingKeeper.SetRedelegation(cCtx, red)
			}

			amount := sdk.NewDecFromInt(math.NewInt(tc.delegatorShares))
			err := stakingKeeper.TransferRedelegationsOfTokenizedShares(
				cCtx,
				types.Delegation{
					DelegatorAddress: srcDelAddr.String(),
					ValidatorAddress: validatorDstAddress.String(),
					Shares:           amount,
					ValidatorBond:    false,
				},
				sdk.NewDecFromInt(math.NewInt(tc.amountTokenized)),
				srcDelAddr,
				lsmModuleAddr,
			)

			if !tc.expErr {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}

			delegatorRedsTotalShares, _ := getRedelegationsTotalSharesAndEntriesNum(stakingKeeper.GetRedelegations(cCtx, srcDelAddr, uint16(10)))
			tokenizedRedsShares, _ := getRedelegationsTotalSharesAndEntriesNum(stakingKeeper.GetRedelegations(cCtx, lsmModuleAddr, uint16(10)))

			suite.Require().Equal(tc.expDelegatorRedsShares, delegatorRedsTotalShares)
			suite.Require().Equal(tc.expTokenizedRedsShares, tokenizedRedsShares)
		})
	}
}

func (suite *KeeperTestSuite) TestRedeemTokensForRedelegationShares() {
	srcDelAddr := sdk.AccAddress([]byte("SrcDelAddr"))
	lsmModuleAddr := sdk.AccAddress([]byte("lsmAddr"))
	validatorDstAddress := sdk.ValAddress([]byte("ValDstAddr"))

	reds := []types.Redelegation{
		{
			DelegatorAddress:    lsmModuleAddr.String(),
			ValidatorSrcAddress: sdk.ValAddress([]byte("ValSrcAddr")).String(),
			ValidatorDstAddress: validatorDstAddress.String(),
			Entries: []types.RedelegationEntry{
				{
					SharesDst: sdk.NewDecFromInt(math.NewInt(5)),
				},
				{
					SharesDst: sdk.NewDecFromInt(math.NewInt(5)),
				},
			},
		},
		{
			DelegatorAddress:    lsmModuleAddr.String(),
			ValidatorSrcAddress: sdk.ValAddress([]byte("ValSrcAddr1")).String(),
			ValidatorDstAddress: validatorDstAddress.String(),
			Entries: []types.RedelegationEntry{
				{
					SharesDst: sdk.NewDecFromInt(math.NewInt(2)),
				},
			},
		},
		{
			DelegatorAddress:    lsmModuleAddr.String(),
			ValidatorSrcAddress: sdk.ValAddress([]byte("ValSrcAddr0")).String(),
			ValidatorDstAddress: validatorDstAddress.String(),
			Entries: []types.RedelegationEntry{
				{
					SharesDst: sdk.NewDecFromInt(math.NewInt(5)),
				},
				{
					SharesDst: sdk.NewDecFromInt(math.NewInt(1)),
				},
				{
					SharesDst: sdk.NewDecFromInt(math.NewInt(2)),
				},
			},
		},
	}

	testCases := []struct {
		name                   string
		amountRedeemed         int64
		tokenizedShares        int64
		redelegations          []types.Redelegation
		expTokenizedRedsShares sdk.Dec
		expRedeemRedsShares    sdk.Dec
		expErr                 bool
	}{
		{
			name:                   "expect to fail when the redemption amount is greater than the total delegator shares",
			amountRedeemed:         10,
			tokenizedShares:        5,
			redelegations:          reds[0:1], // total of 10 shares
			expTokenizedRedsShares: sdk.NewDecFromInt(math.NewInt(10)),
			expRedeemRedsShares:    sdk.ZeroDec(), // no shares transferred
			expErr:                 true,
		},
		{
			name:                   "expect to redeem all redelegations",
			amountRedeemed:         10,
			tokenizedShares:        10,
			redelegations:          reds, // total of 20 shares
			expTokenizedRedsShares: sdk.NewDecFromInt(math.NewInt(20)),
			expRedeemRedsShares:    sdk.ZeroDec(),
			expErr:                 true,
		},
		{
			name:                   "expect to redeem half of the redelegations",
			amountRedeemed:         10,
			tokenizedShares:        20,
			redelegations:          reds, // total of 20 shares
			expTokenizedRedsShares: sdk.NewDecFromInt(math.NewInt(10)),
			expRedeemRedsShares:    sdk.NewDecFromInt(math.NewInt(10)),
			expErr:                 false,
		},
		{
			name:                   "expect to redeem 3/5 of the redelegations",
			amountRedeemed:         6,
			tokenizedShares:        10,
			redelegations:          reds[0:1], // total of 10 shares
			expRedeemRedsShares:    sdk.NewDecFromInt(math.NewInt(6)),
			expTokenizedRedsShares: sdk.NewDecFromInt(math.NewInt(4)),
			expErr:                 false,
		},
		{
			name:                   "expect to redeem 3/8 of the redelegations",
			amountRedeemed:         5,
			tokenizedShares:        10,
			redelegations:          reds[2:], // total of 8 shares
			expRedeemRedsShares:    sdk.NewDecFromInt(math.NewInt(3)),
			expTokenizedRedsShares: sdk.NewDecFromInt(math.NewInt(5)),
			expErr:                 false,
		},
	}

	ctx, stakingKeeper := suite.ctx, suite.stakingKeeper
	stakingKeeper.SetValidator(ctx, types.Validator{
		OperatorAddress: validatorDstAddress.String(),
	})

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			cCtx, _ := ctx.CacheContext()
			for _, red := range tc.redelegations {
				stakingKeeper.SetRedelegation(cCtx, red)
			}

			err := stakingKeeper.TransferRedelegationsOfTokenizedShares(
				cCtx,
				types.Delegation{
					ValidatorAddress: validatorDstAddress.String(),
					DelegatorAddress: lsmModuleAddr.String(),
					Shares:           sdk.NewDecFromInt(math.NewInt(tc.tokenizedShares)),
				},
				sdk.NewDecFromInt(math.NewInt(tc.amountRedeemed)),
				lsmModuleAddr,
				srcDelAddr,
			)

			if !tc.expErr {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}

			redeemRedsShares, _ := getRedelegationsTotalSharesAndEntriesNum(stakingKeeper.GetRedelegations(cCtx, srcDelAddr, uint16(10)))
			tokenizedRedsShares, _ := getRedelegationsTotalSharesAndEntriesNum(stakingKeeper.GetRedelegations(cCtx, lsmModuleAddr, uint16(10)))

			suite.Require().Equal(tc.expRedeemRedsShares, redeemRedsShares)
			suite.Require().Equal(tc.expTokenizedRedsShares, tokenizedRedsShares)
		})
	}
}

func (suite *KeeperTestSuite) TestComputeRemainingRedelegatedSharesAfterUnbondings() {
	delAddr := sdk.AccAddress([]byte("delAddr"))
	validatorDstAddress := sdk.ValAddress([]byte("ValDstAddr"))
	timeNow := time.Now()

	reds := []types.Redelegation{
		{
			DelegatorAddress:    delAddr.String(),
			ValidatorSrcAddress: sdk.ValAddress([]byte("ValSrcAddr")).String(),
			ValidatorDstAddress: validatorDstAddress.String(),
			Entries: []types.RedelegationEntry{
				{
					CompletionTime: timeNow,
					SharesDst:      sdk.NewDecFromInt(math.NewInt(5)),
				},
				{
					CompletionTime: timeNow.Add(5 * time.Hour),
					SharesDst:      sdk.NewDecFromInt(math.NewInt(5)),
				},
			},
		},
		{
			DelegatorAddress:    delAddr.String(),
			ValidatorSrcAddress: sdk.ValAddress([]byte("ValSrcAddr1")).String(),
			ValidatorDstAddress: validatorDstAddress.String(),
			Entries: []types.RedelegationEntry{
				{
					CompletionTime: timeNow.Add(10 * time.Hour),
					SharesDst:      sdk.NewDecFromInt(math.NewInt(2)),
				},
			},
		},
		{
			DelegatorAddress:    delAddr.String(),
			ValidatorSrcAddress: sdk.ValAddress([]byte("ValSrcAddr0")).String(),
			ValidatorDstAddress: validatorDstAddress.String(),
			Entries: []types.RedelegationEntry{
				{
					CompletionTime: timeNow.Add(15 * time.Hour),
					SharesDst:      sdk.NewDecFromInt(math.NewInt(5)),
				},
				{
					CompletionTime: timeNow.Add(20 * time.Hour),
					SharesDst:      sdk.NewDecFromInt(math.NewInt(1)),
				},
				{
					CompletionTime: timeNow.Add(30 * time.Hour),
					SharesDst:      sdk.NewDecFromInt(math.NewInt(2)),
				},
			},
		},
	}

	ubd := types.UnbondingDelegation{
		DelegatorAddress: delAddr.String(),
		ValidatorAddress: validatorDstAddress.String(),
		Entries: []types.UnbondingDelegationEntry{
			{
				CompletionTime: timeNow.Add(3 * time.Hour),
				InitialBalance: math.NewInt(5), // 5 - 5 => 0
			},
			{
				CompletionTime: timeNow.Add(8 * time.Hour),
				InitialBalance: math.NewInt(1), // 5 - 1 => 4
			},
			{
				CompletionTime: timeNow.Add(12 * time.Hour),
				InitialBalance: math.NewInt(10), // 4 + 2 - 10 => 0
			},
			{
				CompletionTime: timeNow.Add(25 * time.Hour),
				InitialBalance: math.NewInt(5), // 5 + 1 - 5 + 2 => 3
			},
		},
	}

	ctx, stakingKeeper := suite.ctx, suite.stakingKeeper

	// expect an error when validator isn't set
	_, err := stakingKeeper.ComputeRemainingRedelegatedSharesAfterUnbondings(
		ctx,
		delAddr,
		reds,
		validatorDstAddress,
	)
	suite.Require().Error(err)

	// set validator
	stakingKeeper.SetValidator(ctx, types.Validator{
		OperatorAddress: validatorDstAddress.String(),
		DelegatorShares: sdk.NewDecFromInt(sdk.NewInt(100)),
		Tokens:          sdk.NewInt(100),
	})

	// expect an error when the passed delegator address doesn't match the one in the redelegations
	_, err = stakingKeeper.ComputeRemainingRedelegatedSharesAfterUnbondings(
		ctx,
		sdk.AccAddress([]byte("wrongDelAddr")),
		reds,
		validatorDstAddress,
	)
	suite.Require().Error(err)

	// expect an error when the passed validator address doesn't match the one in the redelegations
	_, err = stakingKeeper.ComputeRemainingRedelegatedSharesAfterUnbondings(
		ctx,
		delAddr,
		reds,
		sdk.ValAddress([]byte("wrongValDstAddr")),
	)
	suite.Require().Error(err)

	// expect no error when no redelegations is passed
	res, err := stakingKeeper.ComputeRemainingRedelegatedSharesAfterUnbondings(
		ctx,
		delAddr,
		[]types.Redelegation{},
		validatorDstAddress,
	)
	suite.Require().NoError(err)
	suite.Require().Equal(sdk.ZeroDec(), res)

	// expect no error when no unbonding delegations exist
	res, err = stakingKeeper.ComputeRemainingRedelegatedSharesAfterUnbondings(
		ctx,
		delAddr,
		reds,
		validatorDstAddress,
	)
	suite.Require().NoError(err)
	suite.Require().Equal(sdk.NewDecFromInt(sdk.NewInt(20)), res)

	stakingKeeper.SetUnbondingDelegation(ctx, ubd)

	// expect no error when no redelegations is passed
	res, err = stakingKeeper.ComputeRemainingRedelegatedSharesAfterUnbondings(
		ctx,
		delAddr,
		[]types.Redelegation{},
		validatorDstAddress,
	)
	suite.Require().NoError(err)
	suite.Require().Equal(sdk.ZeroDec(), res)

	// expect no error
	res, err = stakingKeeper.ComputeRemainingRedelegatedSharesAfterUnbondings(
		ctx,
		delAddr,
		reds,
		validatorDstAddress,
	)
	suite.Require().NoError(err)
	suite.Require().Equal(sdk.NewDecFromInt(sdk.NewInt(3)), res)
}

func TestTransferRedelegationsProperties(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate a random redelegations input and an amount of shares to transfer
		action := getRedelegationsTransferActionGen().Draw(t, "test1")

		// compute the amount of shares from redelegations
		// that required to be tokenized
		delSharesLeft := action.TotalShares.Sub(action.RedelegationsShares)
		if delSharesLeft.GTE(action.Amount) {
			// no redelegations require to be tokenized
			return
		}

		//
		amountToTransfer := action.Amount.Sub(delSharesLeft)
		redelegationsToTransfer, err := keeper.GetMinimumRedelegationsSubsetByShares(amountToTransfer, action.Redelegations)
		if err != nil {
			t.Fatalf(err.Error())
		}

		if err != nil {
			t.Fatalf("do no expect transfer redelegation to fail with action %v\n%s", action, err)
		}

		inputRedsShares, inputRedsEntriesNum := getRedelegationsTotalSharesAndEntriesNum(redelegationsToTransfer)

		transferredRedelegations, remainingRedelegations, ok := keeper.TransferRedelegations(
			amountToTransfer,
			sdk.AccAddress([]byte("lsmModuleAccount")).String(),
			redelegationsToTransfer,
		)

		if !ok {
			t.Fatalf("do no expect transfer redelegation to fail with action %v", action)
		}

		out1TotalShares, ou1tEntriesNum := getRedelegationsTotalSharesAndEntriesNum(remainingRedelegations)
		out2TotalShares, out2EntriesNum := getRedelegationsTotalSharesAndEntriesNum(transferredRedelegations)

		outTotalShares := out1TotalShares.Add(out2TotalShares)
		outEntriesNum := ou1tEntriesNum + out2EntriesNum

		// Properties
		//
		//  - test assumption:
		// For a list of redelegations Reds of length N, where an amount A of shares is to be transferred,
		// it is assumed that the total number of shares from the first redelegation to the second-to-last redelegation in Reds is less than A.
		// Therefore, it is anticipated that all redelegations in Reds will have their shares transferred,
		// except for the last one, which is expected to have at least some of its shares transferred.
		//
		//
		// 	#1 the total shares of redelegations input must be equal
		// 	to the the total shares of redelegations output, i.e. the remaining and transferred redelegations
		require.Equal(
			t,
			inputRedsShares,
			outTotalShares,
		)
		// 	#2 The number of entries in the redelegations output
		// is either equal or one greater than the number entries in the redelegations input
		if inputRedsEntriesNum != outEntriesNum && inputRedsEntriesNum+1 != outEntriesNum {
			t.Fatalf("invalid number of entries in transferred redelegagtions output: have %d expected %d or %d", outEntriesNum, action.EntriesNum+1, action.EntriesNum+1)
		}
		//  #3
		// 	  a) The number of redelegations transferred is equal to the number of redelegations of the input
		require.Equal(t, len(redelegationsToTransfer), len(transferredRedelegations))

		// 	  b) The number of remaining redelegations is either zero or 1.
		require.LessOrEqual(t, len(remainingRedelegations), 1)

		// Check that the redelegations output data are updated as expected
		verifyRedelegationsOutput(t, redelegationsToTransfer, remainingRedelegations, transferredRedelegations, sdk.AccAddress([]byte("dstAddr")).String())
	})
}

type redelegationsTransferAction struct {
	Amount              sdk.Dec
	Redelegations       []types.Redelegation
	RedelegationsShares sdk.Dec
	TotalShares         sdk.Dec
	EntriesNum          int
}

func getRedelegationsTransferActionGen() *rapid.Generator[redelegationsTransferAction] {
	return rapid.Custom(func(t *rapid.T) redelegationsTransferAction {
		// generate total shares
		totalShares := math.NewInt(rapid.Int64Range(1, 1_000_000).Draw(t, "amountToTransfer"))
		// generate upper bound of redelegated shares
		redelegatedSharesUB := math.NewInt(rapid.Int64Range(1, totalShares.Int64()).Draw(t, "amountToTransfer"))
		// generate amount to transfer
		amount := math.NewInt(rapid.Int64Range(1, totalShares.Int64()).Draw(t, "amountToTransfer"))

		action := redelegationsTransferAction{}
		delegatorAddr := sdk.AccAddress([]byte("addr1"))
		validatorDstAddress := sdk.ValAddress([]byte("ValDstAddr"))

		i, totalEntriesNum, redsGenShares := 0, 0, sdk.ZeroDec()
		for {
			red := types.Redelegation{
				DelegatorAddress:    delegatorAddr.String(),
				ValidatorDstAddress: validatorDstAddress.String(),
				ValidatorSrcAddress: sdk.ValAddress([]byte("ValSrcAddr" + strconv.Itoa(i))).String(),
				Entries:             []types.RedelegationEntry{},
			}
			entriesNum := rapid.Int16Range(1, 7).Draw(t, "entriesNum")
			redShares := sdk.ZeroDec()
			for j := 0; j < int(entriesNum); j++ {
				tokensAmount := math.NewInt(rapid.Int64Range(1, redelegatedSharesUB.Int64()).Draw(t, "tokenAmount"))
				sharesAmount := sdk.NewDecFromIntWithPrec(tokensAmount, rapid.Int64Range(0, 6).Draw(t, "sharesAmount"))

				red.Entries = append(red.Entries, types.RedelegationEntry{
					CreationHeight:          rapid.Int64Range(1, 1_000_000).Draw(t, ""),
					CompletionTime:          getTimeGen().Draw(t, ""),
					InitialBalance:          tokensAmount,
					SharesDst:               sharesAmount,
					UnbondingId:             uint64(i),
					UnbondingOnHoldRefCount: int64(j),
				})
				redShares = redShares.Add(sharesAmount)
			}

			if redelegatedSharesUB.ToLegacyDec().LT(redsGenShares.Add(redShares)) {
				break
			}
			redsGenShares = redsGenShares.Add(redShares)
			action.Redelegations = append(action.Redelegations, red)
			totalEntriesNum += int(entriesNum)
			i++
		}
		action.RedelegationsShares = redsGenShares
		action.Amount = amount.ToLegacyDec()
		action.EntriesNum = totalEntriesNum
		action.TotalShares = totalShares.ToLegacyDec()

		return action
	})
}

func getTimeGen() *rapid.Generator[time.Time] {
	return rapid.Custom(func(t *rapid.T) time.Time {
		return time.Unix(rapid.Int64Range(-5.9959e+10, 1.5779e+11).Draw(t, "unix time"), 0).UTC()
	})
}

func getRedelegationsTotalSharesAndEntriesNum(redelegations []types.Redelegation) (sdk.Dec, int) {
	totalShares := sdk.ZeroDec()
	entriesNum := 0
	for _, red := range redelegations {
		for _, entry := range red.Entries {
			totalShares = totalShares.Add(entry.SharesDst)
			entriesNum++
		}
	}
	return totalShares, entriesNum
}

// verify the two redelegations slice output against the redelegations input by ensuring that:
// 1) all redelegations input, except the last one, must have been copied into
// the second redelegations list output with an updated delegator address
// 2) in the last redelegation of the input, at most one entry can be split into two, each of which is distributed
// to the to each of the output lists of the redelegations
func verifyRedelegationsOutput(t *rapid.T, redsIn, redsOut1, redsOut2 []types.Redelegation, delAddr string) bool {
	t.Helper()

	// assume that redsIn and redsOut2 are the same length from propriety #3
	for i := 0; i < len(redsIn); i++ {
		// All redelegations, except the last one, must have identical
		// entries and delegate address
		if i < len(redsIn)-1 {
			if redsOut2[i].DelegatorAddress != delAddr {
				return false
			}
			require.Equal(t, redsIn[i].Entries, redsOut2[i].Entries)
			require.Equal(t, redsIn[i].ValidatorSrcAddress, redsOut2[i].ValidatorSrcAddress)
			require.Equal(t, redsIn[i].ValidatorDstAddress, redsOut2[i].ValidatorDstAddress)
			// If the last redelegation has a split entry,
			// check that the shares amount correctly distributed
		} else {
			split := false
			for j := 0; j < len(redsOut2[i].Entries); j++ {
				// Check if entries have different shares amount
				if entryIn, entryOut2 := redsIn[i].Entries[j], redsOut2[i].Entries[j]; !entryIn.SharesDst.Equal(entryOut2.SharesDst) {
					// ensure that only one entry should is split
					if split {
						return false
					}
					split = true

					// verify that entry is split in two between the redelegations output pair
					require.Len(t, redsOut1, 1)
					require.GreaterOrEqual(t, len(redsOut1[0].Entries), 1)
					entryOut1 := redsOut1[0].Entries[0]
					// verify that their shares sum is equal to the input
					entryIn.SharesDst.Equal(entryOut2.SharesDst.Add(entryOut1.SharesDst))

					// Compare entries
					require.Equal(t, entryIn.CompletionTime, entryOut2.CompletionTime)
					require.Equal(t, entryIn.CreationHeight, entryOut2.CreationHeight)
					require.Equal(t, entryIn.InitialBalance, entryOut2.InitialBalance)
					require.Equal(t, entryIn.UnbondingId, entryOut2.UnbondingId)
					require.Equal(t, entryIn.UnbondingOnHoldRefCount, entryOut2.UnbondingOnHoldRefCount)

					require.Equal(t, entryIn.CompletionTime, entryOut1.CompletionTime)
					require.Equal(t, entryIn.CreationHeight, entryOut1.CreationHeight)
					require.Equal(t, entryIn.InitialBalance, entryOut1.InitialBalance)
					require.Equal(t, entryIn.UnbondingId, entryOut1.UnbondingId)
					require.Equal(t, entryIn.UnbondingOnHoldRefCount, entryOut1.UnbondingOnHoldRefCount)
				} else {
					require.Equal(t, entryIn, entryOut2)
				}
			}
		}
	}

	return true
}
