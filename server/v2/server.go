package serverv2

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"

	"cosmossdk.io/core/transaction"
	"cosmossdk.io/log"
)

// ServerComponent is a server module that can be started and stopped.
type ServerComponent[T transaction.Tx] interface {
	Name() string

	Start(context.Context) error
	Stop(context.Context) error
	Init(AppI[T], *viper.Viper, log.Logger) error
}

// HasStartFlags is a server module that has start flags.
type HasStartFlags interface {
	// StartCmdFlags returns server start flags.
	// Those flags should be prefixed with the server name.
	// They are then merged with the server config in one viper instance.
	StartCmdFlags() *pflag.FlagSet
}

// HasConfig is a server module that has a config.
type HasConfig interface {
	Config() any
}

// HasCLICommands is a server module that has CLI commands.
type HasCLICommands interface {
	CLICommands() CLIConfig
}

// CLIConfig defines the CLI configuration for a module server.
type CLIConfig struct {
	// Commands defines the main command of a module server.
	Commands []*cobra.Command
	// Queries defines the query commands of a module server.
	// Those commands are meant to be added in the root query command.
	Queries []*cobra.Command
	// Txs defines the tx commands of a module server.
	// Those commands are meant to be added in the root tx command.
	Txs []*cobra.Command
}

var _ ServerComponent[transaction.Tx] = (*Server[transaction.Tx])(nil)

type Server[T transaction.Tx] struct {
	logger     log.Logger
	components []ServerComponent[T]
}

func NewServer[T transaction.Tx](
	logger log.Logger,
	components ...ServerComponent[T],
) *Server[T] {
	return &Server[T]{
		logger:     logger,
		components: components,
	}
}

func (s *Server[T]) Name() string {
	return "server"
}

// Start starts all components concurrently.
func (s *Server[T]) Start(ctx context.Context) error {
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

	<-ctx.Done()

	return nil
}

// Stop stops all components concurrently.
func (s *Server[T]) Stop(ctx context.Context) error {
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
func (s *Server[T]) CLICommands() CLIConfig {
	compart := func(name string, cmds ...*cobra.Command) *cobra.Command {
		if len(cmds) == 1 && strings.HasPrefix(cmds[0].Use, name) {
			return cmds[0]
		}

		subCmd := &cobra.Command{
			Use:   name,
			Short: fmt.Sprintf("Commands from the %s server component", name),
		}
		subCmd.AddCommand(cmds...)

		return subCmd
	}

	commands := CLIConfig{}
	for _, mod := range s.components {
		if climod, ok := mod.(HasCLICommands); ok {
			srvCmd := climod.CLICommands()

			if len(srvCmd.Commands) > 0 {
				commands.Commands = append(commands.Commands, compart(mod.Name(), srvCmd.Commands...))
			}

			if len(srvCmd.Txs) > 0 {
				commands.Txs = append(commands.Txs, compart(mod.Name(), srvCmd.Txs...))
			}

			if len(srvCmd.Queries) > 0 {
				commands.Queries = append(commands.Queries, compart(mod.Name(), srvCmd.Queries...))
			}
		}
	}

	return commands
}

// Configs returns all configs of all server components.
func (s *Server[T]) Configs() map[string]any {
	cfgs := make(map[string]any)
	for _, mod := range s.components {
		if configmod, ok := mod.(HasConfig); ok {
			cfg := configmod.Config()
			cfgs[mod.Name()] = cfg
		}
	}

	return cfgs
}

// Init initializes all server components with the provided application, configuration, and logger.
// It returns an error if any component fails to initialize.
func (s *Server[T]) Init(appI AppI[T], v *viper.Viper, logger log.Logger) error {
	var components []ServerComponent[T]
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
func (s *Server[T]) WriteConfig(configPath string) error {
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
		// undocumented interface to write the component default config in another file than app.toml
		// it is used by cometbft for backward compatibility
		// it should not be used by other components
		if mod, ok := component.(interface{ WriteCustomConfigAt(string) error }); ok {
			if err := mod.WriteCustomConfigAt(configPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// StartFlags returns all flags of all server components.
func (s *Server[T]) StartFlags() []*pflag.FlagSet {
	flags := []*pflag.FlagSet{}
	for _, mod := range s.components {
		if startmod, ok := mod.(HasStartFlags); ok {
			flags = append(flags, startmod.StartCmdFlags())
		}
	}

	return flags
}
