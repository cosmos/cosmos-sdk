package store

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"cosmossdk.io/core/server"
	"cosmossdk.io/core/transaction"
	serverv2 "cosmossdk.io/server/v2"
	storev2 "cosmossdk.io/store/v2"
	"cosmossdk.io/store/v2/root"
)

var (
	_ serverv2.ServerComponent[transaction.Tx] = (*Server[transaction.Tx])(nil)
	_ serverv2.HasConfig                       = (*Server[transaction.Tx])(nil)
	_ serverv2.HasCLICommands                  = (*Server[transaction.Tx])(nil)
)

const ServerName = "store"

// Server manages store config and contains prune & snapshot commands
type Server[T transaction.Tx] struct {
	config *root.Config
	store  storev2.Backend
}

func New[T transaction.Tx](store storev2.Backend, cfg server.ConfigMap) (*Server[T], error) {
	config, err := UnmarshalConfig(cfg)
	if err != nil {
		return nil, err
	}
	return &Server[T]{
		store:  store,
		config: config,
	}, nil
}

func (s *Server[T]) Name() string {
	return ServerName
}

func (s *Server[T]) Start(context.Context) error {
	return nil
}

func (s *Server[T]) Stop(context.Context) error {
	return nil
}

func (s *Server[T]) CLICommands() serverv2.CLIConfig {
	return serverv2.CLIConfig{
		Commands: []*cobra.Command{
			s.PrunesCmd(),
			s.ExportSnapshotCmd(),
			s.DeleteSnapshotCmd(),
			s.ListSnapshotsCmd(),
			s.DumpArchiveCmd(),
			s.LoadArchiveCmd(),
			s.RestoreSnapshotCmd(),
		},
	}
}

func (s *Server[T]) Config() any {
	if s.config == nil || s.config.AppDBBackend == "" {
		return root.DefaultConfig()
	}

	return s.config
}

// UnmarshalConfig unmarshals the store config from the given map.
// If the config is not found in the map, the default config is returned.
// If the home directory is found in the map, it sets the home directory in the config.
// An empty home directory *is* permitted at this stage, but attempting to build
// the store with an empty home directory will fail.
func UnmarshalConfig(cfg map[string]any) (*root.Config, error) {
	config := root.DefaultConfig()
	if err := serverv2.UnmarshalSubConfig(cfg, ServerName, config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal store config: %w", err)
	}
	home := cfg[serverv2.FlagHome]
	if home != nil {
		config.Home = home.(string)
	}
	return config, nil
}
