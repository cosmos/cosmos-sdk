package distribution_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/teststaking"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

const (
	totalValidators  = 6
	lazyValidatorIdx = 2
	power            = 100 / totalValidators
)

var (
	valTokens                = sdk.TokensFromConsensusPower(50, sdk.DefaultPowerReduction)
	validatorCommissionRates = stakingtypes.NewCommissionRates(sdk.OneDec(), sdk.OneDec(), sdk.OneDec())
)

type validator struct {
	addr   sdk.ValAddress
	pubkey cryptotypes.PubKey
	votes  []abci.VoteInfo
}

// Context in https://github.com/cosmos/cosmos-sdk/issues/9161
func TestVerifyProposerRewardAssignement(t *testing.T) {
	app := simapp.Setup(t, false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	addrs := simapp.AddTestAddrsIncremental(app, ctx, totalValidators, valTokens)
	tstaking := teststaking.NewHelper(t, ctx, app.StakingKeeper)
	tstaking.Commission = validatorCommissionRates

	// create validators
	validators := make([]validator, totalValidators-1)
	for i := range validators {
		validators[i].addr = sdk.ValAddress(addrs[i])
		validators[i].pubkey = ed25519.GenPrivKey().PubKey()
		validators[i].votes = make([]abci.VoteInfo, totalValidators)
		tstaking.CreateValidatorWithValPower(validators[i].addr, validators[i].pubkey, power, true)
	}
	app.EndBlock(abci.RequestEndBlock{})
	require.NotEmpty(t, app.Commit())

	// verify validators lists
	require.Len(t, app.StakingKeeper.GetAllValidators(ctx), totalValidators)
	for i, val := range validators {
		// verify all validator exists
		require.NotNil(t, app.StakingKeeper.ValidatorByConsAddr(ctx, sdk.GetConsAddress(val.pubkey)))

		// populate last commit info
		voteInfos := []abci.VoteInfo{}
		for _, val2 := range validators {
			voteInfos = append(voteInfos, abci.VoteInfo{
				Validator: abci.Validator{
					Address: sdk.GetConsAddress(val2.pubkey),
					Power:   power,
				},
				SignedLastBlock: true,
			})
		}

		// have this validator only submit the minimum amount of pre-commits
		if i == lazyValidatorIdx {
			for j := totalValidators * 2 / 3; j < len(voteInfos); j++ {
				voteInfos[j].SignedLastBlock = false
			}
		}

		validators[i].votes = voteInfos
	}

	// previous block submitted by validator n-1 (with 100% previous commits) and proposed by lazy validator
	app.BeginBlock(abci.RequestBeginBlock{
		Header:         tmproto.Header{Height: app.LastBlockHeight() + 1, ProposerAddress: sdk.GetConsAddress(validators[lazyValidatorIdx].pubkey)},
		LastCommitInfo: abci.LastCommitInfo{Votes: validators[lazyValidatorIdx-1].votes},
	})
	require.NotEmpty(t, app.Commit())

	// previous block submitted by lazy validator (with 67% previous commits) and proposed by validator n+1
	app.BeginBlock(abci.RequestBeginBlock{
		Header:         tmproto.Header{Height: app.LastBlockHeight() + 1, ProposerAddress: sdk.GetConsAddress(validators[lazyValidatorIdx+1].pubkey)},
		LastCommitInfo: abci.LastCommitInfo{Votes: validators[lazyValidatorIdx].votes},
	})
	require.NotEmpty(t, app.Commit())

	// previous block submitted by validator n+1 (with 100% previous commits) and proposed by validator n+2
	app.BeginBlock(abci.RequestBeginBlock{
		Header:         tmproto.Header{Height: app.LastBlockHeight() + 1, ProposerAddress: sdk.GetConsAddress(validators[lazyValidatorIdx+2].pubkey)},
		LastCommitInfo: abci.LastCommitInfo{Votes: validators[lazyValidatorIdx+1].votes},
	})
	require.NotEmpty(t, app.Commit())

	rewardsValidatorBeforeLazyValidator := app.DistrKeeper.GetValidatorOutstandingRewardsCoins(ctx, validators[lazyValidatorIdx+1].addr)
	rewardsLazyValidator := app.DistrKeeper.GetValidatorOutstandingRewardsCoins(ctx, validators[lazyValidatorIdx].addr)
	rewardsValidatorAfterLazyValidator := app.DistrKeeper.GetValidatorOutstandingRewardsCoins(ctx, validators[lazyValidatorIdx+1].addr)
	require.True(t, rewardsLazyValidator[0].Amount.LT(rewardsValidatorAfterLazyValidator[0].Amount))
	require.Equal(t, rewardsValidatorBeforeLazyValidator, rewardsValidatorAfterLazyValidator)
}
