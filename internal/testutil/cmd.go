package testutil

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// ResetArgs sets arguments for the command. It is desired to replace the cmd.SetArgs
// in the case that calling multiple times in a unit test, as cmd.ResetArgs doesn't
// reset the flag value as expected.
//
// **Warning**: this is only compatible with following flag types:
//  1. the implementations of pflag.Value
//  2. the built-in implementations of pflag.SliceValue
//  3. the custom implementations of pflag.SliceValue that are split by comma ","
//
// see https://github.com/spf13/cobra/issues/2079#issuecomment-1867991505 for more detail info
func ResetArgs(cmd *cobra.Command, args []string) {
	// if flags haven't been parsed yet, it's ok to use cmd.SetArgs
	if !cmd.Flags().Parsed() {
		cmd.SetArgs(args)
		return
	}
	// if flags have been parsed yet, we should reset flags's value that don't been set
	cmd.Flags().Visit(func(pf *pflag.Flag) {
		// if the flag hasn't been changed, ignore it
		if !pf.Changed {
			return
		}
		// handle pflag.SliceValue
		if v, ok := pf.Value.(pflag.SliceValue); ok {
			defVal := strings.Trim(pf.DefValue, "[]")
			defSliceVal := make([]string, 0)
			if defVal != "" {
				defSliceVal = strings.Split(defVal, ",")
			}
			if err := v.Replace(defSliceVal); err != nil {
				panic(fmt.Errorf("reset argument<%s> with default value<%+v> error %v", pf.Name, defSliceVal, err))
			}
			return
		}
		// handle pflag.Value
		if err := pf.Value.Set(pf.DefValue); err != nil {
			panic(fmt.Errorf("reset argument<%s> with default value<%s> error %v", pf.Name, pf.DefValue, err))
		}
	})
	// call cmd.SetArgs at last
	cmd.SetArgs(args)
}
