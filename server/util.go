package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	tmcfg "github.com/tendermint/tendermint/config"
	tmcli "github.com/tendermint/tendermint/libs/cli"
	tmflags "github.com/tendermint/tendermint/libs/cli/flags"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server/config"
	"github.com/cosmos/cosmos-sdk/version"
)

// DONTCOVER

// server context
type Context struct {
	Viper  *viper.Viper
	Config *tmcfg.Config
	Logger log.Logger
}

func NewDefaultContext() *Context {
	return NewContext(viper.New(), tmcfg.DefaultConfig(), log.NewTMLogger(log.NewSyncWriter(os.Stdout)))
}

func NewContext(v *viper.Viper, config *tmcfg.Config, logger log.Logger) *Context {
	return &Context{v, config, logger}
}

// PersistentPreRunEFn returns a PersistentPreRunE function for the root daemon
// application command. The provided context is typically the default context,
// where the logger and config are set based on the execution of parsing or
// creating a new Tendermint configuration file (config.toml). The provided
// viper object must be created at the root level and have all necessary flags,
// defined by Tendermint, bound to it.
func PersistentPreRunEFn(ctx *Context) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		rootViper := viper.New()
		rootViper.BindPFlags(cmd.Flags())
		rootViper.BindPFlags(cmd.PersistentFlags())

		if cmd.Name() == version.Cmd.Name() {
			return nil
		}

		config, err := interceptConfigs(ctx, rootViper)
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

		logger = logger.With("module", "main")
		ctx.Config = config
		ctx.Logger = logger

		return nil
	}
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

		conf.ProfListenAddress = "localhost:6060"
		conf.P2P.RecvRate = 5120000
		conf.P2P.SendRate = 5120000
		conf.TxIndex.IndexAllKeys = true
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
func AddCommands(ctx *Context, cdc codec.JSONMarshaler, rootCmd *cobra.Command, appCreator AppCreator, appExport AppExporter) {
	rootCmd.PersistentFlags().String("log_level", ctx.Config.LogLevel, "Log level")

	tendermintCmd := &cobra.Command{
		Use:   "tendermint",
		Short: "Tendermint subcommands",
	}

	tendermintCmd.AddCommand(
		ShowNodeIDCmd(ctx),
		ShowValidatorCmd(ctx),
		ShowAddressCmd(ctx),
		VersionCmd(ctx),
	)

	rootCmd.AddCommand(
		StartCmd(ctx, cdc, appCreator),
		UnsafeResetAllCmd(ctx),
		flags.LineBreak,
		tendermintCmd,
		ExportCmd(ctx, cdc, appExport),
		flags.LineBreak,
		version.Cmd,
	)
}

// InsertKeyJSON inserts a new JSON field/key with a given value to an existing
// JSON message. An error is returned if any serialization operation fails.
//
// NOTE: The ordering of the keys returned as the resulting JSON message is
// non-deterministic, so the client should not rely on key ordering.
func InsertKeyJSON(cdc codec.JSONMarshaler, baseJSON []byte, key string, value json.RawMessage) ([]byte, error) {
	var jsonMap map[string]json.RawMessage

	if err := cdc.UnmarshalJSON(baseJSON, &jsonMap); err != nil {
		return nil, err
	}

	jsonMap[key] = value
	bz, err := codec.MarshalJSONIndent(cdc, jsonMap)

	return json.RawMessage(bz), err
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
