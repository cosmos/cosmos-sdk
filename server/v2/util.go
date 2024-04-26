package serverv2

import (
	"context"
	"os"

	cmtcfg "github.com/cometbft/cometbft/config"

	"github.com/spf13/viper"

	"cosmossdk.io/log"

	"fmt"
	"path"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"

	"github.com/spf13/pflag"
)

// ServerContextKey defines the context key used to retrieve a server.Context from
// a command's Context.
const ContextKey = sdk.ContextKey("server.context")

// Context server context
type Context struct {
	Viper  *viper.Viper
	Config *cmtcfg.Config
	Logger log.Logger
}

// InterceptConfigsPreRunHandler is identical to InterceptConfigsAndCreateContext
// except it also sets the server context on the command and the server logger.
func InterceptConfigsPreRunHandler(cmd *cobra.Command, customAppConfigTemplate string, customAppConfig interface{}, cmtConfig *cmtcfg.Config) error {
	viper, config, err := InterceptConfigsAndCreateContext(viper.New(), cmd, customAppConfigTemplate, customAppConfig, cmtConfig)
	if err != nil {
		return err
	}

	// overwrite default server logger
	logger, err := NewLogger(viper, cmd.OutOrStdout())
	if err != nil {
		return err
	}

	// set server context
	return SetCmdServerContext(cmd, &Context{
		Viper:  viper,
		Config: config,
		Logger: logger,
	})
}

// InterceptConfigsAndCreateContext performs a pre-run function for the root daemon
// application command. It will create a Viper literal and a default server
// Context. The server CometBFT configuration will either be read and parsed
// or created and saved to disk, where the server Context is updated to reflect
// the CometBFT configuration. It takes custom app config template and config
// settings to create a custom CometBFT configuration. If the custom template
// is empty, it uses default-template provided by the server. The Viper literal
// is used to read and parse the application configuration. Command handlers can
// fetch the server Context to get the CometBFT configuration or to get access
// to Viper.
func InterceptConfigsAndCreateContext(viper *viper.Viper, cmd *cobra.Command, customAppConfigTemplate string, customAppConfig interface{}, cmtConfig *cmtcfg.Config) (*viper.Viper, *cmtcfg.Config, error) {
	// serverCtx := NewDefaultContext()

	// Get the executable name and configure the viper instance so that environmental
	// variables are checked based off that name. The underscore character is used
	// as a separator.
	executableName, err := os.Executable()
	if err != nil {
		return nil, nil, err
	}

	basename := path.Base(executableName)

	// configure the viper instance
	if err := viper.BindPFlags(cmd.Flags()); err != nil {
		return nil, nil, err
	}
	if err := viper.BindPFlags(cmd.PersistentFlags()); err != nil {
		return nil, nil, err
	}

	viper.SetEnvPrefix(basename)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	viper.AutomaticEnv()

	// intercept configuration files, using both Viper instances separately
	config, err := interceptConfigs(viper, customAppConfigTemplate, customAppConfig, cmtConfig)
	if err != nil {
		return nil, nil, err
	}

	// return value is a CometBFT configuration object
	if err = bindFlags(basename, cmd, viper); err != nil {
		return nil, nil, err
	}

	return viper, config, nil
}

// interceptConfigs parses and updates a CometBFT configuration file or
// creates a new one and saves it. It also parses and saves the application
// configuration file. The CometBFT configuration file is parsed given a root
// Viper object, whereas the application is parsed with the private package-aware
// viperCfg object.
func interceptConfigs(rootViper *viper.Viper, customAppTemplate string, customConfig interface{}, cmtConfig *cmtcfg.Config) (*cmtcfg.Config, error) {
	// TODO: Should we keep config same as server v1
	return cmtConfig, nil
}

func bindFlags(basename string, cmd *cobra.Command, v *viper.Viper) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("bindFlags failed: %v", r)
		}
	}()

	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		// Environment variables can't have dashes in them, so bind them to their equivalent
		// keys with underscores, e.g. --favorite-color to STING_FAVORITE_COLOR
		err = v.BindEnv(f.Name, fmt.Sprintf("%s_%s", basename, strings.ToUpper(strings.ReplaceAll(f.Name, "-", "_"))))
		if err != nil {
			panic(err)
		}

		err = v.BindPFlag(f.Name, f)
		if err != nil {
			panic(err)
		}

		// Apply the viper config value to the flag when the flag is not set and
		// viper has a value.
		if !f.Changed && v.IsSet(f.Name) {
			val := v.Get(f.Name)
			err = cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val))
			if err != nil {
				panic(err)
			}
		}
	})

	return err
}

// GetServerContextFromCmd returns a Context from a command or an empty Context
// if it has not been set.
func GetServerContextFromCmd(cmd *cobra.Command) *Context {
	if v := cmd.Context().Value(ServerContextKey); v != nil {
		serverCtxPtr := v.(*Context)
		return serverCtxPtr
	}

	return &Context{
		Viper:  viper.New(),
		Config: cmtcfg.DefaultConfig(),
		Logger: log.NewLogger(os.Stdout),
	}
}

// SetCmdServerContext sets a command's Context value to the provided argument.
// If the context has not been set, set the given context as the default.
func SetCmdServerContext(cmd *cobra.Command, serverCtx *Context) error {
	var cmdCtx context.Context

	if cmd.Context() == nil {
		cmdCtx = context.Background()
	} else {
		cmdCtx = cmd.Context()
	}

	cmd.SetContext(context.WithValue(cmdCtx, ServerContextKey, serverCtx))

	return nil
}
