package cli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"golang.org/x/exp/rand"

	bank "cosmossdk.io/api/cosmos/bank/v1beta1"
	base "cosmossdk.io/api/cosmos/base/v1beta1"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	clienttx "github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
)

const (
	flagVestingStart = "vesting-start-time"
	flagVestingEnd   = "vesting-end-time"
	flagVestingAmt   = "vesting-amount"
	flagAppendMode   = "append"
	flagModuleName   = "module-name"
)

// AddGenesisAccountCmd returns add-genesis-account cobra Command.
// This command is provided as a default, applications are expected to provide their own command if custom genesis accounts are needed.
func AddGenesisAccountCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-genesis-account <address_or_key_name> <coin>[,<coin>...]",
		Short: "Add a genesis account to genesis.json",
		Long: `Add a genesis account to genesis.json. The provided account must specify
the account address or key name and a list of initial coins. If a key name is given,
the address will be looked up in the local Keybase. The list of initial tokens must
contain valid denominations. Accounts may optionally be supplied with vesting parameters.
`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			config := client.GetConfigFromCmd(cmd)

			addressCodec := clientCtx.TxConfig.SigningContext().AddressCodec()
			var kr keyring.Keyring
			addr, err := addressCodec.StringToBytes(args[0])
			if err != nil {
				inBuf := bufio.NewReader(cmd.InOrStdin())
				keyringBackend, _ := cmd.Flags().GetString(flags.FlagKeyringBackend)

				if keyringBackend != "" && clientCtx.Keyring == nil {
					var err error
					kr, err = keyring.New(sdk.KeyringServiceName(), keyringBackend, clientCtx.HomeDir, inBuf, clientCtx.Codec)
					if err != nil {
						return err
					}
				} else {
					kr = clientCtx.Keyring
				}

				k, err := kr.Key(args[0])
				if err != nil {
					return fmt.Errorf("failed to get address from Keyring: %w", err)
				}

				addr, err = k.GetAddress()
				if err != nil {
					return err
				}
			}

			appendflag, _ := cmd.Flags().GetBool(flagAppendMode)
			vestingStart, _ := cmd.Flags().GetInt64(flagVestingStart)
			vestingEnd, _ := cmd.Flags().GetInt64(flagVestingEnd)
			vestingAmtStr, _ := cmd.Flags().GetString(flagVestingAmt)
			moduleNameStr, _ := cmd.Flags().GetString(flagModuleName)

			addrStr, err := addressCodec.BytesToString(addr)
			if err != nil {
				return err
			}

			coins, err := sdk.ParseCoinsNormalized(args[1])
			if err != nil {
				return err
			}

			vestingAmt, err := sdk.ParseCoinsNormalized(vestingAmtStr)
			if err != nil {
				return err
			}

			accounts := []genutil.GenesisAccount{
				{
					Address:      addrStr,
					Coins:        coins,
					VestingAmt:   vestingAmt,
					VestingStart: vestingStart,
					VestingEnd:   vestingEnd,
					ModuleName:   moduleNameStr,
				},
			}

			return genutil.AddGenesisAccounts(clientCtx.Codec, clientCtx.AddressCodec, accounts, appendflag, config.GenesisFile())
		},
	}

	cmd.Flags().String(flags.FlagKeyringBackend, flags.DefaultKeyringBackend, "Select keyring's backend (os|file|kwallet|pass|test)")
	cmd.Flags().String(flagVestingAmt, "", "amount of coins for vesting accounts")
	cmd.Flags().Int64(flagVestingStart, 0, "schedule start time (unix epoch) for vesting accounts")
	cmd.Flags().Int64(flagVestingEnd, 0, "schedule end time (unix epoch) for vesting accounts")
	cmd.Flags().Bool(flagAppendMode, false, "append the coins to an account already in the genesis.json file")
	cmd.Flags().String(flagModuleName, "", "module account name")
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// AddBulkGenesisAccountCmd returns bulk-add-genesis-account cobra Command.
// This command is provided as a default, applications are expected to provide their own command if custom genesis accounts are needed.
func AddBulkGenesisAccountCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bulk-add-genesis-account [/file/path.json]",
		Short: "Bulk add genesis accounts to genesis.json",
		Example: `bulk-add-genesis-account accounts.json

where accounts.json is:

[
    {
        "address": "cosmos139f7kncmglres2nf3h4hc4tade85ekfr8sulz5",
        "coins": [
            { "denom": "umuon", "amount": "100000000" },
            { "denom": "stake", "amount": "200000000" }
        ]
    },
    {
        "address": "cosmos1e0jnq2sun3dzjh8p2xq95kk0expwmd7shwjpfg",
        "coins": [
            { "denom": "umuon", "amount": "500000000" }
        ],
        "vesting_amt": [
            { "denom": "umuon", "amount": "400000000" }
        ],
        "vesting_start": 1724711478,
        "vesting_end": 1914013878
    }
]
`,
		Long: `Add genesis accounts in bulk to genesis.json. The provided account must specify
the account address and a list of initial coins. The list of initial tokens must
contain valid denominations. Accounts may optionally be supplied with vesting parameters.
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			config := client.GetConfigFromCmd(cmd)

			f, err := os.Open(args[0])
			if err != nil {
				return fmt.Errorf("failed to open file: %w", err)
			}
			defer f.Close()

			var accounts []genutil.GenesisAccount
			if err := json.NewDecoder(f).Decode(&accounts); err != nil {
				return fmt.Errorf("failed to decode JSON: %w", err)
			}

			appendflag, _ := cmd.Flags().GetBool(flagAppendMode)

			return genutil.AddGenesisAccounts(clientCtx.Codec, clientCtx.AddressCodec, accounts, appendflag, config.GenesisFile())
		},
	}

	cmd.Flags().Bool(flagAppendMode, false, "append the coins to an account already in the genesis.json file")
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func GenerateSendTransactions() *cobra.Command {
	var (
		seed  uint64
		numTx uint64
	)
	cmd := &cobra.Command{
		Use:   "generate-send-txs",
		Short: "Generate genesis transactions for sending coins to genesis accounts",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			addressCodec := clientCtx.TxConfig.SigningContext().AddressCodec()
			kr, err := keyring.New(
				sdk.KeyringServiceName(),
				keyring.BackendTest,
				clientCtx.HomeDir,
				os.Stdin,
				clientCtx.Codec,
			)
			if err != nil {
				return err
			}

			records, err := kr.List()
			if err != nil {
				return err
			}

			sequences := make(map[string]uint64)

			rnd := rand.New(rand.NewSource(seed))
			for i := 0; i < int(numTx); i++ {
				n := rnd.Intn(len(records))
				if records[n].Name == "alice" {
					i--
					continue
				}
				addr, err := records[n].GetAddress()
				if err != nil {
					return err
				}
				from, err := addressCodec.BytesToString(addr)
				if err != nil {
					return err
				}
				toBytes := make([]byte, 32)
				_, err = rnd.Read(toBytes)
				if err != nil {
					return err
				}
				to, err := addressCodec.BytesToString(toBytes)
				if err != nil {
					return err
				}
				sendMsg := &bank.MsgSend{
					FromAddress: from,
					ToAddress:   to,
					Amount: []*base.Coin{
						{Denom: "stake", Amount: "10"},
					},
				}
				txf, err := clienttx.NewFactoryCLI(clientCtx, cmd.Flags())
				if err != nil {
					return err
				}
				txf = txf.WithSequence(sequences[from])
				sequences[from]++
				tx, err := txf.BuildUnsignedTx(sendMsg)
				if err != nil {
					return err
				}

				err = clienttx.Sign(clientCtx, txf, records[n].Name, tx, false)
				if err != nil {
					return err
				}
				enc := clientCtx.TxConfig.TxJSONEncoder()

				jsonBz, err := enc(tx.GetTx())
				if err != nil {
					return err
				}
				fmt.Println(string(jsonBz))
			}

			return nil
		},
	}
	cmd.Flags().Uint64Var(&seed, "seed", 0, "seed for generation")
	cmd.Flags().Uint64Var(&numTx, "num-txs", 10_000, "the number of txs to generate")
	return cmd
}
