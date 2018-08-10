package cli

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/spf13/cobra"
)

func ExportCmd(storeName string, cdc *wire.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export",
		Short: "export all keypair which begin with args from global store.(args can be 'gov','stake' eg,or full path such as 'gov/votingprocedure/votingPeriod')",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.NewCoreContextFromViper()
			res, err := ctx.QuerySubspace(cdc, []byte(args[0]), storeName)
			if err != nil {
				fmt.Println(err.Error())
				return nil
			}
			var result []KVPair
			for _, pair := range res {
				var v string
				cdc.UnmarshalBinary(pair.Value, &v)
				kv := KVPair{
					K: string(pair.Key),
					V: v,
				}
				result = append(result, kv)
			}
			output, err := wire.MarshalJSONIndent(cdc, result)
			if err != nil {
				return err
			}
			fmt.Println(string(output))
			return nil
		},
	}
	return cmd
}

type KVPair struct {
	K string `json:"key"`
	V string `json:"value"`
}
