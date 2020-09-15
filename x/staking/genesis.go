package staking

import (
	"encoding/json"
	"fmt"

	abci "github.com/tendermint/tendermint/abci/types"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankexported "github.com/cosmos/cosmos-sdk/x/bank/exported"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/cosmos/cosmos-sdk/x/staking/exported"
	"github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/x/staking/client/cli"
	flag "github.com/spf13/pflag"
	"github.com/tendermint/tendermint/crypto"
)

// InitGenesis sets the pool and parameters for the provided keeper.  For each
// validator in data, it sets that validator in the keeper along with manually
// setting the indexes. In addition, it also sets any delegations found in
// data. Finally, it updates the bonded validators.
// Returns final validator set after applying all declaration and delegations
func InitGenesis(
	ctx sdk.Context, keeper keeper.Keeper, accountKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper, data *types.GenesisState,
) (res []abci.ValidatorUpdate) {
	bondedTokens := sdk.ZeroInt()
	notBondedTokens := sdk.ZeroInt()

	// We need to pretend to be "n blocks before genesis", where "n" is the
	// validator update delay, so that e.g. slashing periods are correctly
	// initialized for the validator set e.g. with a one-block offset - the
	// first TM block is at height 1, so state updates applied from
	// genesis.json are in block 0.
	ctx = ctx.WithBlockHeight(1 - sdk.ValidatorUpdateDelay)

	keeper.SetParams(ctx, data.Params)
	keeper.SetLastTotalPower(ctx, data.LastTotalPower)

	for _, validator := range data.Validators {
		keeper.SetValidator(ctx, validator)

		// Manually set indices for the first time
		keeper.SetValidatorByConsAddr(ctx, validator)
		keeper.SetValidatorByPowerIndex(ctx, validator)

		// Call the creation hook if not exported
		if !data.Exported {
			keeper.AfterValidatorCreated(ctx, validator.OperatorAddress)
		}

		// update timeslice if necessary
		if validator.IsUnbonding() {
			keeper.InsertUnbondingValidatorQueue(ctx, validator)
		}

		switch validator.GetStatus() {
		case sdk.Bonded:
			bondedTokens = bondedTokens.Add(validator.GetTokens())
		case sdk.Unbonding, sdk.Unbonded:
			notBondedTokens = notBondedTokens.Add(validator.GetTokens())
		default:
			panic("invalid validator status")
		}
	}

	for _, delegation := range data.Delegations {
		// Call the before-creation hook if not exported
		if !data.Exported {
			keeper.BeforeDelegationCreated(ctx, delegation.DelegatorAddress, delegation.ValidatorAddress)
		}

		keeper.SetDelegation(ctx, delegation)
		// Call the after-modification hook if not exported
		if !data.Exported {
			keeper.AfterDelegationModified(ctx, delegation.DelegatorAddress, delegation.ValidatorAddress)
		}
	}

	for _, ubd := range data.UnbondingDelegations {
		keeper.SetUnbondingDelegation(ctx, ubd)

		for _, entry := range ubd.Entries {
			keeper.InsertUBDQueue(ctx, ubd, entry.CompletionTime)
			notBondedTokens = notBondedTokens.Add(entry.Balance)
		}
	}

	for _, red := range data.Redelegations {
		keeper.SetRedelegation(ctx, red)

		for _, entry := range red.Entries {
			keeper.InsertRedelegationQueue(ctx, red, entry.CompletionTime)
		}
	}

	bondedCoins := sdk.NewCoins(sdk.NewCoin(data.Params.BondDenom, bondedTokens))
	notBondedCoins := sdk.NewCoins(sdk.NewCoin(data.Params.BondDenom, notBondedTokens))

	// check if the unbonded and bonded pools accounts exists
	bondedPool := keeper.GetBondedPool(ctx)
	if bondedPool == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.BondedPoolName))
	}

	// TODO remove with genesis 2-phases refactor https://github.com/cosmos/cosmos-sdk/issues/2862
	// add coins if not provided on genesis
	if bankKeeper.GetAllBalances(ctx, bondedPool.GetAddress()).IsZero() {
		if err := bankKeeper.SetBalances(ctx, bondedPool.GetAddress(), bondedCoins); err != nil {
			panic(err)
		}

		accountKeeper.SetModuleAccount(ctx, bondedPool)
	}

	notBondedPool := keeper.GetNotBondedPool(ctx)
	if notBondedPool == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.NotBondedPoolName))
	}

	if bankKeeper.GetAllBalances(ctx, notBondedPool.GetAddress()).IsZero() {
		if err := bankKeeper.SetBalances(ctx, notBondedPool.GetAddress(), notBondedCoins); err != nil {
			panic(err)
		}

		accountKeeper.SetModuleAccount(ctx, notBondedPool)
	}

	// don't need to run Tendermint updates if we exported
	if data.Exported {
		for _, lv := range data.LastValidatorPowers {
			keeper.SetLastValidatorPower(ctx, lv.Address, lv.Power)
			validator, found := keeper.GetValidator(ctx, lv.Address)

			if !found {
				panic(fmt.Sprintf("validator %s not found", lv.Address))
			}

			update := validator.ABCIValidatorUpdate()
			update.Power = lv.Power // keep the next-val-set offset, use the last power for the first block
			res = append(res, update)
		}
	} else {
		res = keeper.ApplyAndReturnValidatorSetUpdates(ctx)
	}

	return res
}

// ExportGenesis returns a GenesisState for a given context and keeper. The
// GenesisState will contain the pool, params, validators, and bonds found in
// the keeper.
func ExportGenesis(ctx sdk.Context, keeper keeper.Keeper) *types.GenesisState {
	var unbondingDelegations []types.UnbondingDelegation

	keeper.IterateUnbondingDelegations(ctx, func(_ int64, ubd types.UnbondingDelegation) (stop bool) {
		unbondingDelegations = append(unbondingDelegations, ubd)
		return false
	})

	var redelegations []types.Redelegation

	keeper.IterateRedelegations(ctx, func(_ int64, red types.Redelegation) (stop bool) {
		redelegations = append(redelegations, red)
		return false
	})

	var lastValidatorPowers []types.LastValidatorPower

	keeper.IterateLastValidatorPowers(ctx, func(addr sdk.ValAddress, power int64) (stop bool) {
		lastValidatorPowers = append(lastValidatorPowers, types.LastValidatorPower{Address: addr, Power: power})
		return false
	})

	return &types.GenesisState{
		Params:               keeper.GetParams(ctx),
		LastTotalPower:       keeper.GetLastTotalPower(ctx),
		LastValidatorPowers:  lastValidatorPowers,
		Validators:           keeper.GetAllValidators(ctx),
		Delegations:          keeper.GetAllDelegations(ctx),
		UnbondingDelegations: unbondingDelegations,
		Redelegations:        redelegations,
		Exported:             true,
	}
}

// WriteValidators returns a slice of bonded genesis validators.
func WriteValidators(ctx sdk.Context, keeper keeper.Keeper) (vals []tmtypes.GenesisValidator) {
	keeper.IterateLastValidators(ctx, func(_ int64, validator exported.ValidatorI) (stop bool) {
		vals = append(vals, tmtypes.GenesisValidator{
			Address: validator.GetConsAddr().Bytes(),
			PubKey:  validator.GetConsPubKey(),
			Power:   validator.GetConsensusPower(),
			Name:    validator.GetMoniker(),
		})

		return false
	})

	return
}

// ValidateGenesis validates the provided staking genesis state to ensure the
// expected invariants holds. (i.e. params in correct bounds, no duplicate validators)
func ValidateGenesis(data *types.GenesisState) error {
	if err := validateGenesisStateValidators(data.Validators); err != nil {
		return err
	}

	return data.Params.Validate()
}

func validateGenesisStateValidators(validators []types.Validator) (err error) {
	addrMap := make(map[string]bool, len(validators))

	for i := 0; i < len(validators); i++ {
		val := validators[i]
		strKey := string(val.GetConsPubKey().Bytes())

		if _, ok := addrMap[strKey]; ok {
			return fmt.Errorf("duplicate validator in genesis state: moniker %v, address %v", val.Description.Moniker, val.GetConsAddr())
		}

		if val.Jailed && val.IsBonded() {
			return fmt.Errorf("validator is bonded and jailed in genesis state: moniker %v, address %v", val.Description.Moniker, val.GetConsAddr())
		}

		if val.DelegatorShares.IsZero() && !val.IsUnbonding() {
			return fmt.Errorf("bonded/unbonded genesis validator cannot have zero delegator shares, validator: %v", val)
		}

		addrMap[strKey] = true
	}

	return
}

// ValidateAccountParamsnGenesis checks that the provided account has a sufficient
// balance in the set of genesis accounts. Used in gentx
func ValidateAccountParamsInGenesis(
	appGenesisState map[string]json.RawMessage, genBalIterator genutiltypes.GenesisBalancesIterator,
	addr sdk.Address, coins sdk.Coins, cdc codec.JSONMarshaler,
) error {
	genesisStakingData := appGenesisState[types.ModuleName]
	var stakingData types.GenesisState
	cdc.MustUnmarshalJSON(genesisStakingData, &stakingData)
	bondDenom := stakingData.Params.BondDenom

	var err error

	accountIsInGenesis := false

	genBalIterator.IterateGenesisBalances(cdc, appGenesisState,
		func(bal bankexported.GenesisBalance) (stop bool) {
			accAddress := bal.GetAddress()
			accCoins := bal.GetCoins()

			// ensure that account is in genesis
			if accAddress.Equals(addr) {
				// ensure account contains enough funds of default bond denom
				if coins.AmountOf(bondDenom).GT(accCoins.AmountOf(bondDenom)) {
					err = fmt.Errorf(
						"account %s has a balance in genesis, but it only has %v%s available to stake, not %v%s",
						addr, accCoins.AmountOf(bondDenom), bondDenom, coins.AmountOf(bondDenom), bondDenom,
					)

					return true
				}

				accountIsInGenesis = true
				return true
			}

			return false
		},
	)

	if err != nil {
		return err
	}

	if !accountIsInGenesis {
		return fmt.Errorf("account %s does not have a balance in the genesis state", addr)
	}
	return nil
}

// ValidateMsgInGenesis is used in collect-gentx to verify a genesis message
func ValidateMsgInGenesis(msg sdk.Msg, balancesMap map[string]bankexported.GenesisBalance, appGenTxs []sdk.Tx,
	persistentPeers string, addressesIPs []string, nodeAddrIP string, moniker string,
) (genTxs []sdk.Tx, perPeers string, ips []string, err error) {
	createValMsg := msg.(*types.MsgCreateValidator)

	// validate delegator and validator addresses and funds against the accounts in the state
	delAddr := createValMsg.DelegatorAddress.String()
	valAddr := sdk.AccAddress(createValMsg.ValidatorAddress).String()

	delBal, delOk := balancesMap[delAddr]
	if !delOk {
		return appGenTxs, persistentPeers, addressesIPs, fmt.Errorf("account %s balance not in genesis state: %+v", delAddr, balancesMap)
	}

	_, valOk := balancesMap[valAddr]
	if !valOk {
		return appGenTxs, persistentPeers, addressesIPs, fmt.Errorf("account %s balance not in genesis state: %+v", valAddr, balancesMap)
	}

	if delBal.GetCoins().AmountOf(createValMsg.Value.Denom).LT(createValMsg.Value.Amount) {
		return appGenTxs, persistentPeers, addressesIPs, fmt.Errorf(
			"insufficient fund for delegation %v: %v < %v",
			delBal.GetAddress().String(), delBal.GetCoins().AmountOf(createValMsg.Value.Denom), createValMsg.Value.Amount,
		)
	}

	// exclude itself from persistent peers
	if createValMsg.Description.Moniker != moniker {
		addressesIPs = append(addressesIPs, nodeAddrIP)
	}

	return appGenTxs, persistentPeers, addressesIPs, nil
}

/////////////////////////////
// Genesis Message Helpers //
/////////////////////////////

type ValidatorMsgBuildingHelpers struct{}

// CreateValidatorMsgHelpers - used for gentx
func (ValidatorMsgBuildingHelpers) CreateValidatorMsgFlagSet(ipDefault string) (fs *flag.FlagSet, defaultsDesc string) {
	return cli.CreateValidatorMsgFlagSet(ipDefault)
}

// PrepareFlagsForTxCreateValidator - used for gentx
func (ValidatorMsgBuildingHelpers) PrepareConfigForTxCreateValidator(flagSet *flag.FlagSet, moniker, nodeID,
	chainID string, valPubKey crypto.PubKey) (interface{}, error) {
	config, err := cli.PrepareConfigForTxCreateValidator(flagSet, moniker, nodeID, chainID, valPubKey)
	return config, err
}

// BuildCreateValidatorMsg - used for gentx
func (ValidatorMsgBuildingHelpers) BuildCreateValidatorMsg(cliCtx client.Context, config interface{}, txBldr tx.Factory,
	generateOnly bool) (tx.Factory, sdk.Msg, error) {
	return cli.BuildCreateValidatorMsg(cliCtx, config.(cli.TxCreateValidatorConfig), txBldr, generateOnly)
}

// ValidateAccountInGenesis - used for gentx
func (ValidatorMsgBuildingHelpers) ValidateAccountInGenesis(
	appGenesisState map[string]json.RawMessage, genBalIterator genutiltypes.GenesisBalancesIterator,
	addr sdk.Address, coins sdk.Coins, cdc codec.JSONMarshaler,
) error {
	return ValidateAccountParamsInGenesis(appGenesisState, genBalIterator, addr, coins, cdc)
}
