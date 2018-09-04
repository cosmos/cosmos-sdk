package stake

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/stake/keeper"
)

// burn burn burn
func BurnFeeHandler(ctx sdk.Context, _ sdk.Tx, collectedFees sdk.Coins) {}

// Called every block, process inflation, update validator set
func EndBlocker(ctx sdk.Context, k keeper.Keeper) (ValidatorUpdates []abci.Validator) {
	pool := k.GetPool(ctx)

	// Process provision inflation
	blockTime := ctx.BlockHeader().Time
	if blockTime.Sub(pool.InflationLastTime) >= time.Hour {
		params := k.GetParams(ctx)
		pool.InflationLastTime = blockTime
		pool = pool.ProcessProvisions(params)
		k.SetPool(ctx, pool)
	}

	// reset the intra-transaction counter
	k.SetIntraTxCounter(ctx, 0)

	// calculate validator set changes
	ValidatorUpdates = k.GetTendermintUpdates(ctx)
	k.ClearTendermintUpdates(ctx)
	return
}

/*
func AllocateFees(feesCollected sdk.Coins, global Global, proposer ValidatorDistribution,
              sumPowerPrecommitValidators, totalBondedTokens, communityTax,
              proposerCommissionRate sdk.Dec)

     feesCollectedDec = MakeDecCoins(feesCollected)
     proposerReward = feesCollectedDec * (0.01 + 0.04
                       * sumPowerPrecommitValidators / totalBondedTokens)

     commission = proposerReward * proposerCommissionRate
     proposer.PoolCommission += commission
     proposer.Pool += proposerReward - commission

     communityFunding = feesCollectedDec * communityTax
     global.CommunityFund += communityFunding

     poolReceived = feesCollectedDec - proposerReward - communityFunding
     global.Pool += poolReceived
     global.EverReceivedPool += poolReceived
     global.LastReceivedPool = poolReceived

     SetValidatorDistribution(proposer)
     SetGlobal(global)

/////////////////////////////////////////////////////////////////////////////////////////////////////////

type TxWithdrawDelegationRewardsAll struct {
    delegatorAddr sdk.AccAddress
    withdrawAddr  sdk.AccAddress // address to make the withdrawal to
}

func WithdrawDelegationRewardsAll(delegatorAddr, withdrawAddr sdk.AccAddress)
    height = GetHeight()
    withdraw = GetDelegatorRewardsAll(delegatorAddr, height)
    AddCoins(withdrawAddr, withdraw.TruncateDecimal())

func GetDelegatorRewardsAll(delegatorAddr sdk.AccAddress, height int64) DecCoins

    // get all distribution scenarios
    delegations = GetDelegations(delegatorAddr)

    // collect all entitled rewards
    withdraw = 0
    pool = stake.GetPool()
    global = GetGlobal()
    for delegation = range delegations
        delInfo = GetDelegationDistInfo(delegation.DelegatorAddr,
                        delegation.ValidatorAddr)
        valInfo = GetValidatorDistInfo(delegation.ValidatorAddr)
        validator = GetValidator(delegation.ValidatorAddr)

        global, diWithdraw = delInfo.WithdrawRewards(global, valInfo, height, pool.BondedTokens,
                   validator.Tokens, validator.DelegatorShares, validator.Commission)
        withdraw += diWithdraw

    SetGlobal(global)
    return withdraw

/////////////////////////////////////////////////////////////////////////////////////////////////////////

type TxWithdrawDelegationReward struct {
    delegatorAddr sdk.AccAddress
    validatorAddr sdk.AccAddress
    withdrawAddr  sdk.AccAddress // address to make the withdrawal to
}

func WithdrawDelegationReward(delegatorAddr, validatorAddr, withdrawAddr sdk.AccAddress)
    height = GetHeight()

    // get all distribution scenarios
    pool = stake.GetPool()
    global = GetGlobal()
    delInfo = GetDelegationDistInfo(delegatorAddr,
                    validatorAddr)
    valInfo = GetValidatorDistInfo(validatorAddr)
    validator = GetValidator(validatorAddr)

    global, withdraw = delInfo.WithdrawRewards(global, valInfo, height, pool.BondedTokens,
               validator.Tokens, validator.DelegatorShares, validator.Commission)

    SetGlobal(global)
    AddCoins(withdrawAddr, withdraw.TruncateDecimal())

/////////////////////////////////////////////////////////////////////////////////////////////////////////

type TxWithdrawValidatorRewardsAll struct {
    operatorAddr sdk.AccAddress // validator address to withdraw from
    withdrawAddr sdk.AccAddress // address to make the withdrawal to
}

func WithdrawValidatorRewardsAll(operatorAddr, withdrawAddr sdk.AccAddress)

    height = GetHeight()
    global = GetGlobal()
    pool = GetPool()
    ValInfo = GetValidatorDistInfo(delegation.ValidatorAddr)
    validator = GetValidator(delegation.ValidatorAddr)

    // withdraw self-delegation
    withdraw = GetDelegatorRewardsAll(validator.OperatorAddr, height)

    // withdrawal validator commission rewards
    global, commission = valInfo.WithdrawCommission(global, valInfo, height, pool.BondedTokens,
               validator.Tokens, validator.Commission)
    withdraw += commission
    SetGlobal(global)

    AddCoins(withdrawAddr, withdraw.TruncateDecimal())

*/
