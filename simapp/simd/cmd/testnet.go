package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	cmtconfig "github.com/cometbft/cometbft/config"
	cmttime "github.com/cometbft/cometbft/types/time"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"cosmossdk.io/math"
	"cosmossdk.io/math/unsafe"
	banktypes "cosmossdk.io/x/bank/types"
	stakingtypes "cosmossdk.io/x/staking/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/server"
	srvconfig "github.com/cosmos/cosmos-sdk/server/config"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
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
	flagEnableLogging     = "enable-logging"
	flagGRPCAddress       = "grpc.address"
	flagRPCAddress        = "rpc.address"
	flagAPIAddress        = "api.address"
	flagPrintMnemonic     = "print-mnemonic"
	flagStakingDenom      = "staking-denom"
	flagCommitTimeout     = "commit-timeout"
	flagSingleHost        = "single-host"

	// default values
	defaultRPCPort           = 26657
	defaultAPIPort           = 1317
	defaultGRPCPort          = 9090
	defaultListenIPAddress   = "127.0.0.1"
	defaultStartingIPAddress = "192.168.0.1"
	defaultNodeDirPrefix     = "node"
	defaultNodeDaemonHome    = "simd"
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

	// start command arguments
	apiListenAddress  string
	grpcListenAddress string
	rpcPort           int
	apiPort           int
	grpcPort          int
	enableLogging     bool
	printMnemonic     bool
}

func addTestnetFlagsToCmd(cmd *cobra.Command) {
	cmd.Flags().IntP(flagNumValidators, "n", 4, "Number of validators to initialize the testnet with")
	cmd.Flags().StringP(flagOutputDir, "o", "./.testnets", "Directory to store initialization data for the testnet")
	cmd.Flags().String(flags.FlagChainID, "", "genesis file chain-id, if left blank will be randomly created")
	cmd.Flags().String(server.FlagMinGasPrices, fmt.Sprintf("0.000006%s", sdk.DefaultBondDenom), "Minimum gas prices to accept for transactions; All fees in a tx must meet this minimum (e.g. 0.01photino,0.001stake)")
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
func NewTestnetCmd(mm *module.Manager) *cobra.Command {
	testnetCmd := &cobra.Command{
		Use:                        "testnet",
		Short:                      "subcommands for starting or configuring local testnets",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	testnetCmd.AddCommand(testnetStartCmd(mm))
	testnetCmd.AddCommand(testnetInitFilesCmd(mm))

	return testnetCmd
}

// testnetInitFilesCmd returns a cmd to initialize all files for CometBFT testnet and application
func testnetInitFilesCmd(mm *module.Manager) *cobra.Command {
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

			args := initArgs{
				rpcPort:           defaultRPCPort,
				apiPort:           defaultAPIPort,
				grpcPort:          defaultGRPCPort,
				apiListenAddress:  defaultListenIPAddress,
				grpcListenAddress: defaultListenIPAddress,
			}
			args.outputDir, _ = cmd.Flags().GetString(flagOutputDir)
			args.keyringBackend, _ = cmd.Flags().GetString(flags.FlagKeyringBackend)
			args.chainID, _ = cmd.Flags().GetString(flags.FlagChainID)
			args.minGasPrices, _ = cmd.Flags().GetString(server.FlagMinGasPrices)
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

			if args.chainID == "" {
				args.chainID = "chain-" + unsafe.Str(6)
			}

			return initTestnetFiles(clientCtx, cmd, config, mm, args)
		},
	}

	addTestnetFlagsToCmd(cmd)
	cmd.Flags().String(flagNodeDirPrefix, defaultNodeDirPrefix, "Prefix for the name of per-validator subdirectories (to be number-suffixed like node0, node1, ...)")
	cmd.Flags().String(flagNodeDaemonHome, defaultNodeDaemonHome, "Home directory of the node's daemon configuration")
	cmd.Flags().String(flagStartingIPAddress, defaultStartingIPAddress, "Starting IP address (192.168.0.1 results in persistent peers list ID0@192.168.0.1:46656, ID1@192.168.0.2:46656, ...)")
	cmd.Flags().String(flagListenIPAddress, defaultListenIPAddress, "TCP or UNIX socket IP address for the RPC server to listen on")
	cmd.Flags().String(flags.FlagKeyringBackend, flags.DefaultKeyringBackend, "Select keyring's backend (os|file|test)")
	cmd.Flags().Duration(flagCommitTimeout, 5*time.Second, "Time to wait after a block commit before starting on the new height")
	cmd.Flags().Bool(flagSingleHost, false, "Cluster runs on a single host machine with different ports")
	cmd.Flags().String(flagStakingDenom, sdk.DefaultBondDenom, "Default staking token denominator")

	return cmd
}

// testnetStartCmd returns a cmd to start multi validator in-process testnet
func testnetStartCmd(mm *module.Manager) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Launch an in-process multi-validator testnet",
		Long: fmt.Sprintf(`testnet will launch an in-process multi-validator testnet,
and generate a directory for each validator populated with necessary
configuration files (private validator, genesis, config, etc.).

Example:
	%s testnet --validator-count4 --output-dir ./.testnets
	`, version.AppName),
		RunE: func(cmd *cobra.Command, _ []string) (err error) {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			config := client.GetConfigFromCmd(cmd)

			args := initArgs{
				singleMachine:  true,
				bondTokenDenom: sdk.DefaultBondDenom,
				nodeDaemonHome: defaultNodeDaemonHome,
				nodeDirPrefix:  defaultNodeDirPrefix,
				keyringBackend: keyring.BackendTest,
			}
			args.outputDir, _ = cmd.Flags().GetString(flagOutputDir)
			args.chainID, _ = cmd.Flags().GetString(flags.FlagChainID)
			args.minGasPrices, _ = cmd.Flags().GetString(server.FlagMinGasPrices)
			args.numValidators, _ = cmd.Flags().GetInt(flagNumValidators)
			args.algo, _ = cmd.Flags().GetString(flags.FlagKeyType)
			args.enableLogging, _ = cmd.Flags().GetBool(flagEnableLogging)

			rpcAddress, _ := cmd.Flags().GetString(flagRPCAddress)
			args.listenIPAddress, args.rpcPort, err = parseURL(rpcAddress)
			if err != nil {
				return fmt.Errorf("invalid rpc address: %w", err)
			}

			apiAddress, _ := cmd.Flags().GetString(flagAPIAddress)
			args.apiListenAddress, args.apiPort, err = parseURL(apiAddress)
			if err != nil {
				return fmt.Errorf("invalid api address: %w", err)
			}

			grpcAddress, _ := cmd.Flags().GetString(flagGRPCAddress)
			// add scheme to avoid issues with parsing
			if !strings.Contains(grpcAddress, "://") {
				grpcAddress = "tcp://" + grpcAddress
			}
			args.grpcListenAddress, args.grpcPort, err = parseURL(grpcAddress)
			if err != nil {
				return fmt.Errorf("invalid grpc address: %w", err)
			}

			args.printMnemonic, _ = cmd.Flags().GetBool(flagPrintMnemonic)

			if args.chainID == "" {
				args.chainID = "chain-" + unsafe.Str(6)
			}

			return startTestnet(clientCtx, cmd, config, mm, args)
		},
	}

	addTestnetFlagsToCmd(cmd)
	cmd.Flags().Bool(flagEnableLogging, false, "Enable INFO logging of CometBFT validator nodes")
	cmd.Flags().String(flagRPCAddress, "tcp://127.0.0.1:26657", "the RPC address to listen on")
	cmd.Flags().String(flagAPIAddress, "tcp://127.0.0.1:1317", "the address to listen on for REST API")
	cmd.Flags().String(flagGRPCAddress, "127.0.0.1:9090", "the gRPC server address to listen on")
	cmd.Flags().Bool(flagPrintMnemonic, true, "print mnemonic of first validator to stdout for manual testing")
	return cmd
}

func parseURL(str string) (host string, port int, err error) {
	u, err := url.Parse(str)
	if err != nil {
		return
	}

	host = u.Hostname()

	port, err = strconv.Atoi(u.Port())
	return
}

const nodeDirPerm = 0o755

// initTestnetFiles initializes testnet files for a testnet to be run in a separate process
func initTestnetFiles(
	clientCtx client.Context,
	cmd *cobra.Command,
	nodeConfig *cmtconfig.Config,
	mm *module.Manager,
	args initArgs,
) error {
	nodeIDs := make([]string, args.numValidators)
	valPubKeys := make([]cryptotypes.PubKey, args.numValidators)

	appConfig := srvconfig.DefaultConfig()
	appConfig.MinGasPrices = args.minGasPrices
	appConfig.API.Enable = true
	appConfig.GRPC.Enable = true
	appConfig.Telemetry.Enabled = true
	appConfig.Telemetry.PrometheusRetentionTime = 60
	appConfig.Telemetry.EnableHostnameLabel = false
	appConfig.Telemetry.GlobalLabels = [][]string{{"chain_id", args.chainID}}

	var (
		genAccounts []authtypes.GenesisAccount
		genBalances []banktypes.Balance
		genFiles    []string
	)
	var (
		rpcPort  = args.rpcPort
		apiPort  = args.apiPort
		grpcPort = args.grpcPort
	)
	p2pPortStart := 26656

	inBuf := bufio.NewReader(cmd.InOrStdin())
	// generate private keys, node IDs, and initial transactions
	for i := 0; i < args.numValidators; i++ {
		var portOffset int
		if args.singleMachine {
			portOffset = i
			p2pPortStart = 16656 // use different start point to not conflict with rpc port
			nodeConfig.P2P.AddrBookStrict = false
			nodeConfig.P2P.PexReactor = false
			nodeConfig.P2P.AllowDuplicateIP = true
			appConfig.API.Address = fmt.Sprintf("tcp://%s:%d", args.apiListenAddress, apiPort+portOffset)
			appConfig.GRPC.Address = fmt.Sprintf("%s:%d", args.grpcListenAddress, grpcPort+portOffset)
		}

		nodeDirName, nodeDir := getNodeDir(args, i)
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

		// if PrintMnemonic is set to true, we print the first validator node's secret to the network's logger
		// for debugging and manual testing
		if args.printMnemonic && i == 0 {
			printMnemonic(secret)
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
			stakingtypes.NewDescription(nodeDirName, "", "", "", "", stakingtypes.Metadata{}),
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

		if err := srvconfig.SetConfigTemplate(srvconfig.DefaultConfigTemplate); err != nil {
			return err
		}

		if err := srvconfig.WriteConfigFile(filepath.Join(nodeDir, "config", "app.toml"), appConfig); err != nil {
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
	client.GetViperFromCmd(cmd).Set(flags.FlagHome, nodeConfig.RootDir)

	cmd.PrintErrf("Successfully initialized %d node directories\n", args.numValidators)
	return nil
}

func initGenFiles(
	clientCtx client.Context, mm *module.Manager, chainID string,
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

func writeFile(name, dir string, contents []byte) error {
	file := filepath.Join(dir, name)

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("could not create directory %q: %w", dir, err)
	}

	return os.WriteFile(file, contents, 0o600)
}

// printMnemonic prints a provided mnemonic seed phrase on a network logger
// for debugging and manual testing
func printMnemonic(secret string) {
	lines := []string{
		"THIS MNEMONIC IS FOR TESTING PURPOSES ONLY",
		"DO NOT USE IN PRODUCTION",
		"",
		strings.Join(strings.Fields(secret)[0:8], " "),
		strings.Join(strings.Fields(secret)[8:16], " "),
		strings.Join(strings.Fields(secret)[16:24], " "),
	}

	lineLengths := make([]int, len(lines))
	for i, line := range lines {
		lineLengths[i] = len(line)
	}

	maxLineLength := 0
	for _, lineLen := range lineLengths {
		if lineLen > maxLineLength {
			maxLineLength = lineLen
		}
	}

	fmt.Printf("\n\n")
	fmt.Println(strings.Repeat("+", maxLineLength+8))
	for _, line := range lines {
		fmt.Printf("++  %s  ++\n", centerText(line, maxLineLength))
	}
	fmt.Println(strings.Repeat("+", maxLineLength+8))
	fmt.Printf("\n\n")
}

// centerText centers text across a fixed width, filling either side with whitespace buffers
func centerText(text string, width int) string {
	textLen := len(text)
	leftBuffer := strings.Repeat(" ", (width-textLen)/2)
	rightBuffer := strings.Repeat(" ", (width-textLen)/2+(width-textLen)%2)

	return fmt.Sprintf("%s%s%s", leftBuffer, text, rightBuffer)
}

func getNodeDir(args initArgs, nodeID int) (nodeDirName, nodeDir string) {
	nodeDirName = fmt.Sprintf("%s%d", args.nodeDirPrefix, nodeID)
	nodeDir = filepath.Join(args.outputDir, nodeDirName, args.nodeDaemonHome)
	return
}

// startTestnet starts an in-process testnet
func startTestnet(
	clientCtx client.Context,
	cmd *cobra.Command,
	nodeConfig *cmtconfig.Config,
	mm *module.Manager,
	args initArgs,
) error {
	fmt.Printf(`Preparing test network with chain-id "%s"`, args.chainID)

	args.outputDir = fmt.Sprintf("%s/%s", args.outputDir, args.chainID)
	err := initTestnetFiles(clientCtx, cmd, nodeConfig, mm, args)
	if err != nil {
		return err
	}

	// slice to keep track of validator processes
	var processes []*exec.Cmd

	// channel to signal shutdown
	shutdownCh := make(chan struct{})

	fmt.Println("Starting test network...")
	// Start each validator in a separate process
	for i := 0; i < args.numValidators; i++ {
		_, nodeDir := getNodeDir(args, i)

		// run start command
		cmdArgs := []string{"start", fmt.Sprintf("--%s=%s", flags.FlagHome, nodeDir)}
		runCmd := exec.Command(os.Args[0], cmdArgs...) // spawn new process

		// Set stdout and stderr based on enableLogging flag
		if args.enableLogging {
			runCmd.Stdout = os.Stdout
			runCmd.Stderr = os.Stderr
		} else {
			runCmd.Stdout = io.Discard // discard output when logging is disabled
			runCmd.Stderr = io.Discard
		}

		if err := runCmd.Start(); err != nil {
			return fmt.Errorf("failed to start validator %d: %w", i, err)
		}
		fmt.Printf("Started Validator %d\n", i+1)
		processes = append(processes, runCmd) // add to processes slice
	}

	// goroutine to listen for Enter key press
	go func() {
		fmt.Println("Press the Enter Key to terminate all validator processes")
		if _, err := fmt.Scanln(); err == nil {
			close(shutdownCh) // Signal shutdown
		}
	}()

	// goroutine to listen for Ctrl+C (SIGINT)
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	go func() {
		<-sigCh // Wait for Ctrl+C
		fmt.Println("\nCtrl+C detected, terminating validator processes...")
		close(shutdownCh) // Signal shutdown
	}()

	// block until shutdown signal is received
	<-shutdownCh

	// terminate all validator processes
	fmt.Println("Shutting down validator processes...")
	for i, p := range processes {
		if err := p.Process.Kill(); err != nil {
			fmt.Printf("Failed to terminate validator %d process: %v\n", i+1, err)
		} else {
			fmt.Printf("Validator %d terminated\n", i+1)
		}
	}
	_ = os.RemoveAll(args.outputDir) // Clean up the output directory
	fmt.Println("Finished cleaning up test network")

	return nil
}
