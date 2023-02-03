package server

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tendermint/tendermint/p2p"
	pvm "github.com/tendermint/tendermint/privval"
	tversion "github.com/tendermint/tendermint/version"
	"sigs.k8s.io/yaml"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	rpc "github.com/cosmos/cosmos-sdk/client/rpc"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/version"
	auth "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
)

// ShowNodeIDCmd - ported from Tendermint, dump node ID to stdout
func ShowNodeIDCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show-node-id",
		Short: "Show this node's ID",
		RunE: func(cmd *cobra.Command, args []string) error {
			serverCtx := GetServerContextFromCmd(cmd)
			cfg := serverCtx.Config

			nodeKey, err := p2p.LoadNodeKey(cfg.NodeKeyFile())
			if err != nil {
				return err
			}

			fmt.Println(nodeKey.ID())
			return nil
		},
	}
}

// ShowValidatorCmd - ported from Tendermint, show this node's validator info
func ShowValidatorCmd() *cobra.Command {
	cmd := cobra.Command{
		Use:   "show-validator",
		Short: "Show this node's tendermint validator info",
		RunE: func(cmd *cobra.Command, args []string) error {
			serverCtx := GetServerContextFromCmd(cmd)
			cfg := serverCtx.Config

			privValidator := pvm.LoadFilePV(cfg.PrivValidatorKeyFile(), cfg.PrivValidatorStateFile())
			pk, err := privValidator.GetPubKey()
			if err != nil {
				return err
			}

			sdkPK, err := cryptocodec.FromTmPubKeyInterface(pk)
			if err != nil {
				return err
			}

			clientCtx := client.GetClientContextFromCmd(cmd)
			bz, err := clientCtx.Codec.MarshalInterfaceJSON(sdkPK)
			if err != nil {
				return err
			}

			fmt.Println(string(bz))
			return nil
		},
	}

	return &cmd
}

// ShowAddressCmd - show this node's validator address
func ShowAddressCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show-address",
		Short: "Shows this node's tendermint validator consensus address",
		RunE: func(cmd *cobra.Command, args []string) error {
			serverCtx := GetServerContextFromCmd(cmd)
			cfg := serverCtx.Config

			privValidator := pvm.LoadFilePV(cfg.PrivValidatorKeyFile(), cfg.PrivValidatorStateFile())

			valConsAddr := (sdk.ConsAddress)(privValidator.GetAddress())
			fmt.Println(valConsAddr.String())
			return nil
		},
	}

	return cmd
}

// VersionCmd prints tendermint and ABCI version numbers.
func VersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print tendermint libraries' version",
		Long:  "Print protocols' and libraries' version numbers against which this app has been compiled.",
		RunE: func(cmd *cobra.Command, args []string) error {
			bs, err := yaml.Marshal(&struct {
				Tendermint    string
				ABCI          string
				BlockProtocol uint64
				P2PProtocol   uint64
			}{
				Tendermint:    tversion.TMCoreSemVer,
				ABCI:          tversion.ABCIVersion,
				BlockProtocol: tversion.BlockProtocol,
				P2PProtocol:   tversion.P2PProtocol,
			})
			if err != nil {
				return err
			}

			fmt.Println(string(bs))
			return nil
		},
	}
}

// QueryBlocksByEventsCmd returns a command to search through blocks by events.
func QueryBlocksByEventsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "blocks",
		Short: "Query for paginated blocks that match a set of events",
		Long: `Search for blocks that match the exact given events where results are paginated.
The events query is directly passed to Tendermint's RPC BlockSearch method and must
conform to Tendermint's query syntax.
Please refer to each module's documentation for the full set of events to query
for. Each module documents its respective events under 'xx_events.md'.
`,
		Example: fmt.Sprintf(
			"$ %s query blocks --query \"message.sender='cosmos1...' AND block.height > 7\" --page 1 --limit 30 --order-by ASC",
			version.AppName,
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			query, _ := cmd.Flags().GetString(auth.FlagQuery)
			page, _ := cmd.Flags().GetInt(flags.FlagPage)
			limit, _ := cmd.Flags().GetInt(flags.FlagLimit)
			orderBy, _ := cmd.Flags().GetString(auth.FlagOrderBy)

			blocks, err := rpc.QueryBlocksByEvents(clientCtx, page, limit, query, orderBy)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(blocks)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	cmd.Flags().Int(flags.FlagPage, query.DefaultPage, "Query a specific page of paginated results")
	cmd.Flags().Int(flags.FlagLimit, query.DefaultLimit, "Query number of transactions results per page returned")
	cmd.Flags().String(auth.FlagQuery, "", "The blocks events query per Tendermint's query semantics")
	cmd.Flags().String(auth.FlagOrderBy, "", "The ordering semantics (asc|dsc)")
	_ = cmd.MarkFlagRequired(auth.FlagQuery)

	return cmd
}

// QueryBlockCmd implements the default command for a Block query.
func QueryBlockCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "block --type=[height|hash] [height|hash]",
		Short: "Query for a committed block by height, hash, or event(s)",
		Long: strings.TrimSpace(fmt.Sprintf(`
Example:
$ %s query block --%s=%s <height>
$ %s query block --%s=%s <hash>
`,
			version.AppName, auth.FlagType, auth.TypeHeight,
			version.AppName, auth.FlagType, auth.TypeHash)),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			typ, _ := cmd.Flags().GetString(auth.FlagType)

			switch typ {
			case auth.TypeHeight:
				{
					if args[0] == "" {
						return fmt.Errorf("argument should be a block height")
					}

					var height *int64

					// optional height
					if len(args) > 0 {
						h, err := strconv.Atoi(args[0])
						if err != nil {
							return err
						}
						if h > 0 {
							tmp := int64(h)
							height = &tmp
						}
					}

					output, err := rpc.GetBlockByHeight(clientCtx, height)
					if err != nil {
						return err
					}

					if output.Header.Height == 0 {
						return fmt.Errorf("no block found with height %s", args[0])
					}

					return clientCtx.PrintProto(output)
				}
			case auth.TypeHash:
				{
					if args[0] == "" {
						return fmt.Errorf("argument should be a tx hash")
					}

					// If hash is given, then query the tx by hash.
					output, err := rpc.GetBlockByHash(clientCtx, args[0])
					if err != nil {
						return err
					}

					if output.Header.AppHash == nil {
						return fmt.Errorf("no block found with hash %s", args[0])
					}

					return clientCtx.PrintProto(output)
				}
			default:
				return fmt.Errorf("unknown --%s value %s", auth.FlagType, typ)
			}
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	cmd.Flags().String(auth.FlagType, auth.TypeHash, fmt.Sprintf("The type to be used when querying tx, can be one of \"%s\", \"%s\"", auth.TypeHeight, auth.TypeHash))

	return cmd
}
