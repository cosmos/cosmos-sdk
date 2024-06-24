package serverv2

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"

	"cosmossdk.io/core/transaction"
	"cosmossdk.io/log"
)

// ServerComponent is a server module that can be started and stopped.
type ServerComponent[AppT AppI[T], T transaction.Tx] interface {
	Name() string

	Start(context.Context) error
	Stop(context.Context) error
	Init(AppT, *viper.Viper, log.Logger) error
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

var _ ServerComponent[AppI[transaction.Tx], transaction.Tx] = (*Server[AppI[transaction.Tx], transaction.Tx])(nil)

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

type Server[AppT AppI[T], T transaction.Tx] struct {
	logger     log.Logger
	components []ServerComponent[AppT, T]
}

func NewServer[AppT AppI[T], T transaction.Tx](
	logger log.Logger, components ...ServerComponent[AppT, T],
) *Server[AppT, T] {
	return &Server[AppT, T]{
		logger:     logger,
		components: components,
	}
}

func (s *Server[AppT, T]) Name() string {
	return "server"
}

// Start starts all components concurrently.
func (s *Server[AppT, T]) Start(ctx context.Context) error {
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
func (s *Server[AppT, T]) Stop(ctx context.Context) error {
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
func (s *Server[AppT, T]) CLICommands() CLIConfig {
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
func (s *Server[AppT, T]) Configs() map[string]any {
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
func (s *Server[AppT, T]) Init(appI AppT, v *viper.Viper, logger log.Logger) error {
	var components []ServerComponent[AppT, T]
	for _, mod := range s.components {
		mod := mod
		if err := mod.Init(appI, v, logger); err != nil {
			return err
		}

		components = append(components, mod)
	}

	s.components = components
	return nil
}

// WriteConfig writes the config to the given path.
// Note: it does not use viper.WriteConfigAs because we do not want to store flag values in the config.
func (s *Server[AppT, T]) WriteConfig(configPath string) error {
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
func (s *Server[AppT, T]) StartFlags() []*pflag.FlagSet {
	flags := []*pflag.FlagSet{}
	for _, mod := range s.components {
		if startmod, ok := mod.(HasStartFlags); ok {
			flags = append(flags, startmod.StartCmdFlags())
		}
	}

	return flags
}
