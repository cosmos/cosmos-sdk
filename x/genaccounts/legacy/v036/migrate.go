// DONTCOVER
// nolint
package v036

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	v034distr "github.com/cosmos/cosmos-sdk/x/distribution/legacy/v034"
	v034accounts "github.com/cosmos/cosmos-sdk/x/genaccounts/legacy/v034"
	v034gov "github.com/cosmos/cosmos-sdk/x/gov/legacy/v034"
	v034staking "github.com/cosmos/cosmos-sdk/x/staking/legacy/v034"

	"github.com/tendermint/tendermint/crypto"
)

const (
	notBondedPoolName = "not_bonded_tokens_pool"
	bondedPoolName    = "bonded_tokens_pool"
	feeCollectorName  = "fee_collector"
	mintModuleName    = "mint"

	basic   = "basic"
	minter  = "minter"
	burner  = "burner"
	staking = "staking"
)

// Migrate accepts exported genesis state from v0.34 and migrates it to v0.36
// genesis state. It deletes the governance base accounts and creates the new module accounts.
// The remaining accounts are updated to the new GenesisAccount type from 0.36
func Migrate(
	oldGenState v034accounts.GenesisState, fees sdk.Coins, communityPool sdk.DecCoins,
	deposits []v034gov.DepositWithMetadata, vals v034staking.Validators, ubds []v034staking.UnbondingDelegation,
	valOutRewards []v034distr.ValidatorOutstandingRewardsRecord, bondDenom, distrModuleName, govModuleName string,
) GenesisState {

	depositedCoinsAccAddr := sdk.AccAddress(crypto.AddressHash([]byte("govDepositedCoins")))
	burnedDepositCoinsAccAddr := sdk.AccAddress(crypto.AddressHash([]byte("govBurnedDepositCoins")))

	bondedAmt := sdk.ZeroInt()
	notBondedAmt := sdk.ZeroInt()

	// remove the two previous governance base accounts for deposits and burned
	// coins from rejected proposals add six new module accounts:
	// distribution, gov, mint, fee collector, bonded and not bonded pool
	var (
		newGenState   GenesisState
		govCoins      sdk.Coins
		extraAccounts = 6
	)

	for _, acc := range oldGenState {
		switch {
		case acc.Address.Equals(depositedCoinsAccAddr):
			// remove gov deposits base account
			govCoins = acc.Coins
			extraAccounts -= 1

		case acc.Address.Equals(burnedDepositCoinsAccAddr):
			// remove gov burned deposits base account
			extraAccounts -= 1

		default:
			newGenState = append(
				newGenState,
				NewGenesisAccount(
					acc.Address, acc.Coins, acc.Sequence,
					acc.OriginalVesting, acc.DelegatedFree, acc.DelegatedVesting,
					acc.StartTime, acc.EndTime, "", []string{},
				),
			)
		}
	}

	var expDeposits sdk.Coins
	for _, deposit := range deposits {
		expDeposits = expDeposits.Add(deposit.Deposit.Amount...)
	}

	if !expDeposits.IsEqual(govCoins) {
		panic(
			fmt.Sprintf(
				"pre migration deposit base account coins ≠ stored deposits coins (%s ≠ %s)",
				expDeposits.String(), govCoins.String(),
			),
		)
	}

	// get staking module accounts coins
	for _, validator := range vals {
		switch validator.Status {
		case v034staking.Bonded:
			bondedAmt = bondedAmt.Add(validator.Tokens)

		case v034staking.Unbonding, v034staking.Unbonded:
			notBondedAmt = notBondedAmt.Add(validator.Tokens)

		default:
			panic("invalid validator status")
		}
	}

	for _, ubd := range ubds {
		for _, entry := range ubd.Entries {
			notBondedAmt = notBondedAmt.Add(entry.Balance)
		}
	}

	bondedCoins := sdk.NewCoins(sdk.NewCoin(bondDenom, bondedAmt))
	notBondedCoins := sdk.NewCoins(sdk.NewCoin(bondDenom, notBondedAmt))

	// get distr module account coins
	var distrDecCoins sdk.DecCoins
	for _, reward := range valOutRewards {
		distrDecCoins = distrDecCoins.Add(reward.OutstandingRewards...)
	}

	distrCoins, _ := distrDecCoins.Add(communityPool...).TruncateDecimal()

	// get module account addresses
	feeCollectorAddr := sdk.AccAddress(crypto.AddressHash([]byte(feeCollectorName)))
	govAddr := sdk.AccAddress(crypto.AddressHash([]byte(govModuleName)))
	bondedAddr := sdk.AccAddress(crypto.AddressHash([]byte(bondedPoolName)))
	notBondedAddr := sdk.AccAddress(crypto.AddressHash([]byte(notBondedPoolName)))
	distrAddr := sdk.AccAddress(crypto.AddressHash([]byte(distrModuleName)))
	mintAddr := sdk.AccAddress(crypto.AddressHash([]byte(mintModuleName)))

	// create module genesis accounts
	feeCollectorModuleAcc := NewGenesisAccount(
		feeCollectorAddr, fees, 0,
		sdk.Coins{}, sdk.Coins{}, sdk.Coins{},
		0, 0, feeCollectorName, []string{basic},
	)
	govModuleAcc := NewGenesisAccount(
		govAddr, govCoins, 0,
		sdk.Coins{}, sdk.Coins{}, sdk.Coins{},
		0, 0, govModuleName, []string{burner},
	)
	distrModuleAcc := NewGenesisAccount(
		distrAddr, distrCoins, 0,
		sdk.Coins{}, sdk.Coins{}, sdk.Coins{},
		0, 0, distrModuleName, []string{basic},
	)
	bondedModuleAcc := NewGenesisAccount(
		bondedAddr, bondedCoins, 0,
		sdk.Coins{}, sdk.Coins{}, sdk.Coins{},
		0, 0, bondedPoolName, []string{burner, staking},
	)
	notBondedModuleAcc := NewGenesisAccount(
		notBondedAddr, notBondedCoins, 0,
		sdk.Coins{}, sdk.Coins{}, sdk.Coins{},
		0, 0, notBondedPoolName, []string{burner, staking},
	)
	mintModuleAcc := NewGenesisAccount(
		mintAddr, sdk.Coins{}, 0,
		sdk.Coins{}, sdk.Coins{}, sdk.Coins{},
		0, 0, mintModuleName, []string{minter},
	)

	newGenState = append(
		newGenState,
		[]GenesisAccount{
			feeCollectorModuleAcc, govModuleAcc, distrModuleAcc,
			bondedModuleAcc, notBondedModuleAcc, mintModuleAcc,
		}...,
	)

	// verify the total number of accounts is correct
	if len(newGenState) != len(oldGenState)+extraAccounts {
		panic(
			fmt.Sprintf(
				"invalid total number of genesis accounts; got: %d, expected: %d",
				len(newGenState), len(oldGenState)+extraAccounts),
		)
	}

	return newGenState
}
