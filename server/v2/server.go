package serverv2

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"

	"cosmossdk.io/core/transaction"
	"cosmossdk.io/log"
	"github.com/pelletier/go-toml/v2"
)

// ServerComponent is a server module that can be started and stopped.
type ServerComponent[T transaction.Tx] interface {
	Name() string

	Start(context.Context) error
	Stop(context.Context) error
	Init(App[T], *viper.Viper, log.Logger) (ServerComponent[T], error)
}

// HasCLICommands is a server module that has CLI commands.
type HasCLICommands interface {
	CLICommands() CLIConfig
}

// HasConfig is a server module that has a config.
type HasConfig interface {
	Config() any
}

// HasStartFlags is a server module that has start flags.
type HasStartFlags interface {
	StartCmdFlags() *pflag.FlagSet
}

var _ ServerComponent[transaction.Tx] = (*Server)(nil)

// Configs returns a viper instance of the config file
func ReadConfig(configPath string) (*viper.Viper, error) {
	v := viper.New()

	v.SetConfigType("toml")
	v.SetConfigName("config")
	v.AddConfigPath(configPath)
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config: %s: %w", configPath, err)
	}

	v.SetConfigName("app")
	if err := v.MergeInConfig(); err != nil {
		return nil, fmt.Errorf("failed to merge configuration: %w", err)
	}

	v.WatchConfig()

	return v, nil
}

type Server struct {
	logger     log.Logger
	components []ServerComponent[transaction.Tx]
}

func NewServer(logger log.Logger, components ...ServerComponent[transaction.Tx]) *Server {
	return &Server{
		logger:  logger,
		components: components,
	}
}

func (s *Server) Name() string {
	return "server"
}

// Start starts all components concurrently.
func (s *Server) Start(ctx context.Context) error {
	s.logger.Info("starting servers...")

	g, ctx := errgroup.WithContext(ctx)
	for _, mod := range s.components {
		mod := mod
		g.Go(func() error {
			return mod.Start(ctx)
		})
	}

	if err := g.Wait(); err != nil {
		return fmt.Errorf("failed to start servers: %w", err)
	}

	serverCfg := ctx.Value(ServerContextKey).(Config)
	if serverCfg.StartBlock {
		<-ctx.Done()
	}

	return nil
}

// Stop stops all components concurrently.
func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("stopping servers...")

	g, ctx := errgroup.WithContext(ctx)
	for _, mod := range s.components {
		mod := mod
		g.Go(func() error {
			return mod.Stop(ctx)
		})
	}

	return g.Wait()
}

// CLICommands returns all CLI commands of all components.
func (s *Server) CLICommands() CLIConfig {
	commands := CLIConfig{}
	for _, mod := range s.components {
		if climod, ok := mod.(HasCLICommands); ok {
			commands.Commands = append(commands.Commands, climod.CLICommands().Commands...)
			commands.Queries = append(commands.Queries, climod.CLICommands().Queries...)
			commands.Txs = append(commands.Txs, climod.CLICommands().Txs...)
		}
	}

	return commands
}

// Configs returns all configs of all server components.
func (s *Server) Configs() map[string]any {
	cfgs := make(map[string]any)
	for _, mod := range s.components {
		if configmod, ok := mod.(HasConfig); ok {
			cfg := configmod.Config()
			cfgs[mod.Name()] = cfg
		}
	}

	return cfgs
}

// Configs returns all configs of all server components.
func (s *Server) Init(appI App[transaction.Tx], v *viper.Viper, logger log.Logger) (ServerComponent[transaction.Tx], error) {
	var components []ServerComponent[transaction.Tx]
	for _, mod := range s.components {
		mod := mod
		module, err := mod.Init(appI, v, logger)
		if err != nil {
			return nil, err
		}
		components = append(components, module)
	}
	s.components = components

	return s, nil
}

// WriteConfig writes the config to the given path.
// Note: it does not use viper.WriteConfigAs because we do not want to store flag values in the config.
func (s *Server) WriteConfig(configPath string) error {
	cfgs := s.Configs()
	b, err := toml.Marshal(cfgs)
	if err != nil {
		return err
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		if err := os.MkdirAll(configPath, os.ModePerm); err != nil {
			return err
		}
	}

	if err := os.WriteFile(filepath.Join(configPath, "app.toml"), b, 0o600); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	for _, component := range s.components {
		if mod, ok := component.(interface{ WriteDefaultConfigAt(string) error }); ok {
			if err := mod.WriteDefaultConfigAt(configPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// Flags returns all flags of all server components.
func (s *Server) StartFlags() []*pflag.FlagSet {
	flags := []*pflag.FlagSet{}
	for _, mod := range s.components {
		if startmod, ok := mod.(HasStartFlags); ok {
			flags = append(flags, startmod.StartCmdFlags())
		}
	}

	return flags
}
