package app

import (
	"database/sql"
	types2 "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/tendermint/tendermint/abci/types"
)

func (app *GaiaApp) BeginBlockHook(Database *sql.DB) types2.BeginBlocker {
	return func(ctx types2.Context, req types.RequestBeginBlock) types.ResponseBeginBlock {
		res := app.BeginBlocker(ctx, req)

		var (
			tx, _                  = Database.Begin()
			tx2, _                 = Database.Begin()
			tx3, _                 = Database.Begin()
			balanceStatement, _    = tx.Prepare("INSERT INTO balance (address,balance,denom,height,timestamp) VALUES (?,?,?,?,?)")
			rewardsStatement, _    = tx2.Prepare("INSERT INTO rewards (address,validator,rewards,denom,height,timestamp) VALUES (?,?,?,?,?,?)")
			valRewardsStatement, _ = tx3.Prepare("INSERT INTO val_rewards (validator,rewards,denom,height,timestamp) VALUES (?,?,?,?,?)")
		)
		defer balanceStatement.Close()
		defer rewardsStatement.Close()
		defer valRewardsStatement.Close()

		processAcc := func(account auth.Account) (bool) {
			balance := account.GetCoins()
			for _, coin := range balance {
				if _, err := balanceStatement.Exec(
					account.GetAddress().String(),
					uint64(coin.Amount.Int64()),
					coin.Denom,
					uint64(req.Header.Height),
					req.Header.Time,
				); err != nil {
					panic(err)
				}
			}
			wrap, _ := ctx.CacheContext()
			app.stakingKeeper.IterateDelegations(ctx, account.GetAddress(), func(index int64, del types2.Delegation) (stop bool) {
				val, _ := app.stakingKeeper.GetValidator(wrap, del.GetValidatorAddr())
				rew := app.distrKeeper.IncrementValidatorPeriod(wrap, val)
				rewards := app.distrKeeper.CalculateDelegationRewards(wrap, val, del, rew)

				for _, coin := range rewards {
					if _, err := rewardsStatement.Exec(
						account.GetAddress().String(),
						del.GetValidatorAddr().String(),
						uint64(coin.Amount.TruncateInt64()),
						coin.Denom,
						uint64(req.Header.Height),
						req.Header.Time,
					); err != nil {
						panic(err)
					}
				}
				return false
			})

			return false
		}

		app.accountKeeper.IterateAccounts(ctx, processAcc) // iterate over every account, every block :o

		vals := app.stakingKeeper.GetValidators(ctx, 500)

		for _, valObj := range vals {
			commission := app.distrKeeper.GetValidatorAccumulatedCommission(ctx, valObj.OperatorAddress)
			for _, coin := range commission {
				if _, err := valRewardsStatement.Exec(
					valObj.OperatorAddress.String(),
					uint64(coin.Amount.TruncateInt64()),
					coin.Denom,
					uint64(req.Header.Height),
					req.Header.Time,
				); err != nil {
					panic(err)
				}
			}
		}

		if err := tx.Commit(); err != nil {
			panic(err)
		}
		if err := tx2.Commit(); err != nil {
			panic(err)
		}
		if err := tx3.Commit(); err != nil {
			panic(err)
		}

		return res
	}
}
