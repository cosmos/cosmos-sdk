package cli

import (
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/spf13/cobra"
	"github.com/cosmos/cosmos-sdk/client/context"
	"fmt"

)

// create edit validator command
func QueryParam(storeName string, cdc *wire.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "query",
		Short: "query param value from global store",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.NewCoreContextFromViper()
			res, err := ctx.QueryStore([]byte(args[0]), storeName)
			if err != nil {
				fmt.Println(err.Error())
				return nil
			}
			var result string
			if cdc.UnmarshalBinary(res, &result) == nil {
				fmt.Println(result)
			}
			return nil
		},
	}
	return cmd
}
