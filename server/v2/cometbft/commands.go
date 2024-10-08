package cometbft

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	cmtcfg "github.com/cometbft/cometbft/config"
	cmtjson "github.com/cometbft/cometbft/libs/json"
	"github.com/cometbft/cometbft/node"
	"github.com/cometbft/cometbft/p2p"
	pvm "github.com/cometbft/cometbft/privval"
	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	cmtversion "github.com/cometbft/cometbft/version"
	gogoproto "github.com/cosmos/gogoproto/proto"
	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"

	"cosmossdk.io/server/v2/cometbft/client/rpc"

	"github.com/cosmos/cosmos-sdk/client"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/version"
)

func rpcClient(cmd *cobra.Command) (rpc.CometRPC, error) {
	rpcURI, err := cmd.Flags().GetString(FlagNode)
	if err != nil {
		return nil, err
	}
	if rpcURI == "" {
		return nil, errors.New("rpc URI is empty")
	}

	return rpchttp.New(rpcURI)
}

// StatusCommand returns the command to return the status of the network.
func StatusCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Query remote node for status",
		RunE: func(cmd *cobra.Command, _ []string) error {
			rpcclient, err := rpcClient(cmd)
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

	cmd.Flags().StringP(FlagNode, "n", "tcp://localhost:26657", "Node to connect to")
	cmd.Flags().StringP(FlagOutput, "o", "json", "Output format (text|json)")

	return cmd
}

// ShowNodeIDCmd - ported from CometBFT, dump node ID to stdout
func ShowNodeIDCmd() *cobra.Command {
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
func ShowValidatorCmd() *cobra.Command {
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

			clientCtx := client.GetClientContextFromCmd(cmd)
			bz, err := clientCtx.Codec.MarshalInterfaceJSON(sdkPK)
			if err != nil {
				return err
			}

			cmd.Println(string(bz))
			return nil
		},
	}

	return &cmd
}

// ShowAddressCmd - show this node's validator address
func ShowAddressCmd() *cobra.Command {
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
func VersionCmd() *cobra.Command {
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
func QueryBlocksCmd() *cobra.Command {
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
			"$ %s query blocks --query \"message.sender='cosmos1...' AND block.height > 7\" --page 1 --limit 30 --order_by asc",
			version.AppName,
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			rpcclient, err := rpcClient(cmd)
			if err != nil {
				return err
			}

			query, _ := cmd.Flags().GetString(FlagQuery)
			page, _ := cmd.Flags().GetInt(FlagPage)
			limit, _ := cmd.Flags().GetInt(FlagLimit)
			orderBy, _ := cmd.Flags().GetString(FlagOrderBy)

			blocks, err := rpc.QueryBlocks(cmd.Context(), rpcclient, page, limit, query, orderBy)
			if err != nil {
				return err
			}

			bz, err := gogoproto.Marshal(blocks)
			if err != nil {
				return err
			}

			return printOutput(cmd, bz)
		},
	}

	AddQueryFlagsToCmd(cmd)
	cmd.Flags().Int(FlagPage, query.DefaultPage, "Query a specific page of paginated results")
	cmd.Flags().Int(FlagLimit, query.DefaultLimit, "Query number of transactions results per page returned")
	cmd.Flags().String(FlagQuery, "", "The blocks events query per CometBFT's query semantics")
	cmd.Flags().String(FlagOrderBy, "", "The ordering semantics (asc|dsc)")
	_ = cmd.MarkFlagRequired(FlagQuery)

	return cmd
}

// QueryBlockCmd implements the default command for a Block query.
func QueryBlockCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "block --type={height|hash} [height|hash]",
		Short: "Query for a committed block by height, hash, or event(s)",
		Long:  "Query for a specific committed block using the CometBFT RPC `block` and `block_by_hash` method",
		Example: strings.TrimSpace(fmt.Sprintf(`
$ %s query block --%s=%s <height>
$ %s query block --%s=%s <hash>
`,
			version.AppName, FlagType, TypeHeight,
			version.AppName, FlagType, TypeHash)),
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rpcclient, err := rpcClient(cmd)
			if err != nil {
				return err
			}

			typ, _ := cmd.Flags().GetString(FlagType)
			if len(args) == 0 {
				// do not break default v0.50 behavior of block hash
				// if no args are provided, set the type to height
				typ = TypeHeight
			}

			switch typ {
			case TypeHeight:
				var (
					err    error
					height int64
				)
				heightStr := ""
				if len(args) > 0 {
					heightStr = args[0]
				}

				if heightStr == "" {
					cmd.Println("Falling back to latest block height:")
					height, err = rpc.GetChainHeight(cmd.Context(), rpcclient)
					if err != nil {
						return fmt.Errorf("failed to get chain height: %w", err)
					}
				} else {
					height, err = strconv.ParseInt(heightStr, 10, 64)
					if err != nil {
						return fmt.Errorf("failed to parse block height: %w", err)
					}
				}

				output, err := rpc.GetBlockByHeight(cmd.Context(), rpcclient, &height)
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
			case TypeHash:

				if args[0] == "" {
					return errors.New("argument should be a tx hash")
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
				return fmt.Errorf("unknown --%s value %s", FlagType, typ)
			}
		},
	}

	AddQueryFlagsToCmd(cmd)
	cmd.Flags().String(FlagType, TypeHash, fmt.Sprintf("The type to be used when querying tx, can be one of \"%s\", \"%s\"", TypeHeight, TypeHash))

	return cmd
}

// QueryBlockResultsCmd implements the default command for a BlockResults query.
func QueryBlockResultsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "block-results [height]",
		Short: "Query for a committed block's results by height",
		Long:  "Query for a specific committed block's results using the CometBFT RPC `block_results` method",
		Args:  cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			node, err := rpcClient(cmd)
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

	AddQueryFlagsToCmd(cmd)

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
				height, err = s.Consensus.store.GetLatestVersion()
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
	outFlag, err := cmd.Flags().GetString(FlagOutput)
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
