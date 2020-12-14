package cli

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/evidence/types"
)

// GetQueryCmd returns the CLI command with all evidence module query commands
// mounted.
func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   types.ModuleName,
		Short: "Query for evidence by hash or for all (paginated) submitted evidence",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query for specific submitted evidence by hash or query for all (paginated) evidence:

Example:
$ %s query %s DF0C23E8634E480F84B9D5674A7CDC9816466DEC28A3358F73260F68D28D7660
$ %s query %s --page=2 --limit=50
`,
				version.AppName, types.ModuleName, version.AppName, types.ModuleName,
			),
		),
		Args:                       cobra.MaximumNArgs(1),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       QueryEvidenceCmd(),
	}

	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "evidence")

	return cmd
}

// QueryEvidenceCmd returns the command handler for evidence querying. Evidence
// can be queried for by hash or paginated evidence can be returned.
func QueryEvidenceCmd() func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		clientCtx, err := client.GetClientQueryContext(cmd)
		if err != nil {
			return err
		}
		if len(args) > 0 {
			return queryEvidence(clientCtx, args[0])
		}

		pageReq, err := client.ReadPageRequest(cmd.Flags())
		if err != nil {
			return err
		}

		return queryAllEvidence(clientCtx, pageReq)
	}
}

func queryEvidence(clientCtx client.Context, hash string) error {
	decodedHash, err := hex.DecodeString(hash)
	if err != nil {
		return fmt.Errorf("invalid evidence hash: %w", err)
	}

	queryClient := types.NewQueryClient(clientCtx)

	params := &types.QueryEvidenceRequest{EvidenceHash: decodedHash}
	res, err := queryClient.Evidence(context.Background(), params)

	if err != nil {
		return err
	}

	return clientCtx.PrintProto(res.Evidence)
}

func queryAllEvidence(clientCtx client.Context, pageReq *query.PageRequest) error {
	queryClient := types.NewQueryClient(clientCtx)

	params := &types.QueryAllEvidenceRequest{
		Pagination: pageReq,
	}

	res, err := queryClient.AllEvidence(context.Background(), params)
	if err != nil {
		return err
	}

	return clientCtx.PrintProto(res)
}
