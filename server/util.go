package server

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	tmcfg "github.com/tendermint/tendermint/config"
	tmcli "github.com/tendermint/tendermint/libs/cli"
	tmflags "github.com/tendermint/tendermint/libs/cli/flags"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server/config"
	"github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
)

// DONTCOVER

// ServerContextKey defines the context key used to retrieve a server.Context from
// a command's Context.
const ServerContextKey = sdk.ContextKey("server.context")

// server context
type Context struct {
	Viper  *viper.Viper
	Config *tmcfg.Config
	Logger log.Logger
}

// ErrorCode contains the exit code for server exit.
type ErrorCode struct {
	Code int
}

func (e ErrorCode) Error() string {
	return strconv.Itoa(e.Code)
}

func NewDefaultContext() *Context {
	return NewContext(viper.New(), tmcfg.DefaultConfig(), log.NewTMLogger(log.NewSyncWriter(os.Stdout)))
}

func NewContext(v *viper.Viper, config *tmcfg.Config, logger log.Logger) *Context {
	return &Context{v, config, logger}
}

// InterceptConfigsPreRunHandler performs a pre-run function for the root daemon
// application command. It will create a Viper literal and a default server
// Context. The server Tendermint configuration will either be read and parsed
// or created and saved to disk, where the server Context is updated to reflect
// the Tendermint configuration. The Viper literal is used to read and parse
// the application configuration. Command handlers can fetch the server Context
// to get the Tendermint configuration or to get access to Viper.
func InterceptConfigsPreRunHandler(cmd *cobra.Command) error {
	rootViper := viper.New()
	rootViper.BindPFlags(cmd.Flags())
	rootViper.BindPFlags(cmd.PersistentFlags())

	serverCtx := NewDefaultContext()
	config, err := interceptConfigs(serverCtx, rootViper)
	if err != nil {
		return err
	}

	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout))
	logger, err = tmflags.ParseLogLevel(config.LogLevel, logger, tmcfg.DefaultLogLevel())
	if err != nil {
		return err
	}

	if rootViper.GetBool(tmcli.TraceFlag) {
		logger = log.NewTracingLogger(logger)
	}

	serverCtx.Config = config
	serverCtx.Logger = logger.With("module", "main")

	return SetCmdServerContext(cmd, serverCtx)
}

// GetServerContextFromCmd returns a Context from a command or an empty Context
// if it has not been set.
func GetServerContextFromCmd(cmd *cobra.Command) *Context {
	if v := cmd.Context().Value(ServerContextKey); v != nil {
		serverCtxPtr := v.(*Context)
		return serverCtxPtr
	}

	return NewDefaultContext()
}

// SetCmdServerContext sets a command's Context value to the provided argument.
func SetCmdServerContext(cmd *cobra.Command, serverCtx *Context) error {
	v := cmd.Context().Value(ServerContextKey)
	if v == nil {
		return errors.New("server context not set")
	}

	serverCtxPtr := v.(*Context)
	*serverCtxPtr = *serverCtx

	return nil
}

// interceptConfigs parses and updates a Tendermint configuration file or
// creates a new one and saves it. It also parses and saves the application
// configuration file. The Tendermint configuration file is parsed given a root
// Viper object, whereas the application is parsed with the private package-aware
// viperCfg object.
func interceptConfigs(ctx *Context, rootViper *viper.Viper) (*tmcfg.Config, error) {
	rootDir := rootViper.GetString(flags.FlagHome)
	configPath := filepath.Join(rootDir, "config")
	configFile := filepath.Join(configPath, "config.toml")

	conf := tmcfg.DefaultConfig()

	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		tmcfg.EnsureRoot(rootDir)

		if err = conf.ValidateBasic(); err != nil {
			return nil, fmt.Errorf("error in config file: %v", err)
		}

		conf.RPC.PprofListenAddress = "localhost:6060"
		conf.P2P.RecvRate = 5120000
		conf.P2P.SendRate = 5120000
		conf.Consensus.TimeoutCommit = 5 * time.Second
		tmcfg.WriteConfigFile(configFile, conf)
	} else {
		rootViper.SetConfigType("toml")
		rootViper.SetConfigName("config")
		rootViper.AddConfigPath(configPath)
		if err := rootViper.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("failed to read in app.toml: %w", err)
		}

		if err := rootViper.Unmarshal(conf); err != nil {
			return nil, err
		}
	}

	conf.SetRoot(rootDir)

	appConfigFilePath := filepath.Join(configPath, "app.toml")
	if _, err := os.Stat(appConfigFilePath); os.IsNotExist(err) {
		appConf, err := config.ParseConfig(ctx.Viper)
		if err != nil {
			return nil, fmt.Errorf("failed to parse app.toml: %w", err)
		}

		config.WriteConfigFile(appConfigFilePath, appConf)
	}

	ctx.Viper.SetConfigType("toml")
	ctx.Viper.SetConfigName("app")
	ctx.Viper.AddConfigPath(configPath)
	if err := ctx.Viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read in app.toml: %w", err)
	}

	return conf, nil
}

// add server commands
func AddCommands(rootCmd *cobra.Command, defaultNodeHome string, appCreator types.AppCreator, appExport types.AppExporter) {
	tendermintCmd := &cobra.Command{
		Use:   "tendermint",
		Short: "Tendermint subcommands",
	}

	tendermintCmd.AddCommand(
		ShowNodeIDCmd(),
		ShowValidatorCmd(),
		ShowAddressCmd(),
		VersionCmd(),
	)

	rootCmd.AddCommand(
		StartCmd(appCreator, defaultNodeHome),
		UnsafeResetAllCmd(),
		flags.LineBreak,
		tendermintCmd,
		ExportCmd(appExport, defaultNodeHome),
		flags.LineBreak,
		version.NewVersionCommand(),
	)
}

// https://stackoverflow.com/questions/23558425/how-do-i-get-the-local-ip-address-in-go
// TODO there must be a better way to get external IP
func ExternalIP() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	for _, iface := range ifaces {
		if skipInterface(iface) {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return "", err
		}

		for _, addr := range addrs {
			ip := addrToIP(addr)
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			return ip.String(), nil
		}
	}
	return "", errors.New("are you connected to the network?")
}

// TrapSignal traps SIGINT and SIGTERM and terminates the server correctly.
func TrapSignal(cleanupFunc func()) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigs

		if cleanupFunc != nil {
			cleanupFunc()
		}
		exitCode := 128

		switch sig {
		case syscall.SIGINT:
			exitCode += int(syscall.SIGINT)
		case syscall.SIGTERM:
			exitCode += int(syscall.SIGTERM)
		}

		os.Exit(exitCode)
	}()
}

// WaitForQuitSignals waits for SIGINT and SIGTERM and returns.
func WaitForQuitSignals() ErrorCode {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigs
	return ErrorCode{Code: int(sig.(syscall.Signal)) + 128}
}

func skipInterface(iface net.Interface) bool {
	if iface.Flags&net.FlagUp == 0 {
		return true // interface down
	}

	if iface.Flags&net.FlagLoopback != 0 {
		return true // loopback interface
	}

	return false
}

func addrToIP(addr net.Addr) net.IP {
	var ip net.IP

	switch v := addr.(type) {
	case *net.IPNet:
		ip = v.IP
	case *net.IPAddr:
		ip = v.IP
	}
	return ip
}

func openDB(rootDir string) (dbm.DB, error) {
	dataDir := filepath.Join(rootDir, "data")
	return sdk.NewLevelDB("application", dataDir)
}

func openTraceWriter(traceWriterFile string) (w io.Writer, err error) {
	if traceWriterFile == "" {
		return
	}
	return os.OpenFile(
		traceWriterFile,
		os.O_WRONLY|os.O_APPEND|os.O_CREATE,
		0666,
	)
}
