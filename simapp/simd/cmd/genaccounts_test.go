package cmd_test

import (
	"context"
	"fmt"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/simapp"
	simcmd "github.com/cosmos/cosmos-sdk/simapp/simd/cmd"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltest "github.com/cosmos/cosmos-sdk/x/genutil/client/testutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
)

var testMbm = module.NewBasicManager(genutil.AppModuleBasic{})

func TestAddGenesisAccountCmd(t *testing.T) {
	_, _, addr1 := testdata.KeyTestPubAddr()
	tests := []struct {
		name                  string
		addr                  string
		denom                 string
		vestingStartTime      uint64
		vestingEndTime        uint64
		vestingAmount         string
		vestingPeriodsNumber  uint64
		vestingPeriodsAmounts string

		withKeyring bool
		expectErr   bool
		want        authtypes.GenesisAccounts
	}{
		{
			name:        "invalid address",
			addr:        "",
			denom:       "1000atom",
			withKeyring: false,
			expectErr:   true,
		},
		{
			name:        "valid address",
			addr:        addr1.String(),
			denom:       "1000atom",
			withKeyring: false,
			expectErr:   false,
		},
		{
			name:        "multiple denoms",
			addr:        addr1.String(),
			denom:       "1000atom, 2000stake",
			withKeyring: false,
			expectErr:   false,
		},
		{
			name:        "with keyring",
			addr:        "ser",
			denom:       "1000atom",
			withKeyring: true,
			expectErr:   false,
		},
		{
			name:             "continuous vesting",
			addr:             addr1.String(),
			denom:            "1000atom",
			vestingAmount:    "1000atom",
			vestingStartTime: 1640620000,
			vestingEndTime:   1640630000,
			expectErr:        false,
			want: []authtypes.GenesisAccount{
				&vestingtypes.ContinuousVestingAccount{
					BaseVestingAccount: &vestingtypes.BaseVestingAccount{
						BaseAccount: &authtypes.BaseAccount{
							Address: addr1.String(),
						},
						OriginalVesting:  types.NewCoins(types.NewCoin("atom", types.NewInt(1000))),
						DelegatedFree:    nil,
						DelegatedVesting: nil,
						EndTime:          1640630000,
					},
					StartTime: 1640620000,
				},
			},
		},
		{
			name:                 "periodic vesting with vesting-periods-number",
			addr:                 addr1.String(),
			denom:                "1000atom",
			vestingAmount:        "1000atom",
			vestingStartTime:     1640620000,
			vestingEndTime:       1640629000,
			vestingPeriodsNumber: 10,
			expectErr:            false,
			want: []authtypes.GenesisAccount{
				&vestingtypes.PeriodicVestingAccount{
					BaseVestingAccount: &vestingtypes.BaseVestingAccount{
						BaseAccount: &authtypes.BaseAccount{
							Address: addr1.String(),
						},
						OriginalVesting:  types.NewCoins(types.NewCoin("atom", types.NewInt(1000))),
						DelegatedFree:    nil,
						DelegatedVesting: nil,
						EndTime:          1640629000,
					},
					StartTime: 1640620000,
					VestingPeriods: func() vestingtypes.Periods {
						vestingPeriods := make(vestingtypes.Periods, 0)
						// vest each 1000 millis 10 times
						for i := 0; i < 10; i++ {
							vestingPeriods = append(vestingPeriods, vestingtypes.Period{
								Length: 1000,
								Amount: types.NewCoins(types.NewCoin("atom", types.NewInt(100))),
							})
						}
						vestingPeriods[0].Length = 0
						return vestingPeriods
					}(),
				},
			},
		},
		{
			name:                  "periodic vesting with vesting-periods-amounts",
			addr:                  addr1.String(),
			denom:                 "1000atom",
			vestingAmount:         "1000atom",
			vestingStartTime:      1640620000,
			vestingEndTime:        1640630000,
			vestingPeriodsAmounts: "9000|100atom,900|800atom,100|100atom",
			expectErr:             false,
			want: []authtypes.GenesisAccount{
				&vestingtypes.PeriodicVestingAccount{
					BaseVestingAccount: &vestingtypes.BaseVestingAccount{
						BaseAccount: &authtypes.BaseAccount{
							Address: addr1.String(),
						},
						OriginalVesting:  types.NewCoins(types.NewCoin("atom", types.NewInt(1000))),
						DelegatedFree:    nil,
						DelegatedVesting: nil,
						EndTime:          1640630000,
					},
					StartTime: 1640620000,
					VestingPeriods: func() vestingtypes.Periods {
						vestingPeriods := vestingtypes.Periods{
							{
								Length: 9000,
								Amount: types.NewCoins(types.NewCoin("atom", types.NewInt(100))),
							},
							{
								Length: 900,
								Amount: types.NewCoins(types.NewCoin("atom", types.NewInt(800))),
							},
							{
								Length: 100,
								Amount: types.NewCoins(types.NewCoin("atom", types.NewInt(100))),
							},
						}

						return vestingPeriods
					}(),
				},
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			home := t.TempDir()
			logger := log.NewNopLogger()
			cfg, err := genutiltest.CreateDefaultTendermintConfig(home)
			require.NoError(t, err)

			appCodec := simapp.MakeTestEncodingConfig().Marshaler
			err = genutiltest.ExecInitCmd(testMbm, home, appCodec)
			require.NoError(t, err)

			serverCtx := server.NewContext(viper.New(), cfg, logger)
			clientCtx := client.Context{}.WithCodec(appCodec).WithHomeDir(home)

			if tc.withKeyring {
				path := hd.CreateHDPath(118, 0, 0).String()
				kr, err := keyring.New(sdk.KeyringServiceName(), keyring.BackendMemory, home, nil)
				require.NoError(t, err)
				_, _, err = kr.NewMnemonic(tc.addr, keyring.English, path, keyring.DefaultBIP39Passphrase, hd.Secp256k1)
				require.NoError(t, err)
				clientCtx = clientCtx.WithKeyring(kr)
			}

			ctx := context.Background()
			ctx = context.WithValue(ctx, client.ClientContextKey, &clientCtx)
			ctx = context.WithValue(ctx, server.ServerContextKey, serverCtx)

			cmd := simcmd.AddGenesisAccountCmd(home)
			args := []string{
				tc.addr,
				tc.denom,
				fmt.Sprintf("--%s=home", flags.FlagHome),
			}

			if tc.vestingStartTime > 0 {
				args = append(args, fmt.Sprintf("--%s=%d", simcmd.FlagVestingStart, tc.vestingStartTime))
			}
			if tc.vestingEndTime > 0 {
				args = append(args, fmt.Sprintf("--%s=%d", simcmd.FlagVestingEnd, tc.vestingEndTime))
			}
			if tc.vestingAmount != "" {
				args = append(args, fmt.Sprintf("--%s=%s", simcmd.FlagVestingAmt, tc.vestingAmount))
			}
			if tc.vestingPeriodsNumber > 0 {
				args = append(args, fmt.Sprintf("--%s=%d", simcmd.FlagVestingPeriodsNumber, tc.vestingPeriodsNumber))
			}
			if tc.vestingPeriodsAmounts != "" {
				args = append(args, fmt.Sprintf("--%s=%s", simcmd.FlagVestingPeriodsAmts, tc.vestingPeriodsAmounts))
			}

			args = append(args, fmt.Sprintf("--%s=home", flags.FlagHome))

			cmd.SetArgs(args)
			if tc.expectErr {
				require.Error(t, cmd.ExecuteContext(ctx))
			} else {
				require.NoError(t, cmd.ExecuteContext(ctx))
			}

			genFile := cfg.GenesisFile()
			appState, _, err := genutiltypes.GenesisStateFromGenFile(genFile)
			if err != nil {
				t.Fatal(err)
			}

			authGenState := authtypes.GetGenesisStateFromAppState(appCodec, appState)

			accs, err := authtypes.UnpackAccounts(authGenState.Accounts)
			if err != nil {
				t.Fatal(err)
			}

			if tc.want != nil {
				assert.Equal(t, tc.want, accs)
			}
		})
	}
}
