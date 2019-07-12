package v0_36

import (
	"github.com/cosmos/cosmos-sdk/types"
	v034distr "github.com/cosmos/cosmos-sdk/x/distribution/legacy/v0_34"
	v034accounts "github.com/cosmos/cosmos-sdk/x/genaccounts/legacy/v0_34"
	v034gov "github.com/cosmos/cosmos-sdk/x/gov/legacy/v0_34"
	v034staking "github.com/cosmos/cosmos-sdk/x/staking/legacy/v0_34"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	priv                      = secp256k1.GenPrivKey()
	addr                      = types.AccAddress(priv.PubKey().Address())
	depositedCoinsAccAddr     = types.AccAddress(crypto.AddressHash([]byte("govDepositedCoins")))
	burnedDepositCoinsAccAddr = types.AccAddress(crypto.AddressHash([]byte("govBurnedDepositCoins")))

	coins     = types.Coins{types.NewInt64Coin(types.DefaultBondDenom, 10)}
	halfCoins = types.Coins{types.NewInt64Coin(types.DefaultBondDenom, 5)}

	accountDeposited = v034accounts.GenesisAccount{
		Address:       depositedCoinsAccAddr,
		Coins:         coins,
		Sequence:      1,
		AccountNumber: 1,

		OriginalVesting:  coins,
		DelegatedFree:    coins,
		DelegatedVesting: coins,
		StartTime:        0,
		EndTime:          0,
	}

	accountBurned = v034accounts.GenesisAccount{
		Address:       burnedDepositCoinsAccAddr,
		Coins:         coins,
		Sequence:      2,
		AccountNumber: 2,

		OriginalVesting:  coins,
		DelegatedFree:    coins,
		DelegatedVesting: coins,
		StartTime:        0,
		EndTime:          0,
	}

	deposit = v034gov.DepositWithMetadata{
		ProposalID: 1,
		Deposit: v034gov.Deposit{
			ProposalID: 1,
			Depositor:  addr,
			Amount:     coins,
		},
	}
)

func TestMigrateEmptyRecord(t *testing.T) {

	type args struct {
		accounts v034accounts.GenesisState
		deposits []v034gov.DepositWithMetadata
	}
	tests := []struct {
		name string
		args args
	}{
		{"No Accounts", args{v034accounts.GenesisState{}, []v034gov.DepositWithMetadata{}}},
		{"Deposited account", args{v034accounts.GenesisState{accountDeposited}, []v034gov.DepositWithMetadata{deposit}}},
		{"Burned account", args{v034accounts.GenesisState{accountBurned}, []v034gov.DepositWithMetadata{}}},
		{"Burned and deposited accounts", args{v034accounts.GenesisState{accountDeposited, accountBurned}, []v034gov.DepositWithMetadata{deposit}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NotPanics(t, func() {
				Migrate(
					tt.args.accounts,
					types.Coins{},
					types.DecCoins{},
					tt.args.deposits,
					v034staking.Validators{},
					[]v034staking.UnbondingDelegation{},
					[]v034distr.ValidatorOutstandingRewardsRecord{},
					types.DefaultBondDenom,
					v034distr.ModuleName,
					v034gov.ModuleName,
				)
			})
		})
	}
}

func TestMigrateWrongDeposit(t *testing.T) {
	require.Panics(t, func() {
		Migrate(
			v034accounts.GenesisState{
				accountDeposited,
				accountBurned,
			},
			types.Coins{},
			types.DecCoins{},
			[]v034gov.DepositWithMetadata{
				{
					ProposalID: 1,
					Deposit: v034gov.Deposit{
						ProposalID: 1,
						Depositor:  addr,
						Amount:     halfCoins,
					},
				},
			},
			v034staking.Validators{},
			[]v034staking.UnbondingDelegation{},
			[]v034distr.ValidatorOutstandingRewardsRecord{},
			types.DefaultBondDenom,
			v034distr.ModuleName,
			v034gov.ModuleName,
		)
	})
}
