package cometbft

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	cmtcfg "github.com/cometbft/cometbft/config"
	cmtjson "github.com/cometbft/cometbft/libs/json"
	"github.com/cometbft/cometbft/node"
	"github.com/cometbft/cometbft/p2p"
	pvm "github.com/cometbft/cometbft/privval"
	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	"github.com/cometbft/cometbft/rpc/client/local"
	cmtversion "github.com/cometbft/cometbft/version"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/protobuf/encoding/protojson"
	"sigs.k8s.io/yaml"

	"cosmossdk.io/log"
	serverv2 "cosmossdk.io/server/v2"
	"cosmossdk.io/server/v2/cometbft/client/rpc"
	"cosmossdk.io/server/v2/cometbft/flags"
	auth "cosmossdk.io/x/auth/client/cli"
	"github.com/cosmos/cosmos-sdk/client"

	pruningtypes "cosmossdk.io/store/pruning/types"
	storev2 "cosmossdk.io/store/v2"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/spf13/cast"
)

func (s *CometBFTServer[T]) rpcClient(cmd *cobra.Command) (rpc.CometRPC, error) {
	if s.config.Standalone {
		client, err := rpchttp.New(client.GetConfigFromCmd(cmd).RPC.ListenAddress)
		if err != nil {
			return nil, err
		}
		return client, nil
	}

	if s.Node == nil || cmd.Flags().Changed(flags.FlagNode) {
		rpcURI, err := cmd.Flags().GetString(flags.FlagNode)
		if err != nil {
			return nil, err
		}
		if rpcURI != "" {
			return rpchttp.New(rpcURI)
		}
	}

	return local.New(s.Node), nil
}

// StatusCommand returns the command to return the status of the network.
func (s *CometBFTServer[T]) StatusCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Query remote node for status",
		RunE: func(cmd *cobra.Command, _ []string) error {
			rpcclient, err := s.rpcClient(cmd)
			if err != nil {
				return err
			}

			status, err := rpcclient.Status(cmd.Context())
			if err != nil {
				return err
			}

			output, err := cmtjson.Marshal(status)
			if err != nil {
				return err
			}

			return printOutput(cmd, output)
		},
	}

	cmd.Flags().StringP(flags.FlagNode, "n", "tcp://localhost:26657", "Node to connect to")
	cmd.Flags().StringP(flags.FlagOutput, "o", "json", "Output format (text|json)")

	return cmd
}

// ShowNodeIDCmd - ported from CometBFT, dump node ID to stdout
func (s *CometBFTServer[T]) ShowNodeIDCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show-node-id",
		Short: "Show this node's ID",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmtConfig := client.GetConfigFromCmd(cmd)
			nodeKey, err := p2p.LoadNodeKey(cmtConfig.NodeKeyFile())
			if err != nil {
				return err
			}

			cmd.Println(nodeKey.ID())
			return nil
		},
	}
}

// ShowValidatorCmd - ported from CometBFT, show this node's validator info
func (s *CometBFTServer[T]) ShowValidatorCmd() *cobra.Command {
	cmd := cobra.Command{
		Use:   "show-validator",
		Short: "Show this node's CometBFT validator info",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := client.GetConfigFromCmd(cmd)
			privValidator := pvm.LoadFilePV(cfg.PrivValidatorKeyFile(), cfg.PrivValidatorStateFile())
			pk, err := privValidator.GetPubKey()
			if err != nil {
				return err
			}

			sdkPK, err := cryptocodec.FromCmtPubKeyInterface(pk)
			if err != nil {
				return err
			}

			cmd.Println(sdkPK) // TODO: figure out if we need the codec here or not, see below

			// clientCtx := client.GetClientContextFromCmd(cmd)
			// bz, err := clientCtx.Codec.MarshalInterfaceJSON(sdkPK)
			// if err != nil {
			// 	return err
			// }

			// cmd.Println(string(bz))
			return nil
		},
	}

	return &cmd
}

// ShowAddressCmd - show this node's validator address
func (s *CometBFTServer[T]) ShowAddressCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show-address",
		Short: "Shows this node's CometBFT validator consensus address",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := client.GetConfigFromCmd(cmd)
			privValidator := pvm.LoadFilePV(cfg.PrivValidatorKeyFile(), cfg.PrivValidatorStateFile())
			// TODO: use address codec?
			valConsAddr := (sdk.ConsAddress)(privValidator.GetAddress())

			cmd.Println(valConsAddr.String())
			return nil
		},
	}

	return cmd
}

// VersionCmd prints CometBFT and ABCI version numbers.
func (s *CometBFTServer[T]) VersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print CometBFT libraries' version",
		Long:  "Print protocols' and libraries' version numbers against which this app has been compiled.",
		RunE: func(cmd *cobra.Command, args []string) error {
			bs, err := yaml.Marshal(&struct {
				CometBFT      string
				ABCI          string
				BlockProtocol uint64
				P2PProtocol   uint64
			}{
				CometBFT:      cmtversion.CMTSemVer,
				ABCI:          cmtversion.ABCIVersion,
				BlockProtocol: cmtversion.BlockProtocol,
				P2PProtocol:   cmtversion.P2PProtocol,
			})
			if err != nil {
				return err
			}

			cmd.Println(string(bs))
			return nil
		},
	}
}

// QueryBlocksCmd returns a command to search through blocks by events.
func (s *CometBFTServer[T]) QueryBlocksCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "blocks",
		Short: "Query for paginated blocks that match a set of events",
		Long: `Search for blocks that match the exact given events where results are paginated.
The events query is directly passed to CometBFT's RPC BlockSearch method and must
conform to CometBFT's query syntax.
Please refer to each module's documentation for the full set of events to query
for. Each module documents its respective events under 'xx_events.md'.
`,
		Example: fmt.Sprintf(
			"$ %s query blocks --query \"message.sender='cosmos1...' AND block.height > 7\" --page 1 --limit 30 --order-by ASC",
			version.AppName,
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			rpcclient, err := s.rpcClient(cmd)
			if err != nil {
				return err
			}

			query, _ := cmd.Flags().GetString(auth.FlagQuery)
			page, _ := cmd.Flags().GetInt(flags.FlagPage)
			limit, _ := cmd.Flags().GetInt(flags.FlagLimit)
			orderBy, _ := cmd.Flags().GetString(auth.FlagOrderBy)

			blocks, err := rpc.QueryBlocks(cmd.Context(), rpcclient, page, limit, query, orderBy)
			if err != nil {
				return err
			}

			bz, err := protojson.Marshal(blocks)
			if err != nil {
				return err
			}

			return printOutput(cmd, bz)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	cmd.Flags().Int(flags.FlagPage, query.DefaultPage, "Query a specific page of paginated results")
	cmd.Flags().Int(flags.FlagLimit, query.DefaultLimit, "Query number of transactions results per page returned")
	cmd.Flags().String(auth.FlagQuery, "", "The blocks events query per CometBFT's query semantics")
	cmd.Flags().String(auth.FlagOrderBy, "", "The ordering semantics (asc|dsc)")
	_ = cmd.MarkFlagRequired(auth.FlagQuery)

	return cmd
}

// QueryBlockCmd implements the default command for a Block query.
func (s *CometBFTServer[T]) QueryBlockCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "block --type=[height|hash] [height|hash]",
		Short: "Query for a committed block by height, hash, or event(s)",
		Long:  "Query for a specific committed block using the CometBFT RPC `block` and `block_by_hash` method",
		Example: strings.TrimSpace(fmt.Sprintf(`
$ %s query block --%s=%s <height>
$ %s query block --%s=%s <hash>
`,
			version.AppName, auth.FlagType, auth.TypeHeight,
			version.AppName, auth.FlagType, auth.TypeHash)),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			typ, _ := cmd.Flags().GetString(auth.FlagType)

			rpcclient, err := s.rpcClient(cmd)
			if err != nil {
				return err
			}

			switch typ {
			case auth.TypeHeight:
				if args[0] == "" {
					return fmt.Errorf("argument should be a block height")
				}

				// optional height
				var height *int64
				if len(args) > 0 {
					height, err = parseOptionalHeight(args[0])
					if err != nil {
						return err
					}
				}

				output, err := rpc.GetBlockByHeight(cmd.Context(), rpcclient, height)
				if err != nil {
					return err
				}

				if output.Header.Height == 0 {
					return fmt.Errorf("no block found with height %s", args[0])
				}

				bz, err := json.Marshal(output)
				if err != nil {
					return err
				}

				return printOutput(cmd, bz)

			case auth.TypeHash:

				if args[0] == "" {
					return fmt.Errorf("argument should be a tx hash")
				}

				// If hash is given, then query the tx by hash.
				output, err := rpc.GetBlockByHash(cmd.Context(), rpcclient, args[0])
				if err != nil {
					return err
				}

				if output.Header.AppHash == nil {
					return fmt.Errorf("no block found with hash %s", args[0])
				}

				bz, err := json.Marshal(output)
				if err != nil {
					return err
				}

				return printOutput(cmd, bz)

			default:
				return fmt.Errorf("unknown --%s value %s", auth.FlagType, typ)
			}
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	cmd.Flags().String(auth.FlagType, auth.TypeHash, fmt.Sprintf("The type to be used when querying tx, can be one of \"%s\", \"%s\"", auth.TypeHeight, auth.TypeHash))

	return cmd
}

// QueryBlockResultsCmd implements the default command for a BlockResults query.
func (s *CometBFTServer[T]) QueryBlockResultsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "block-results [height]",
		Short: "Query for a committed block's results by height",
		Long:  "Query for a specific committed block's results using the CometBFT RPC `block_results` method",
		Args:  cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// clientCtx, err := client.GetClientQueryContext(cmd)
			// if err != nil {
			// 	return err
			// }

			// TODO: we should be able to do this without using client context

			node, err := s.rpcClient(cmd)
			if err != nil {
				return err
			}

			// optional height
			var height *int64
			if len(args) > 0 {
				height, err = parseOptionalHeight(args[0])
				if err != nil {
					return err
				}
			}

			blockRes, err := node.BlockResults(cmd.Context(), height)
			if err != nil {
				return err
			}

			// coretypes.ResultBlockResults doesn't implement proto.Message interface
			// so we can't print it using clientCtx.PrintProto
			// we choose to serialize it to json and print the json instead
			blockResStr, err := json.Marshal(blockRes)
			if err != nil {
				return err
			}

			return printOutput(cmd, blockResStr)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// QueryBlockResultsCmd implements the default command for a BlockResults query.
func (s *CometBFTServer[T]) PrunesCmd(appCreator serverv2.AppCreator[T]) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "prune [pruning-method]",
		Short: "Prune app history states by keeping the recent heights and deleting old heights",
		Long: `Prune app history states by keeping the recent heights and deleting old heights.
The pruning option is provided via the 'pruning' argument or alternatively with '--pruning-keep-recent'

- default: the last 362880 states are kept
- nothing: all historic states will be saved, nothing will be deleted (i.e. archiving node)
- everything: 2 latest states will be kept
- custom: allow pruning options to be manually specified through 'pruning-keep-recent'

Note: When the --app-db-backend flag is not specified, the default backend type is 'goleveldb'.
Supported app-db-backend types include 'goleveldb', 'rocksdb', 'pebbledb'.`,
		Example: fmt.Sprintf("%s prune custom --pruning-keep-recent 100 --app-db-backend 'goleveldb'", version.AppName),
		Args:    cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// bind flags to the Context's Viper so we can get pruning options.
			vp := viper.New()
			if err := vp.BindPFlags(cmd.Flags()); err != nil {
				return err
			}
			if err := vp.BindPFlags(cmd.PersistentFlags()); err != nil {
				return err
			}

			// use the first argument if present to set the pruning method
			if len(args) > 0 {
				vp.Set(serverv2.FlagPruning, args[0])
			} else {
				vp.Set(serverv2.FlagPruning, pruningtypes.PruningOptionDefault)
			}
			pruningOptions, err := getPruningOptionsFromFlags(vp)
			if err != nil {
				return err
			}

			cmd.Printf("get pruning options from command flags, strategy: %v, keep-recent: %v\n",
				pruningOptions.Strategy,
				pruningOptions.KeepRecent,
			)

			logger := log.NewLogger(cmd.OutOrStdout())
			app := appCreator(logger, vp)
			store := app.GetStore()

			rootStore, ok := store.(storev2.RootStore)
			if !ok {
				return fmt.Errorf("currently only support the pruning of rootmulti.Store type")
			}
			latestHeight, err := rootStore.GetLatestVersion()
			if err != nil {
				return err
			}

			// valid heights should be greater than 0.
			if latestHeight <= 0 {
				return fmt.Errorf("the database has no valid heights to prune, the latest height: %v", latestHeight)
			}

			pruningHeight := latestHeight - pruningOptions.KeepRecent
			cmd.Printf("pruning heights up to %v\n", pruningHeight)

			err = rootStore.Prune(pruningHeight)
			if err != nil {
				return err
			}

			cmd.Println("successfully pruned the application root multi stores")
			return nil
		},
	}

	cmd.Flags().String(serverv2.FlagAppDBBackend, "", "The type of database for application and snapshots databases")
	cmd.Flags().Uint64(serverv2.FlagPruningKeepRecent, 0, "Number of recent heights to keep on disk (ignored if pruning is not 'custom')")
	cmd.Flags().Uint64(serverv2.FlagPruningInterval, 10,
		`Height interval at which pruned heights are removed from disk (ignored if pruning is not 'custom'), 
		this is not used by this command but kept for compatibility with the complete pruning options`)

	return cmd
}

func parseOptionalHeight(heightStr string) (*int64, error) {
	h, err := strconv.Atoi(heightStr)
	if err != nil {
		return nil, err
	}

	if h == 0 {
		return nil, nil
	}

	tmp := int64(h)

	return &tmp, nil
}

func (s *CometBFTServer[T]) BootstrapStateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bootstrap-state",
		Short: "Bootstrap CometBFT state at an arbitrary block height using a light client",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := client.GetConfigFromCmd(cmd)
			height, err := cmd.Flags().GetUint64("height")
			if err != nil {
				return err
			}
			if height == 0 {
				height, err = s.App.store.GetLatestVersion()
				if err != nil {
					return err
				}
			}

			// TODO genensis doc provider and apphash
			return node.BootstrapState(cmd.Context(), cfg, cmtcfg.DefaultDBProvider, nil, height, nil)
		},
	}

	cmd.Flags().Int64("height", 0, "Block height to bootstrap state at, if not provided it uses the latest block height in app state")

	return cmd
}

func printOutput(cmd *cobra.Command, out []byte) error {
	// Get flags output
	outFlag, err := cmd.Flags().GetString(flags.FlagOutput) 
	if err != nil {
		return err
	}

	if outFlag == "text" {
		out, err = yaml.JSONToYAML(out)
		if err != nil {
			return err
		}
	}

	writer := cmd.OutOrStdout()
	_, err = writer.Write(out)
	if err != nil {
		return err
	}

	if outFlag != "text" {
		// append new-line for formats besides YAML
		_, err = writer.Write([]byte("\n"))
		if err != nil {
			return err
		}
	}
	return nil
}

func getPruningOptionsFromFlags(v *viper.Viper) (pruningtypes.PruningOptions, error) {
	strategy := strings.ToLower(cast.ToString(v.Get(serverv2.FlagPruning)))

	switch strategy {
	case pruningtypes.PruningOptionDefault, pruningtypes.PruningOptionNothing, pruningtypes.PruningOptionEverything:
		return pruningtypes.NewPruningOptionsFromString(strategy), nil

	case pruningtypes.PruningOptionCustom:
		opts := pruningtypes.NewCustomPruningOptions(
			cast.ToUint64(v.Get(serverv2.FlagPruningKeepRecent)),
			cast.ToUint64(v.Get(serverv2.FlagPruningInterval)),
		)

		if err := opts.Validate(); err != nil {
			return opts, fmt.Errorf("invalid custom pruning options: %w", err)
		}

		return opts, nil

	default:
		return pruningtypes.PruningOptions{}, fmt.Errorf("unknown pruning strategy %s", strategy)
	}
}
