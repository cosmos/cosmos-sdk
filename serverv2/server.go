package serverv2

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/fsnotify/fsnotify"
	"github.com/pelletier/go-toml/v2"
	"github.com/spf13/cobra"
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
	CLICommands() []*cobra.Command
}

// HasConfig is a server module that has a config.
type HasConfig interface {
	// Config returns the config of the module.
	// The returned values are the config struct and the viper instance containing the config.
	Config() (any, *viper.Viper)
}

// HasReload is a server module that can be reloaded.
type HasReload interface {
	Reload(context.Context) error
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

// Reload reloads a module.
func (s *Server) Reload(ctx context.Context, moduleName string) error {
	for _, mod := range s.modules {
		if mod.Name() == moduleName || moduleName == s.Name() {
			if reloadmod, ok := mod.(HasReload); ok {
				s.logger.Debug(fmt.Sprintf("reloading %s server...", moduleName))
				return reloadmod.Reload(ctx)
			}

			return errors.New("module does not support reload")
		}
	}

	return errors.New("module not found")
}

// CLICommands returns all CLI commands of all modules.
func (s *Server) CLICommands() []*cobra.Command {
	var commands []*cobra.Command
	for _, mod := range s.modules {
		if climod, ok := mod.(HasCLICommands); ok {
			commands = append(commands, climod.CLICommands()...)
		}
	}

	return commands
}

// Configs returns the merged config of all modules.
func (s *Server) Config(configPath string) (*viper.Viper, error) {
	v := viper.New()
	v.SetConfigFile(configPath)
	v.SetConfigName("app")
	v.SetConfigType("toml")
	v.ReadInConfig()
	v.OnConfigChange(func(e fsnotify.Event) {
		if e.Op == fsnotify.Write {
			srvName := s.Name()

			s.logger.Info("config file changed", "path", e.Name)
			// TODO(@julienrbrt): find a propoer way to reload a module independently of the other modules.
			if err := s.Reload(context.Background(), srvName); err != nil {
				s.logger.Error(fmt.Sprintf("failed to reload %s server", srvName), "err", err)
			}
		}
	})
	v.WatchConfig()

	var err error
	for _, mod := range s.modules {
		if configmod, ok := mod.(HasConfig); ok {
			_, nv := configmod.Config()
			err = errors.Join(err, v.MergeConfigMap(nv.AllSettings()))
		}
	}

	return v, err
}

// WriteConfig writes the config to the given path.
// Note: it does not use viper.WriteConfigAs because we do not want to store flag values in the config.
func (s *Server) WriteConfig(configPath string) error {
	cfgs := make(map[string]any)
	for _, mod := range s.modules {
		if configmod, ok := mod.(HasConfig); ok {
			cfg, _ := configmod.Config()
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
