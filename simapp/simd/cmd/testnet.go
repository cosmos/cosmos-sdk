package cmd

// DONTCOVER

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	tmconfig "github.com/tendermint/tendermint/config"
	tmos "github.com/tendermint/tendermint/libs/os"
	tmrand "github.com/tendermint/tendermint/libs/rand"
	"github.com/tendermint/tendermint/types"
	tmtime "github.com/tendermint/tendermint/types/time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/server"
	srvconfig "github.com/cosmos/cosmos-sdk/server/config"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

var (
	flagNodeDirPrefix     = "node-dir-prefix"
	flagNumValidators     = "v"
	flagOutputDir         = "output-dir"
	flagNodeDaemonHome    = "node-daemon-home"
	flagStartingIPAddress = "starting-ip-address"
	flagEnableLogging     = "enable-logging"
	flagGRPCAddress       = "grpc.address"
	flagRPCAddress        = "rpc.address"
	flagAPIAddress        = "api.address"
	flagPrintMnemonic     = "print-mnemonic"
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
}

type startArgs struct {
	algo          string
	apiAddress    string
	chainID       string
	enableLogging bool
	grpcAddress   string
	minGasPrices  string
	numValidators int
	outputDir     string
	printMnemonic bool
	rpcAddress    string
}

func addTestnetFlagsToCmd(cmd *cobra.Command) {
	cmd.Flags().Int(flagNumValidators, 4, "Number of validators to initialize the testnet with")
	cmd.Flags().StringP(flagOutputDir, "o", "./.testnets", "Directory to store initialization data for the testnet")
	cmd.Flags().String(flags.FlagChainID, "", "genesis file chain-id, if left blank will be randomly created")
	cmd.Flags().String(server.FlagMinGasPrices, fmt.Sprintf("0.000006%s", sdk.DefaultBondDenom), "Minimum gas prices to accept for transactions; All fees in a tx must meet this minimum (e.g. 0.01photino,0.001stake)")
	cmd.Flags().String(flags.FlagKeyAlgorithm, string(hd.Secp256k1Type), "Key signing algorithm to generate keys for")
}

// NewTestnetCmd creates a root testnet command with subcommands to run an in-process testnet or initialize
// validator configuration files for running a multi-validator testnet in a separate process
func NewTestnetCmd(mbm module.BasicManager, genBalIterator banktypes.GenesisBalancesIterator) *cobra.Command {
	testnetCmd := &cobra.Command{
		Use:                        "testnet",
		Short:                      "subcommands for starting or configuring local testnets",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	testnetCmd.AddCommand(testnetStartCmd())
	testnetCmd.AddCommand(testnetInitFilesCmd(mbm, genBalIterator))

	return testnetCmd
}

// get cmd to initialize all files for tendermint testnet and application
func testnetInitFilesCmd(mbm module.BasicManager, genBalIterator banktypes.GenesisBalancesIterator) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init-files",
		Short: "Initialize config directories & files for a multi-validator testnet running locally via separate processes (e.g. Docker Compose or similar)",
		Long: `init-files will setup "v" number of directories and populate each with
necessary files (private validator, genesis, config, etc.) for running "v" validator nodes.

Booting up a network with these validator folders is intended to be used with Docker Compose,
or a similar setup where each node has a manually configurable IP address.

Note, strict routability for addresses is turned off in the config file.

Example:
	simd testnet init-files --v 4 --output-dir ./.testnets --starting-ip-address 192.168.10.2
	`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			serverCtx := server.GetServerContextFromCmd(cmd)
			config := serverCtx.Config

			args := initArgs{}
			args.outputDir, _ = cmd.Flags().GetString(flagOutputDir)
			args.keyringBackend, _ = cmd.Flags().GetString(flags.FlagKeyringBackend)
			args.chainID, _ = cmd.Flags().GetString(flags.FlagChainID)
			args.minGasPrices, _ = cmd.Flags().GetString(server.FlagMinGasPrices)
			args.nodeDirPrefix, _ = cmd.Flags().GetString(flagNodeDirPrefix)
			args.nodeDaemonHome, _ = cmd.Flags().GetString(flagNodeDaemonHome)
			args.startingIPAddress, _ = cmd.Flags().GetString(flagStartingIPAddress)
			args.numValidators, _ = cmd.Flags().GetInt(flagNumValidators)
			args.algo, _ = cmd.Flags().GetString(flags.FlagKeyAlgorithm)

			return initTestnetFiles(clientCtx, cmd, config, mbm, genBalIterator, args)

		},
	}

	addTestnetFlagsToCmd(cmd)
	cmd.Flags().String(flagNodeDirPrefix, "node", "Prefix the directory name for each node with (node results in node0, node1, ...)")
	cmd.Flags().String(flagNodeDaemonHome, "simd", "Home directory of the node's daemon configuration")
	cmd.Flags().String(flagStartingIPAddress, "192.168.0.1", "Starting IP address (192.168.0.1 results in persistent peers list ID0@192.168.0.1:46656, ID1@192.168.0.2:46656, ...)")
	cmd.Flags().String(flags.FlagKeyringBackend, flags.DefaultKeyringBackend, "Select keyring's backend (os|file|test)")

	return cmd
}

// get cmd to start multi validator in-process testnet
func testnetStartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Launch an in-process multi-validator testnet",
		Long: `testnet will launch an in-process multi-validator testnet,
and generate "v" directories, populated with necessary validator configuration files
(private validator, genesis, config, etc.).

Example:
	simd testnet --v 4 --output-dir ./.testnets
	`,
		RunE: func(cmd *cobra.Command, _ []string) error {

			args := startArgs{}
			args.outputDir, _ = cmd.Flags().GetString(flagOutputDir)
			args.chainID, _ = cmd.Flags().GetString(flags.FlagChainID)
			args.minGasPrices, _ = cmd.Flags().GetString(server.FlagMinGasPrices)
			args.numValidators, _ = cmd.Flags().GetInt(flagNumValidators)
			args.algo, _ = cmd.Flags().GetString(flags.FlagKeyAlgorithm)
			args.enableLogging, _ = cmd.Flags().GetBool(flagEnableLogging)
			args.rpcAddress, _ = cmd.Flags().GetString(flagRPCAddress)
			args.apiAddress, _ = cmd.Flags().GetString(flagAPIAddress)
			args.grpcAddress, _ = cmd.Flags().GetString(flagGRPCAddress)
			args.printMnemonic, _ = cmd.Flags().GetBool(flagPrintMnemonic)

			return startTestnet(cmd, args)

		},
	}

	addTestnetFlagsToCmd(cmd)
	cmd.Flags().Bool(flagEnableLogging, false, "Enable INFO logging of tendermint validator nodes")
	cmd.Flags().String(flagRPCAddress, "tcp://0.0.0.0:26657", "the RPC address to listen on")
	cmd.Flags().String(flagAPIAddress, "tcp://0.0.0.0:1317", "the address to listen on for REST API")
	cmd.Flags().String(flagGRPCAddress, "0.0.0.0:9090", "the gRPC server address to listen on")
	cmd.Flags().Bool(flagPrintMnemonic, true, "print mnemonic of first validator to stdout for manual testing")
	return cmd
}

const nodeDirPerm = 0755

// initTestnetFiles initializes testnet files for a testnet to be run in a separate process
func initTestnetFiles(
	clientCtx client.Context,
	cmd *cobra.Command,
	nodeConfig *tmconfig.Config,
	mbm module.BasicManager,
	genBalIterator banktypes.GenesisBalancesIterator,
	args initArgs,
) error {

	if args.chainID == "" {
		args.chainID = "chain-" + tmrand.NewRand().Str(6)
	}

	nodeIDs := make([]string, args.numValidators)
	valPubKeys := make([]cryptotypes.PubKey, args.numValidators)

	simappConfig := srvconfig.DefaultConfig()
	simappConfig.MinGasPrices = args.minGasPrices
	simappConfig.API.Enable = true
	simappConfig.Telemetry.Enabled = true
	simappConfig.Telemetry.PrometheusRetentionTime = 60
	simappConfig.Telemetry.EnableHostnameLabel = false
	simappConfig.Telemetry.GlobalLabels = [][]string{{"chain_id", args.chainID}}

	var (
		genAccounts []authtypes.GenesisAccount
		genBalances []banktypes.Balance
		genFiles    []string
	)

	inBuf := bufio.NewReader(cmd.InOrStdin())
	// generate private keys, node IDs, and initial transactions
	for i := 0; i < args.numValidators; i++ {
		nodeDirName := fmt.Sprintf("%s%d", args.nodeDirPrefix, i)
		nodeDir := filepath.Join(args.outputDir, nodeDirName, args.nodeDaemonHome)
		gentxsDir := filepath.Join(args.outputDir, "gentxs")

		nodeConfig.SetRoot(nodeDir)
		nodeConfig.RPC.ListenAddress = "tcp://0.0.0.0:26657"

		if err := os.MkdirAll(filepath.Join(nodeDir, "config"), nodeDirPerm); err != nil {
			_ = os.RemoveAll(args.outputDir)
			return err
		}

		nodeConfig.Moniker = nodeDirName

		ip, err := getIP(i, args.startingIPAddress)
		if err != nil {
			_ = os.RemoveAll(args.outputDir)
			return err
		}

		nodeIDs[i], valPubKeys[i], err = genutil.InitializeNodeValidatorFiles(nodeConfig)
		if err != nil {
			_ = os.RemoveAll(args.outputDir)
			return err
		}

		memo := fmt.Sprintf("%s@%s:26656", nodeIDs[i], ip)
		genFiles = append(genFiles, nodeConfig.GenesisFile())

		kb, err := keyring.New(sdk.KeyringServiceName(), args.keyringBackend, nodeDir, inBuf)
		if err != nil {
			return err
		}

		keyringAlgos, _ := kb.SupportedAlgorithms()
		algo, err := keyring.NewSigningAlgoFromString(args.algo, keyringAlgos)
		if err != nil {
			return err
		}

		addr, secret, err := server.GenerateSaveCoinKey(kb, nodeDirName, true, algo)
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
			sdk.NewCoin(sdk.DefaultBondDenom, accStakingTokens),
		}

		genBalances = append(genBalances, banktypes.Balance{Address: addr.String(), Coins: coins.Sort()})
		genAccounts = append(genAccounts, authtypes.NewBaseAccount(addr, nil, 0, 0))

		valTokens := sdk.TokensFromConsensusPower(100, sdk.DefaultPowerReduction)
		createValMsg, err := stakingtypes.NewMsgCreateValidator(
			sdk.ValAddress(addr),
			valPubKeys[i],
			sdk.NewCoin(sdk.DefaultBondDenom, valTokens),
			stakingtypes.NewDescription(nodeDirName, "", "", "", ""),
			stakingtypes.NewCommissionRates(sdk.OneDec(), sdk.OneDec(), sdk.OneDec()),
			sdk.OneInt(),
		)
		if err != nil {
			return err
		}

		txBuilder := clientCtx.TxConfig.NewTxBuilder()
		if err := txBuilder.SetMsgs(createValMsg); err != nil {
			return err
		}

		txBuilder.SetMemo(memo)

		txFactory := tx.Factory{}
		txFactory = txFactory.
			WithChainID(args.chainID).
			WithMemo(memo).
			WithKeybase(kb).
			WithTxConfig(clientCtx.TxConfig)

		if err := tx.Sign(txFactory, nodeDirName, txBuilder, true); err != nil {
			return err
		}

		txBz, err := clientCtx.TxConfig.TxJSONEncoder()(txBuilder.GetTx())
		if err != nil {
			return err
		}

		if err := writeFile(fmt.Sprintf("%v.json", nodeDirName), gentxsDir, txBz); err != nil {
			return err
		}

		srvconfig.WriteConfigFile(filepath.Join(nodeDir, "config/app.toml"), simappConfig)
	}

	if err := initGenFiles(clientCtx, mbm, args.chainID, genAccounts, genBalances, genFiles, args.numValidators); err != nil {
		return err
	}

	err := collectGenFiles(
		clientCtx, nodeConfig, args.chainID, nodeIDs, valPubKeys, args.numValidators,
		args.outputDir, args.nodeDirPrefix, args.nodeDaemonHome, genBalIterator,
	)
	if err != nil {
		return err
	}

	cmd.PrintErrf("Successfully initialized %d node directories\n", args.numValidators)
	return nil
}

func initGenFiles(
	clientCtx client.Context, mbm module.BasicManager, chainID string,
	genAccounts []authtypes.GenesisAccount, genBalances []banktypes.Balance,
	genFiles []string, numValidators int,
) error {

	appGenState := mbm.DefaultGenesis(clientCtx.Codec)

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

	bankGenState.Balances = banktypes.SanitizeGenesisBalances(genBalances)
	for _, bal := range bankGenState.Balances {
		bankGenState.Supply = bankGenState.Supply.Add(bal.Coins...)
	}
	appGenState[banktypes.ModuleName] = clientCtx.Codec.MustMarshalJSON(&bankGenState)

	appGenStateJSON, err := json.MarshalIndent(appGenState, "", "  ")
	if err != nil {
		return err
	}

	genDoc := types.GenesisDoc{
		ChainID:    chainID,
		AppState:   appGenStateJSON,
		Validators: nil,
	}

	// generate empty genesis files for each validator and save
	for i := 0; i < numValidators; i++ {
		if err := genDoc.SaveAs(genFiles[i]); err != nil {
			return err
		}
	}
	return nil
}

func collectGenFiles(
	clientCtx client.Context, nodeConfig *tmconfig.Config, chainID string,
	nodeIDs []string, valPubKeys []cryptotypes.PubKey, numValidators int,
	outputDir, nodeDirPrefix, nodeDaemonHome string, genBalIterator banktypes.GenesisBalancesIterator,
) error {

	var appState json.RawMessage
	genTime := tmtime.Now()

	for i := 0; i < numValidators; i++ {
		nodeDirName := fmt.Sprintf("%s%d", nodeDirPrefix, i)
		nodeDir := filepath.Join(outputDir, nodeDirName, nodeDaemonHome)
		gentxsDir := filepath.Join(outputDir, "gentxs")
		nodeConfig.Moniker = nodeDirName

		nodeConfig.SetRoot(nodeDir)

		nodeID, valPubKey := nodeIDs[i], valPubKeys[i]
		initCfg := genutiltypes.NewInitConfig(chainID, gentxsDir, nodeID, valPubKey)

		genDoc, err := types.GenesisDocFromFile(nodeConfig.GenesisFile())
		if err != nil {
			return err
		}

		nodeAppState, err := genutil.GenAppStateFromConfig(clientCtx.Codec, clientCtx.TxConfig, nodeConfig, initCfg, *genDoc, genBalIterator)
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
		ip, err = server.ExternalIP()
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

func writeFile(name string, dir string, contents []byte) error {
	writePath := filepath.Join(dir)
	file := filepath.Join(writePath, name)

	err := tmos.EnsureDir(writePath, 0755)
	if err != nil {
		return err
	}

	err = tmos.WriteFile(file, contents, 0644)
	if err != nil {
		return err
	}

	return nil
}

// startTestnet starts an in-process testnet
func startTestnet(cmd *cobra.Command, args startArgs) error {
	networkConfig := network.DefaultConfig()

	// Default networkConfig.ChainID is random, and we should only override it if chainID provided
	// is non-empty
	if args.chainID != "" {
		networkConfig.ChainID = args.chainID
	}
	networkConfig.SigningAlgo = args.algo
	networkConfig.MinGasPrices = args.minGasPrices
	networkConfig.NumValidators = args.numValidators
	networkConfig.EnableTMLogging = args.enableLogging
	networkConfig.RPCAddress = args.rpcAddress
	networkConfig.APIAddress = args.apiAddress
	networkConfig.GRPCAddress = args.grpcAddress
	networkConfig.PrintMnemonic = args.printMnemonic
	networkLogger := network.NewCLILogger(cmd)

	baseDir := fmt.Sprintf("%s/%s", args.outputDir, networkConfig.ChainID)
	if _, err := os.Stat(baseDir); !os.IsNotExist(err) {
		return fmt.Errorf(
			"testnests directory already exists for chain-id '%s': %s, please remove or select a new --chain-id",
			networkConfig.ChainID, baseDir)
	}

	testnet, err := network.New(networkLogger, baseDir, networkConfig)
	if err != nil {
		return err
	}

	testnet.WaitForHeight(1)
	cmd.Println("press the Enter Key to terminate")
	fmt.Scanln() // wait for Enter Key
	testnet.Cleanup()

	return nil
}
