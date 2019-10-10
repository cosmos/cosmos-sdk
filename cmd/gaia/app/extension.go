package app

import (
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/tendermint/tendermint/abci/types"
	tm "github.com/tendermint/tendermint/types"
)

type (
	// AddressContainingStruct exists to parse JSON blobs containing sdk.Addressbased fields.
	AddressContainingStruct struct {
		From         sdk.AccAddress `json:"from_address"`
		To           sdk.AccAddress `json:"to_address"`
		Validator    sdk.ValAddress `json:"validator_address"`
		Delegator    sdk.AccAddress `json:"delegator_address"`
		SrcValidator sdk.ValAddress `json:"src_validator_address"`
		DstValidator sdk.ValAddress `json:"dst_validator_address"`
		Proposer     sdk.AccAddress `json:"proposer"`
	}

	// BasicMsgStruct is a simplified reprentation of an sdk.Msg
	BasicMsgStruct struct {
		Type  string                  `json:"type"`
		Value AddressContainingStruct `json:"value"`
	}
)

func (app *GaiaApp) BeginBlockHook(Database *sql.DB, blockerFunctions []func(*GaiaApp, *sql.DB, sdk.Context, types.RequestBeginBlock)) sdk.BeginBlocker {
	return func(ctx sdk.Context, req types.RequestBeginBlock) types.ResponseBeginBlock {
		res := app.BeginBlocker(ctx, req)
		// fucntions
		for _, fn := range blockerFunctions {
			fn(app, Database, ctx, req)
		}
		return res
	}
}

func BalancesBlocker(app *GaiaApp, Database *sql.DB, ctx sdk.Context, req types.RequestBeginBlock) {
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

	processAcc := func(account auth.Account) bool {
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
		app.stakingKeeper.IterateDelegations(wrap, account.GetAddress(), func(index int64, del sdk.Delegation) (stop bool) {
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

	wrap, _ := ctx.CacheContext()
	vals := app.stakingKeeper.GetValidators(wrap, 500)
	for _, valObj := range vals {
		commission := app.distrKeeper.GetValidatorAccumulatedCommission(wrap, valObj.OperatorAddress)
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
}

func DelegationsBlocker(app *GaiaApp, Database *sql.DB, ctx sdk.Context, req types.RequestBeginBlock) {
	var (
		tx, _                   = Database.Begin()
		tx2, _                  = Database.Begin()
		delegationsStatement, _ = tx.Prepare("INSERT INTO delegations (address,validator,shares,height,timestamp) VALUES (?,?,?,?,?)")
		unbondingsStatement, _  = tx2.Prepare("INSERT INTO unbondings (address,validator,tokens,height,completion_timestamp,timestamp) VALUES (?,?,?,?,?,?)")
	)

	defer delegationsStatement.Close()
	defer unbondingsStatement.Close()

	delegations := app.stakingKeeper.GetAllDelegations(ctx)
	for _, delegation := range delegations {
		if _, err := delegationsStatement.Exec(
			delegation.GetDelegatorAddr().String(),
			delegation.GetValidatorAddr().String(),
			uint64(delegation.GetShares().TruncateInt64()),
			uint64(req.Header.Height),
			req.Header.Time,
		); err != nil {
			panic(err)
		}
	}

	vals := app.stakingKeeper.GetValidators(ctx, 500)
	for _, valObj := range vals {
		unbondings := app.stakingKeeper.GetUnbondingDelegationsFromValidator(ctx, valObj.OperatorAddress)
		for _, unbond := range unbondings {
			for _, entry := range unbond.Entries {
				if _, err := unbondingsStatement.Exec(
					unbond.DelegatorAddress.String(),
					unbond.ValidatorAddress.String(),
					uint64(entry.Balance.Int64()),
					uint64(req.Header.Height),
					entry.CompletionTime,
					req.Header.Time,
				); err != nil {
					fmt.Printf("%v", err)
				}
			}
		}
	}

	if err := tx.Commit(); err != nil {
		panic(err)
	}

	if err := tx2.Commit(); err != nil {
		panic(err)
	}
}

func TxsBlockerForBlock(block tm.Block) func(*GaiaApp, *sql.DB, sdk.Context, types.RequestBeginBlock) {

	return func(app *GaiaApp, Database *sql.DB, ctx sdk.Context, req types.RequestBeginBlock) {
		var (
			dbtx, _                 = Database.Begin()
			dbtx2, _                = Database.Begin()
			dbtx3, _                = Database.Begin()
			messagesStatement, _    = dbtx.Prepare("INSERT INTO messages (hash,idx,msgtype,msg,timestamp) VALUES (?,?,?,?,?)")
			addressesStatement, _   = dbtx3.Prepare("INSERT INTO message_addresses (hash,idx,address) VALUES (?,?,?)")
			transactionStatement, _ = dbtx2.Prepare("INSERT INTO transactions (hash,height,code,gasWanted,gasUsed,log,memo,fees,tags,timestamp) VALUES (?,?,?,?,?,?,?,?,?,?)")
		)
		defer messagesStatement.Close()
		defer transactionStatement.Close()
		defer addressesStatement.Close()

		for _, tx := range block.Data.Txs {
			txHash := hex.EncodeToString(tx.Hash())
			decoded, _ := app.BaseApp.GetTxDecoder()(tx)
			sdktx, ok := decoded.(auth.StdTx)
			if ok {
				for msgidx, msg := range sdktx.GetMsgs() {
					fmt.Printf("Msg %d for %s", msgidx, txHash)
					if _, err := messagesStatement.Exec(
						txHash,
						msgidx,
						msg.Type(),
						string(msg.GetSignBytes()),
						block.Header.Time,
					); err != nil {
						panic(err)
					}
					fmt.Printf("Adding addresses for msg %d for %s", msgidx, txHash)
					addAddresses(msg, txHash, msgidx, addressesStatement)

				}
				fmt.Printf("EXECUTING TX %s FOR %d\n", txHash, block.Header.Height)
				result := app.BaseApp.DeliverTx(tx) // cause transaction to be applied to snapshotted db, so we can interrogate results.
				jsonTags, _ := app.GetCodec().MarshalJSON(sdk.TagsToStringTags(result.GetTags()))
				jsonFee, _ := app.GetCodec().MarshalJSON(sdktx.Fee)

				if _, err := transactionStatement.Exec(
					txHash,
					block.Header.Height,
					result.GetCode(),
					result.GetGasWanted(),
					result.GetGasUsed(),
					result.GetLog(),
					sdktx.GetMemo(),
					string(jsonFee),
					string(jsonTags),
					block.Header.Time,
				); err != nil {
					panic(err)
				}
			} else {
				fmt.Println("Assertion Error")
			}
		}
		if err := dbtx.Commit(); err != nil {
			panic(err)
		}
		if err := dbtx2.Commit(); err != nil {
			panic(err)
		}
		if err := dbtx3.Commit(); err != nil {
			panic(err)
		}
	}
}

func addAddresses(msg sdk.Msg, hash string, idx int, stmt *sql.Stmt) {
	// get addresses
	m := BasicMsgStruct{}
	a := make(map[string]bool)

	_ = json.Unmarshal(msg.GetSignBytes(), &m)
	ref := reflect.ValueOf(&m.Value).Elem()
	for i := 0; i < ref.NumField(); i++ {
		addr := ref.Field(i).Interface()
		sdkAddr, ok := addr.(sdk.Address)                   // cast to address interface so we have access to the String() method, which bech32ifies the address
		if ok && !sdkAddr.Empty() && !a[sdkAddr.String()] { // pks in clickhouse aren't unique, so avoid dedupe here.
			a[sdkAddr.String()] = true
			if _, err := stmt.Exec(
				hash,
				idx,
				sdkAddr.String(),
			); err != nil {
				panic(err)
			}
		}
	}

}
