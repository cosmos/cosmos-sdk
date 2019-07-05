// DONTCOVER
// nolint
package v0_36

import (
	"github.com/tendermint/tendermint/crypto"
	sdk "github.com/cosmos/cosmos-sdk/types"
	v034accounts "github.com/cosmos/cosmos-sdk/x/genaccounts/legacy/v0_34"
	v034staking "github.com/cosmos/cosmos-sdk/x/staking/legacy/v0_34"
	v034distr "github.com/cosmos/cosmos-sdk/x/distribution/legacy/v0_34"
)

const (
	notBondedPoolName = "NotBondedTokensPool"
	bondedPoolName    = "BondedTokensPool"
	feeCollectorName 	= "FeeCollector"
	mintModuleName 		= "mint"
)

// Migrate accepts exported genesis state from v0.34 and migrates it to v0.36
// genesis state. All entries are identical except for validator slashing events
// which now include the period.
func Migrate(oldGenState v034accounts.GenesisState, fees sdk.Coins, communityPool sdk.DecCoins,
	vals v034staking.Validators, ubds[]v034staking.UnbondingDelegation, valOutRewards []v034distr.ValidatorOutstandingRewardsRecord,
	bondDenom, distrModuleName, govModuleName string) GenesisState {
	depositedCoinsAccAddr     := sdk.AccAddress(crypto.AddressHash([]byte("govDepositedCoins")))	
	burnedDepositCoinsAccAddr := sdk.AccAddress(crypto.AddressHash([]byte("govBurnedDepositCoins")))

	bondedAmt := sdk.ZeroInt()
	notBondedAmt := sdk.ZeroInt()

	var govCoins sdk.Coins
	newGenState := make(GenesisState, len(oldGenState) + 4) // - 2 gov baseAccs + 6 moduleAccs
	
	for _, acc := range oldGenState {
		switch {
		case acc.Address.Equals(depositedCoinsAccAddr): 
			govCoins = acc.Coins
		case acc.Address.Equals(burnedDepositCoinsAccAddr):
			// do nothing
		default: 
		genAcc := NewGenesisAccount(acc.Address, acc.Coins, acc.Sequence, 
			acc.OriginalVesting, acc.DelegatedFree, acc.DelegatedVesting,
			acc.StartTime, acc.EndTime, "", "")
		newGenState = append(newGenState, genAcc)
		}
	}

	// get staking module accounts coins
	for _, validator := range vals {
		switch validator.Status {
		case sdk.Bonded:
			bondedAmt = bondedAmt.Add(validator.Tokens)
		case sdk.Unbonding, sdk.Unbonded:
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
		distrDecCoins = distrDecCoins.Add(reward.OutstandingRewards)
	}

	distrCoins, _ := distrDecCoins.Add(communityPool).TruncateDecimal()

	// get module account addresses
	feeCollectorAddr := sdk.AccAddress(crypto.AddressHash([]byte(feeCollectorName)))
	govAddr := sdk.AccAddress(crypto.AddressHash([]byte(govModuleName)))
	bondedAddr := sdk.AccAddress(crypto.AddressHash([]byte(bondedPoolName)))
	notBondedAddr := sdk.AccAddress(crypto.AddressHash([]byte(notBondedPoolName)))
	distrAddr := sdk.AccAddress(crypto.AddressHash([]byte(distrModuleName)))
	mintAddr := sdk.AccAddress(crypto.AddressHash([]byte(mintModuleName)))

	// create module genaccs
	feeCollectorModuleAcc := NewGenesisAccount(feeCollectorAddr, fees, 0,
		sdk.Coins{}, sdk.Coins{}, sdk.Coins{}, 0, 0, feeCollectorName, "basic")
	govModuleAcc := NewGenesisAccount(govAddr, govCoins,0,
		sdk.Coins{}, sdk.Coins{}, sdk.Coins{}, 0, 0, govModuleName, "burner")
	distrModuleAcc := NewGenesisAccount(distrAddr, distrCoins,0,
		sdk.Coins{}, sdk.Coins{}, sdk.Coins{}, 0, 0, distrModuleName, "basic")
	bondedModuleAcc := NewGenesisAccount(bondedAddr, bondedCoins,0,
		sdk.Coins{}, sdk.Coins{}, sdk.Coins{}, 0, 0, bondedPoolName, "burner")
	notBondedModuleAcc := NewGenesisAccount(notBondedAddr, notBondedCoins, 0,
		sdk.Coins{}, sdk.Coins{}, sdk.Coins{}, 0, 0, notBondedPoolName, "burner")
	mintModuleAcc := NewGenesisAccount(mintAddr, sdk.Coins{}, 0,
		sdk.Coins{}, sdk.Coins{}, sdk.Coins{}, 0, 0, mintModuleName, "minter")


	newGenState = append(newGenState, []GenesisAccount{
		feeCollectorModuleAcc, govModuleAcc, distrModuleAcc, bondedModuleAcc, notBondedModuleAcc, mintModuleAcc,
	}...)



	return newGenState
}