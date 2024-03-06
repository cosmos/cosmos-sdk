package serverv2

import (
	"context"
	"fmt"
	"os"

	"github.com/fsnotify/fsnotify"
	"github.com/pelletier/go-toml/v2"
	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"

	"cosmossdk.io/log"
)

// Module is a server module that can be started and stopped.
type Module interface {
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
	// Config returns the config of the module.
	Config() any
}

var _ Module = (*Server)(nil)

type Server struct {
	logger  log.Logger
	modules []Module
}

func NewServer(logger log.Logger, modules ...Module) *Server {
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
			commands.Command = append(commands.Command, climod.CLICommands().Command...)
			commands.Query = append(commands.Query, climod.CLICommands().Query...)
			commands.Tx = append(commands.Tx, climod.CLICommands().Tx...)
		}
	}

	return commands
}

// Configs returns a viper instance of the config file
func (s *Server) Config(configPath string) (*viper.Viper, error) {
	v := viper.New()
	v.SetConfigFile(configPath)
	v.SetConfigType("toml")

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config: %s: %w", configPath, err)
	}

	v.OnConfigChange(func(e fsnotify.Event) {
		if e.Op == fsnotify.Write {
			srvName := s.Name()
			s.logger.Info("config file changed", "path", e.Name, "server", srvName)
			// 		// TODO(@julienrbrt): find a propoer way to reload a module independently of the other modules.
			// 		if err := s.Reload(context.Background(), srvName); err != nil {
			// 			s.logger.Error(fmt.Sprintf("failed to reload %s server", srvName), "err", err)
			// 		}
		}
	})
	v.WatchConfig()

	return v, nil
}

// WriteConfig writes the config to the given path.
// Note: it does not use viper.WriteConfigAs because we do not want to store flag values in the config.
func (s *Server) WriteConfig(configPath string) error {
	cfgs := make(map[string]any)
	for _, mod := range s.modules {
		if configmod, ok := mod.(HasConfig); ok {
			cfg := configmod.Config()
			cfgs[mod.Name()] = cfg
		}
	}

	b, err := toml.Marshal(cfgs)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, b, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}
