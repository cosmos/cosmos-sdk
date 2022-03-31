package distribution_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/teststaking"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

const (
	lazyValidatorAddr = "cosmosvaloper15ky9du8a2wlstz6fpx3p4mqpjyrm5cgzyl2307"
	totalValidators   = 15
)

var (
	valTokens                = sdk.TokensFromConsensusPower(50, sdk.DefaultPowerReduction)
	validatorCommissionRates = stakingtypes.NewCommissionRates(sdk.OneDec(), sdk.OneDec(), sdk.OneDec())
)

type validator struct {
	addr   sdk.ValAddress
	pubkey cryptotypes.PubKey
	power  int64
	votes  []abci.VoteInfo
}

// Context in https://github.com/cosmos/cosmos-sdk/issues/9161
func Test_VerifyProposerRewardAssignement(t *testing.T) {
	a := assert.New(t)

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
		validators[i].power = 100 / totalValidators
		validators[i].votes = make([]abci.VoteInfo, totalValidators)
		tstaking.CreateValidatorWithValPower(validators[i].addr, validators[i].pubkey, validators[i].power, true)
	}
	app.EndBlock(abci.RequestEndBlock{})
	a.NotEmpty(app.Commit())

	// verify validators lists
	a.Len(app.StakingKeeper.GetAllValidators(ctx), totalValidators)
	for i, val := range validators {
		// verify all validator exists
		a.NotNil(app.StakingKeeper.ValidatorByConsAddr(ctx, sdk.GetConsAddress(val.pubkey)))

		// populate last commit info
		voteInfos := []abci.VoteInfo{}
		for _, val2 := range validators {
			voteInfos = append(voteInfos, abci.VoteInfo{
				Validator: abci.Validator{
					Address: sdk.GetConsAddress(val2.pubkey),
					Power:   val2.power,
				},
				SignedLastBlock: true,
			})
		}

		// have this validator only submit the minimum amount of pre-commits
		if val.addr.String() == lazyValidatorAddr {
			for i := totalValidators * 2 / 3; i < len(voteInfos); i++ {
				voteInfos[i].SignedLastBlock = false
			}
		}

		validators[i].votes = voteInfos
	}

	var validatorAfterLazyValidator sdk.ValAddress
	for i := 0; i < len(validators); i++ {
		var votes []abci.VoteInfo
		if i > 0 {
			votes = validators[i-1].votes

			if validators[i-1].addr.String() == lazyValidatorAddr {
				validatorAfterLazyValidator = validators[i].addr
			}
		}

		app.BeginBlock(abci.RequestBeginBlock{
			Header:         tmproto.Header{Height: app.LastBlockHeight() + 1, ProposerAddress: sdk.GetConsAddress(validators[i].pubkey)},
			LastCommitInfo: abci.LastCommitInfo{Votes: votes},
		})
		a.NotEmpty(app.Commit())
	}

	lazyValAddr, err := sdk.ValAddressFromBech32(lazyValidatorAddr)
	a.NoError(err)
	rewardsLazyValidator := app.DistrKeeper.GetValidatorOutstandingRewardsCoins(ctx, lazyValAddr)
	rewardsValidatorAfterLazyValidator := app.DistrKeeper.GetValidatorOutstandingRewardsCoins(ctx, validatorAfterLazyValidator)
	a.True(rewardsLazyValidator[0].Amount.LT(rewardsValidatorAfterLazyValidator[0].Amount))
}
