package app

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/streadway/amqp"
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

	RabbitInsert struct {
		Values string `json:"values"`
		Table  string `json:"table"`
	}
)

var (
	BalanceTable       = "balance"
	BalanceFields      = "address,balance,denom,height,timestamp"
	RewardsTable       = "rewards"
	RewardsFields      = "address,validator,rewards,denom,height,timestamp"
	ValRewardsTable    = "val_rewards"
	ValRewardsFields   = "validator,rewards,denom,height,timestamp"
	DelegationsTable   = "delegations"
	UnbondingsTable    = "unbondings"
	DelegationsFields  = "address,validator,shares,height,timestamp"
	UnbondingsFields   = "address,validator,tokens,height,completion_timestamp,timestamp"
	MessagesFields     = "hash,idx,msgtype,msg,timestamp"
	MessagesTable      = "messages"
	TransactionsFields = "hash,height,code,gasWanted,gasUsed,log,memo,fees,tags,msgs,timestamp"
	TransactionsTable  = "transactions"
	AddressesFields    = "hash,idx,address"
	AddressesTable     = "message_addresses"
)

func (app *GaiaApp) BeginBlockHook(rabbit *amqp.Channel, blockerFunctions []func(*GaiaApp, *amqp.Channel, sdk.Context, types.RequestBeginBlock)) sdk.BeginBlocker {
	return func(ctx sdk.Context, req types.RequestBeginBlock) types.ResponseBeginBlock {
		res := app.BeginBlocker(ctx, req)
		// fucntions
		for _, fn := range blockerFunctions {
			fn(app, rabbit, ctx, req)
		}
		return res
	}
}

func BalancesBlocker(app *GaiaApp, rabbit *amqp.Channel, ctx sdk.Context, req types.RequestBeginBlock) {
	processAcc := func(account auth.Account) bool {
		balance := account.GetCoins()
		for _, coin := range balance {
			obj := RabbitInsert{
				Values: fmt.Sprintf("\"%s\",%d,\"%s\",%d,\"%s\"", account.GetAddress().String(), uint64(coin.Amount.Int64()), coin.Denom, uint64(req.Header.Height), req.Header.Time),
				Table:  BalanceTable,
			}
			obj.Insert(rabbit)
		}
		wrap, _ := ctx.CacheContext()
		app.stakingKeeper.IterateDelegations(wrap, account.GetAddress(), func(index int64, del sdk.Delegation) (stop bool) {
			val, _ := app.stakingKeeper.GetValidator(wrap, del.GetValidatorAddr())
			rew := app.distrKeeper.IncrementValidatorPeriod(wrap, val)
			rewards := app.distrKeeper.CalculateDelegationRewards(wrap, val, del, rew)

			for _, coin := range rewards {
				obj := RabbitInsert{
					Values: fmt.Sprintf("\"%s\",\"%s\",%d,\"%s\",%d,\"%s\"", account.GetAddress().String(), del.GetValidatorAddr().String(), uint64(coin.Amount.TruncateInt64()), coin.Denom, uint64(req.Header.Height), req.Header.Time),
					Table:  RewardsTable,
				}
				obj.Insert(rabbit)
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
			obj := RabbitInsert{
				Values: fmt.Sprintf("\"%s\",%d,\"%s\",%d,\"%s\"", valObj.OperatorAddress.String(), uint64(coin.Amount.TruncateInt64()), coin.Denom, uint64(req.Header.Height), req.Header.Time),
				Table:  ValRewardsTable,
			}
			obj.Insert(rabbit)
		}
	}
}

func DelegationsBlocker(app *GaiaApp, rabbit *amqp.Channel, ctx sdk.Context, req types.RequestBeginBlock) {

	delegations := app.stakingKeeper.GetAllDelegations(ctx)
	for _, delegation := range delegations {
		obj := RabbitInsert{
			Values: fmt.Sprintf("\"%s\",\"%s\",%d,%d,\"%s\"", delegation.GetDelegatorAddr().String(), delegation.GetValidatorAddr().String(), uint64(delegation.GetShares().TruncateInt64()), uint64(req.Header.Height), req.Header.Time),
			Table:  DelegationsTable,
		}
		obj.Insert(rabbit)
	}

	vals := app.stakingKeeper.GetValidators(ctx, 500)
	for _, valObj := range vals {
		unbondings := app.stakingKeeper.GetUnbondingDelegationsFromValidator(ctx, valObj.OperatorAddress)
		for _, unbond := range unbondings {
			for _, entry := range unbond.Entries {
				obj := RabbitInsert{
					Values: fmt.Sprintf("\"%s\",\"%s\",%d,%d,\"%s\",\"%s\"", unbond.DelegatorAddress.String(), unbond.ValidatorAddress.String(), uint64(entry.Balance.Int64()), uint64(req.Header.Height), entry.CompletionTime, req.Header.Time),
					Table:  UnbondingsTable,
				}
				obj.Insert(rabbit)
			}
		}
	}
}

func TxsBlockerForBlock(block tm.Block) func(*GaiaApp, *amqp.Channel, sdk.Context, types.RequestBeginBlock) {

	return func(app *GaiaApp, rabbit *amqp.Channel, ctx sdk.Context, req types.RequestBeginBlock) {

		for _, tx := range block.Data.Txs {
			txHash := hex.EncodeToString(tx.Hash())
			decoded, _ := app.BaseApp.GetTxDecoder()(tx)
			sdktx, ok := decoded.(auth.StdTx)
			if ok {
				for msgidx, msg := range sdktx.GetMsgs() {

					obj := RabbitInsert{
						Values: fmt.Sprintf("\"%s\",%d,%s,\"%s\",\"%s\"", txHash, msgidx, msg.Type(), string(msg.GetSignBytes()), block.Header.Time),
						Table:  MessagesTable,
					}
					obj.Insert(rabbit)

					fmt.Printf("Msg %d for %s\n", msgidx, txHash)

					fmt.Printf("Adding addresses for msg %d for %s\n", msgidx, txHash)
					addAddresses(msg, txHash, msgidx, rabbit)

				}
				fmt.Printf("EXECUTING TX %s FOR %d\n", txHash, block.Header.Height)
				result := app.BaseApp.DeliverTx(tx) // cause transaction to be applied to snapshotted db, so we can interrogate results.
				jsonTags, _ := app.GetCodec().MarshalJSON(sdk.TagsToStringTags(result.GetTags()))
				fmt.Println("DEBUG 1")
				jsonMsgs := MsgsToString(sdktx.GetMsgs())
				fmt.Println("DEBUG 2")
				jsonFee, _ := app.GetCodec().MarshalJSON(sdktx.Fee)

				obj := RabbitInsert{
					Values: fmt.Sprintf("\"%s\",%d,%d,%d,%d,\"%s\",\"%s\",\"%s\",\"%s\",\"%s\",\"%s\"",
						txHash,
						block.Header.Height,
						result.GetCode(),
						result.GetGasWanted(),
						result.GetGasUsed(),
						result.GetLog(),
						sdktx.GetMemo(),
						string(jsonFee),
						string(jsonTags),
						string(jsonMsgs),
						block.Header.Time),
					Table: TransactionsTable,
				}
				obj.Insert(rabbit)
			} else {
				fmt.Println("Assertion Error")
			}
		}
	}
}

func addAddresses(msg sdk.Msg, hash string, idx int, rabbit *amqp.Channel) {
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
			obj := RabbitInsert{
				Values: fmt.Sprintf("\"%s\",%d,\"%s\"", hash, idx, sdkAddr.String()),
				Table:  AddressesTable,
			}
			obj.Insert(rabbit)
		}
	}

}

func MsgsToString(msgs []sdk.Msg) string {
	outStrings := []string{}
	for _, msg := range msgs {
		outStrings = append(outStrings, string(msg.GetSignBytes()))
	}

	retval := fmt.Sprintf("[%s]", strings.Join(outStrings, ","))
	fmt.Sprintf("Messages: %s", retval)
	return retval
}

func (i RabbitInsert) Insert(c *amqp.Channel) {
	jsonString, err := json.Marshal(i)
	if err != nil {
		log.Fatal(err)
	}
	if err = c.Publish(
		"",        // exchange
		"inserts", // routing key
		false,     // mandatory
		false,     // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(jsonString),
		}); err != nil {
		log.Fatal(err)
	}
}
