package serverv2

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	cmtcfg "github.com/cometbft/cometbft/config"
	"github.com/pelletier/go-toml/v2"
	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"

	"cosmossdk.io/log"
)

// ServerModule is a server module that can be started and stopped.
type ServerModule interface {
	Name() string

	Start(context.Context) error
	Stop(context.Context) error
}

// HasCLICommands is a server module that has CLI commands.
type HasCLICommands interface {
	CLICommands() CLIConfig
}

// HasConfig is a server module that has a config.
type HasConfig interface {
	Config() any
}

var _ ServerModule = (*Server)(nil)

// Configs returns a viper instance of the config file
// Currently just set comet config to viper to init comet
// TODO: should set both app.toml?
func ReadConfig(configPath string) (*viper.Viper, error) {
	v := viper.New()
	cmtConfigFile := filepath.Join(configPath, "config.toml")
	v.SetConfigFile(cmtConfigFile)
	v.SetConfigType("toml")
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config: %s: %w", configPath, err)
	}
	v.WatchConfig()

	config := cmtcfg.DefaultConfig()
	v.Unmarshal(config)
	return v, nil
}

type Server struct {
	logger  log.Logger
	modules []ServerModule
}

func NewServer(logger log.Logger, modules ...ServerModule) *Server {
	return &Server{
		logger:  logger,
		modules: modules,
	}
}

func (s *Server) Name() string {
	return "server"
}

// Start starts all modules concurrently.
func (s *Server) Start(ctx context.Context) error {
	s.logger.Info("starting servers...")

	g, ctx := errgroup.WithContext(ctx)
	for _, mod := range s.modules {
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

// Stop stops all modules concurrently.
func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("stopping servers...")

	g, ctx := errgroup.WithContext(ctx)
	for _, mod := range s.modules {
		mod := mod
		g.Go(func() error {
			return mod.Stop(ctx)
		})
	}

	return g.Wait()
}

// CLICommands returns all CLI commands of all modules.
func (s *Server) CLICommands() CLIConfig {
	commands := CLIConfig{}
	for _, mod := range s.modules {
		if climod, ok := mod.(HasCLICommands); ok {
			commands.Commands = append(commands.Commands, climod.CLICommands().Commands...)
			commands.Queries = append(commands.Queries, climod.CLICommands().Queries...)
			commands.Txs = append(commands.Txs, climod.CLICommands().Txs...)
		}
	}

	return commands
}

// Configs returns all configs of all server modules.
func (s *Server) Configs() map[string]any {
	cfgs := make(map[string]any)
	for _, mod := range s.modules {
		if configmod, ok := mod.(HasConfig); ok {
			cfg := configmod.Config()
			cfgs[mod.Name()] = cfg
		}
	}

	return cfgs
}

// WriteConfig writes the config to the given path.
// Note: it does not use viper.WriteConfigAs because we do not want to store flag values in the config.
func (s *Server) WriteConfig(configPath string) error {
	cfgs := s.Configs()
	b, err := toml.Marshal(cfgs)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, b, 0o600); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}
