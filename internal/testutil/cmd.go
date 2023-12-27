package testutil

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// SetArgs sets arguments for the command. It is desired to replace the cmd.SetArgs in all test case, as cmd.SetArgs doesn't reset flag value as expected.
//
// see https://github.com/spf13/cobra/issues/2079#issuecomment-1867991505 for more detail info
func SetArgs(cmd *cobra.Command, args []string) {
	if cmd.Flags().Parsed() {
		cmd.Flags().Visit(func(pf *pflag.Flag) {
			if err := pf.Value.Set(pf.DefValue); err != nil {
				panic(fmt.Errorf("reset argument[%s] value error %v", pf.Name, err))
			}
		})
	}
	cmd.SetArgs(args)
}
