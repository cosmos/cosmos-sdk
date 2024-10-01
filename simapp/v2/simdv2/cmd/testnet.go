package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"

	cmtconfig "github.com/cometbft/cometbft/config"
	cmttime "github.com/cometbft/cometbft/types/time"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"cosmossdk.io/core/transaction"
	"cosmossdk.io/log"
	"cosmossdk.io/math"
	"cosmossdk.io/math/unsafe"
	runtimev2 "cosmossdk.io/runtime/v2"
	serverv2 "cosmossdk.io/server/v2"
	"cosmossdk.io/server/v2/api/grpc"
	"cosmossdk.io/server/v2/cometbft"
	"cosmossdk.io/server/v2/store"
	banktypes "cosmossdk.io/x/bank/types"
	bankv2types "cosmossdk.io/x/bank/v2/types"
	stakingtypes "cosmossdk.io/x/staking/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
)

var (
	flagNodeDirPrefix     = "node-dir-prefix"
	flagNumValidators     = "validator-count"
	flagOutputDir         = "output-dir"
	flagNodeDaemonHome    = "node-daemon-home"
	flagStartingIPAddress = "starting-ip-address"
	flagListenIPAddress   = "listen-ip-address"
	flagStakingDenom      = "staking-denom"
	flagCommitTimeout     = "commit-timeout"
	flagSingleHost        = "single-host"
)

type initArgs struct {
	algo              string
	chainID           string
	keyringBackend    string
	minGasPrices      string
	nodeDaemonHome    string
	nodeDirPrefix     string
	numValidators     int
	outputDir         string
	startingIPAddress string
	listenIPAddress   string
	singleMachine     bool
	bondTokenDenom    string
}

func addTestnetFlagsToCmd(cmd *cobra.Command) {
	cmd.Flags().IntP(flagNumValidators, "n", 4, "Number of validators to initialize the testnet with")
	cmd.Flags().StringP(flagOutputDir, "o", "./.testnets", "Directory to store initialization data for the testnet")
	cmd.Flags().String(flags.FlagChainID, "", "genesis file chain-id, if left blank will be randomly created")
	cmd.Flags().String(serverv2.FlagMinGasPrices, fmt.Sprintf("0.000006%s", sdk.DefaultBondDenom), "Minimum gas prices to accept for transactions; All fees in a tx must meet this minimum (e.g. 0.01photino,0.001stake)")
	cmd.Flags().String(flags.FlagKeyType, string(hd.Secp256k1Type), "Key signing algorithm to generate keys for")

	// support old flags name for backwards compatibility
	cmd.Flags().SetNormalizeFunc(func(f *pflag.FlagSet, name string) pflag.NormalizedName {
		if name == flags.FlagKeyAlgorithm {
			name = flags.FlagKeyType
		}

		return pflag.NormalizedName(name)
	})
}

// NewTestnetCmd creates a root testnet command with subcommands to run an in-process testnet or initialize
// validator configuration files for running a multi-validator testnet in a separate process
func NewTestnetCmd[T transaction.Tx](mm *runtimev2.MM[T]) *cobra.Command {
	testnetCmd := &cobra.Command{
		Use:                        "testnet",
		Short:                      "subcommands for starting or configuring local testnets",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	testnetCmd.AddCommand(testnetInitFilesCmd(mm))

	return testnetCmd
}

// testnetInitFilesCmd returns a cmd to initialize all files for CometBFT testnet and application
func testnetInitFilesCmd[T transaction.Tx](mm *runtimev2.MM[T]) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init-files",
		Short: "Initialize config directories & files for a multi-validator testnet running locally via separate processes (e.g. Docker Compose or similar)",
		Long: fmt.Sprintf(`init-files will setup one directory per validator and populate each with
necessary files (private validator, genesis, config, etc.) for running validator nodes.

Booting up a network with these validator folders is intended to be used with Docker Compose,
or a similar setup where each node has a manually configurable IP address.

Note, strict routability for addresses is turned off in the config file.

Example:
	%s testnet init-files --validator-count 4 --output-dir ./.testnets --starting-ip-address 192.168.10.2
	`, version.AppName),
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			config := client.GetConfigFromCmd(cmd)

			args := initArgs{}
			args.outputDir, _ = cmd.Flags().GetString(flagOutputDir)
			args.keyringBackend, _ = cmd.Flags().GetString(flags.FlagKeyringBackend)
			args.chainID, _ = cmd.Flags().GetString(flags.FlagChainID)
			args.minGasPrices, _ = cmd.Flags().GetString(serverv2.FlagMinGasPrices)
			args.nodeDirPrefix, _ = cmd.Flags().GetString(flagNodeDirPrefix)
			args.nodeDaemonHome, _ = cmd.Flags().GetString(flagNodeDaemonHome)
			args.startingIPAddress, _ = cmd.Flags().GetString(flagStartingIPAddress)
			args.listenIPAddress, _ = cmd.Flags().GetString(flagListenIPAddress)
			args.numValidators, _ = cmd.Flags().GetInt(flagNumValidators)
			args.algo, _ = cmd.Flags().GetString(flags.FlagKeyType)
			args.bondTokenDenom, _ = cmd.Flags().GetString(flagStakingDenom)
			args.singleMachine, _ = cmd.Flags().GetBool(flagSingleHost)
			config.Consensus.TimeoutCommit, err = cmd.Flags().GetDuration(flagCommitTimeout)
			if err != nil {
				return err
			}

			return initTestnetFiles(clientCtx, cmd, config, mm, args)
		},
	}

	addTestnetFlagsToCmd(cmd)
	cmd.Flags().String(flagNodeDirPrefix, "node", "Prefix for the name of per-validator subdirectories (to be number-suffixed like node0, node1, ...)")
	cmd.Flags().String(flagNodeDaemonHome, "simdv2", "Home directory of the node's daemon configuration")
	cmd.Flags().String(flagStartingIPAddress, "192.168.0.1", "Starting IP address (192.168.0.1 results in persistent peers list ID0@192.168.0.1:46656, ID1@192.168.0.2:46656, ...)")
	cmd.Flags().String(flagListenIPAddress, "127.0.0.1", "TCP or UNIX socket IP address for the RPC server to listen on")
	cmd.Flags().String(flags.FlagKeyringBackend, flags.DefaultKeyringBackend, "Select keyring's backend (os|file|test)")
	cmd.Flags().Duration(flagCommitTimeout, 5*time.Second, "Time to wait after a block commit before starting on the new height")
	cmd.Flags().Bool(flagSingleHost, false, "Cluster runs on a single host machine with different ports")
	cmd.Flags().String(flagStakingDenom, sdk.DefaultBondDenom, "Default staking token denominator")

	return cmd
}

const nodeDirPerm = 0o755

// initTestnetFiles initializes testnet files for a testnet to be run in a separate process
func initTestnetFiles[T transaction.Tx](
	clientCtx client.Context,
	cmd *cobra.Command,
	nodeConfig *cmtconfig.Config,
	mm *runtimev2.MM[T],
	args initArgs,
) error {
	if args.chainID == "" {
		args.chainID = "chain-" + unsafe.Str(6)
	}
	nodeIDs := make([]string, args.numValidators)
	valPubKeys := make([]cryptotypes.PubKey, args.numValidators)

	var (
		genAccounts []authtypes.GenesisAccount
		genBalances []banktypes.Balance
		genFiles    []string
	)
	const (
		rpcPort  = 26657
		apiPort  = 1317
		grpcPort = 9090
	)
	p2pPortStart := 26656

	inBuf := bufio.NewReader(cmd.InOrStdin())
	// generate private keys, node IDs, and initial transactions
	for i := 0; i < args.numValidators; i++ {
		var portOffset int
		grpcConfig := grpc.DefaultConfig()
		if args.singleMachine {
			portOffset = i
			p2pPortStart = 16656 // use different start point to not conflict with rpc port
			nodeConfig.P2P.AddrBookStrict = false
			nodeConfig.P2P.PexReactor = false
			nodeConfig.P2P.AllowDuplicateIP = true

			grpcConfig = &grpc.Config{
				Enable:         true,
				Address:        fmt.Sprintf("127.0.0.1:%d", grpcPort+portOffset),
				MaxRecvMsgSize: grpc.DefaultConfig().MaxRecvMsgSize,
				MaxSendMsgSize: grpc.DefaultConfig().MaxSendMsgSize,
			}
		}

		nodeDirName := fmt.Sprintf("%s%d", args.nodeDirPrefix, i)
		nodeDir := filepath.Join(args.outputDir, nodeDirName, args.nodeDaemonHome)
		cmd.Println("Node dir", nodeDir)
		gentxsDir := filepath.Join(args.outputDir, "gentxs")

		nodeConfig.SetRoot(nodeDir)
		nodeConfig.Moniker = nodeDirName
		nodeConfig.RPC.ListenAddress = fmt.Sprintf("tcp://%s:%d", args.listenIPAddress, rpcPort+portOffset)

		if err := os.MkdirAll(filepath.Join(nodeDir, "config"), nodeDirPerm); err != nil {
			_ = os.RemoveAll(args.outputDir)
			return err
		}
		var (
			err error
			ip  string
		)
		if args.singleMachine {
			ip = "127.0.0.1"
		} else {
			ip, err = getIP(i, args.startingIPAddress)
			if err != nil {
				_ = os.RemoveAll(args.outputDir)
				return err
			}
		}

		nodeIDs[i], valPubKeys[i], err = genutil.InitializeNodeValidatorFiles(nodeConfig, args.algo)
		if err != nil {
			_ = os.RemoveAll(args.outputDir)
			return err
		}

		memo := fmt.Sprintf("%s@%s:%d", nodeIDs[i], ip, p2pPortStart+portOffset)
		genFiles = append(genFiles, nodeConfig.GenesisFile())

		kb, err := keyring.New(sdk.KeyringServiceName(), args.keyringBackend, nodeDir, inBuf, clientCtx.Codec)
		if err != nil {
			return err
		}

		keyringAlgos, _ := kb.SupportedAlgorithms()
		algo, err := keyring.NewSigningAlgoFromString(args.algo, keyringAlgos)
		if err != nil {
			return err
		}

		addr, secret, err := testutil.GenerateSaveCoinKey(kb, nodeDirName, "", true, algo, sdk.GetFullBIP44Path())
		if err != nil {
			_ = os.RemoveAll(args.outputDir)
			return err
		}

		info := map[string]string{"secret": secret}

		cliPrint, err := json.Marshal(info)
		if err != nil {
			return err
		}

		// save private key seed words
		if err := writeFile(fmt.Sprintf("%v.json", "key_seed"), nodeDir, cliPrint); err != nil {
			return err
		}

		accTokens := sdk.TokensFromConsensusPower(1000, sdk.DefaultPowerReduction)
		accStakingTokens := sdk.TokensFromConsensusPower(500, sdk.DefaultPowerReduction)
		coins := sdk.Coins{
			sdk.NewCoin("testtoken", accTokens),
			sdk.NewCoin(args.bondTokenDenom, accStakingTokens),
		}

		addrStr, err := clientCtx.AddressCodec.BytesToString(addr)
		if err != nil {
			return err
		}
		genBalances = append(genBalances, banktypes.Balance{Address: addrStr, Coins: coins.Sort()})
		genAccounts = append(genAccounts, authtypes.NewBaseAccount(addr, nil, 0, 0))

		valStr, err := clientCtx.ValidatorAddressCodec.BytesToString(addr)
		if err != nil {
			return err
		}
		valTokens := sdk.TokensFromConsensusPower(100, sdk.DefaultPowerReduction)
		createValMsg, err := stakingtypes.NewMsgCreateValidator(
			valStr,
			valPubKeys[i],
			sdk.NewCoin(args.bondTokenDenom, valTokens),
			stakingtypes.NewDescription(nodeDirName, "", "", "", ""),
			stakingtypes.NewCommissionRates(math.LegacyOneDec(), math.LegacyOneDec(), math.LegacyOneDec()),
			math.OneInt(),
		)
		if err != nil {
			return err
		}

		txBuilder := clientCtx.TxConfig.NewTxBuilder()
		if err := txBuilder.SetMsgs(createValMsg); err != nil {
			return err
		}

		txBuilder.SetMemo(memo)
		txBuilder.SetGasLimit(flags.DefaultGasLimit)
		txBuilder.SetFeePayer(addr)

		txFactory := tx.Factory{}
		txFactory = txFactory.
			WithChainID(args.chainID).
			WithMemo(memo).
			WithKeybase(kb).
			WithTxConfig(clientCtx.TxConfig)

		if err := tx.Sign(clientCtx, txFactory, nodeDirName, txBuilder, true); err != nil {
			return err
		}

		txBz, err := clientCtx.TxConfig.TxJSONEncoder()(txBuilder.GetTx())
		if err != nil {
			return err
		}

		if err := writeFile(fmt.Sprintf("%v.json", nodeDirName), gentxsDir, txBz); err != nil {
			return err
		}

		serverCfg := serverv2.DefaultServerConfig()
		serverCfg.MinGasPrices = args.minGasPrices

		// Write server config
		cometServer := cometbft.New[T](
			&genericTxDecoder[T]{clientCtx.TxConfig},
			cometbft.ServerOptions[T]{},
			cometbft.OverwriteDefaultConfigTomlConfig(nodeConfig),
		)
		storeServer := store.New[T](newApp)
		grpcServer := grpc.New[T](grpc.OverwriteDefaultConfig(grpcConfig))
		server := serverv2.NewServer(log.NewNopLogger(), serverCfg, cometServer, grpcServer, storeServer)
		err = server.WriteConfig(filepath.Join(nodeDir, "config"))
		if err != nil {
			return err
		}
	}

	if err := initGenFiles(clientCtx, mm, args.chainID, genAccounts, genBalances, genFiles, args.numValidators); err != nil {
		return err
	}

	err := collectGenFiles(
		clientCtx, nodeConfig, args.chainID, nodeIDs, valPubKeys, args.numValidators,
		args.outputDir, args.nodeDirPrefix, args.nodeDaemonHome,
		rpcPort, p2pPortStart, args.singleMachine,
	)
	if err != nil {
		return err
	}

	// Update viper root since root dir become rootdir/node/simd
	serverv2.GetViperFromCmd(cmd).Set(flags.FlagHome, nodeConfig.RootDir)

	cmd.PrintErrf("Successfully initialized %d node directories\n", args.numValidators)
	return nil
}

func initGenFiles[T transaction.Tx](
	clientCtx client.Context, mm *runtimev2.MM[T], chainID string,
	genAccounts []authtypes.GenesisAccount, genBalances []banktypes.Balance,
	genFiles []string, numValidators int,
) error {
	appGenState := mm.DefaultGenesis()

	// set the accounts in the genesis state
	var authGenState authtypes.GenesisState
	clientCtx.Codec.MustUnmarshalJSON(appGenState[authtypes.ModuleName], &authGenState)

	accounts, err := authtypes.PackAccounts(genAccounts)
	if err != nil {
		return err
	}

	authGenState.Accounts = accounts
	appGenState[authtypes.ModuleName] = clientCtx.Codec.MustMarshalJSON(&authGenState)

	// set the balances in the genesis state
	var bankGenState banktypes.GenesisState
	clientCtx.Codec.MustUnmarshalJSON(appGenState[banktypes.ModuleName], &bankGenState)

	bankGenState.Balances, err = banktypes.SanitizeGenesisBalances(genBalances, clientCtx.AddressCodec)
	if err != nil {
		return err
	}
	for _, bal := range bankGenState.Balances {
		bankGenState.Supply = bankGenState.Supply.Add(bal.Coins...)
	}
	appGenState[banktypes.ModuleName] = clientCtx.Codec.MustMarshalJSON(&bankGenState)

	var bankV2GenState bankv2types.GenesisState
	clientCtx.Codec.MustUnmarshalJSON(appGenState[bankv2types.ModuleName], &bankV2GenState)
	if len(bankV2GenState.Balances) == 0 {
		bankV2GenState = getBankV2GenesisFromV1(bankGenState)
	}
	appGenState[bankv2types.ModuleName] = clientCtx.Codec.MustMarshalJSON(&bankV2GenState)

	appGenStateJSON, err := json.MarshalIndent(appGenState, "", "  ")
	if err != nil {
		return err
	}

	appGenesis := genutiltypes.NewAppGenesisWithVersion(chainID, appGenStateJSON)
	// generate empty genesis files for each validator and save
	for i := 0; i < numValidators; i++ {
		if err := appGenesis.SaveAs(genFiles[i]); err != nil {
			return err
		}
	}
	return nil
}

func collectGenFiles(
	clientCtx client.Context, nodeConfig *cmtconfig.Config, chainID string,
	nodeIDs []string, valPubKeys []cryptotypes.PubKey, numValidators int,
	outputDir, nodeDirPrefix, nodeDaemonHome string,
	rpcPortStart, p2pPortStart int,
	singleMachine bool,
) error {
	var appState json.RawMessage
	genTime := cmttime.Now()

	for i := 0; i < numValidators; i++ {
		if singleMachine {
			portOffset := i
			nodeConfig.RPC.ListenAddress = fmt.Sprintf("tcp://127.0.0.1:%d", rpcPortStart+portOffset)
			nodeConfig.P2P.ListenAddress = fmt.Sprintf("tcp://127.0.0.1:%d", p2pPortStart+portOffset)
		}

		nodeDirName := fmt.Sprintf("%s%d", nodeDirPrefix, i)
		nodeDir := filepath.Join(outputDir, nodeDirName, nodeDaemonHome)
		gentxsDir := filepath.Join(outputDir, "gentxs")
		nodeConfig.Moniker = nodeDirName

		nodeConfig.SetRoot(nodeDir)

		nodeID, valPubKey := nodeIDs[i], valPubKeys[i]
		initCfg := genutiltypes.NewInitConfig(chainID, gentxsDir, nodeID, valPubKey)

		appGenesis, err := genutiltypes.AppGenesisFromFile(nodeConfig.GenesisFile())
		if err != nil {
			return err
		}

		nodeAppState, err := genutil.GenAppStateFromConfig(clientCtx.Codec, clientCtx.TxConfig, nodeConfig, initCfg, appGenesis, genutiltypes.DefaultMessageValidator,
			clientCtx.ValidatorAddressCodec, clientCtx.AddressCodec)
		if err != nil {
			return err
		}

		if appState == nil {
			// set the canonical application state (they should not differ)
			appState = nodeAppState
		}

		genFile := nodeConfig.GenesisFile()

		// overwrite each validator's genesis file to have a canonical genesis time
		if err := genutil.ExportGenesisFileWithTime(genFile, chainID, nil, appState, genTime); err != nil {
			return err
		}
	}

	return nil
}

func getIP(i int, startingIPAddr string) (ip string, err error) {
	if len(startingIPAddr) == 0 {
		ip, err = serverv2.ExternalIP()
		if err != nil {
			return "", err
		}
		return ip, nil
	}
	return calculateIP(startingIPAddr, i)
}

func calculateIP(ip string, i int) (string, error) {
	ipv4 := net.ParseIP(ip).To4()
	if ipv4 == nil {
		return "", fmt.Errorf("%v: non ipv4 address", ip)
	}

	for j := 0; j < i; j++ {
		ipv4[3]++
	}

	return ipv4.String(), nil
}

func writeFile(name, dir string, contents []byte) error {
	file := filepath.Join(dir, name)

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("could not create directory %q: %w", dir, err)
	}

	return os.WriteFile(file, contents, 0o600)
}

// getBankV2GenesisFromV1 clones bank/v1 state to bank/v2
// since we not migrate yet
// TODO: Remove
func getBankV2GenesisFromV1(v1GenesisState banktypes.GenesisState) bankv2types.GenesisState {
	var v2GenesisState bankv2types.GenesisState
	for _, balance := range v1GenesisState.Balances {
		v2Balance := bankv2types.Balance(balance)
		v2GenesisState.Balances = append(v2GenesisState.Balances, v2Balance)
		v2GenesisState.Supply = v2GenesisState.Supply.Add(balance.Coins...)
	}
	return v2GenesisState
}
