package testutil_test

import (
	"fmt"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/testutil"
)

// TestSetArgsWithOriginalMethod is used to illustrate cobra.Command.SetArgs won't reset args as expected
func TestSetArgsWithOriginalMethod(t *testing.T) {
	getCMD := func() *cobra.Command {
		cmd := &cobra.Command{
			Use: "testcmd",
			RunE: func(cmd *cobra.Command, args []string) error {
				a, _ := cmd.Flags().GetBool("a")
				b, _ := cmd.Flags().GetBool("b")
				c, _ := cmd.Flags().GetBool("c")
				switch {
				case a && b, a && c, b && c:
					return fmt.Errorf("a,b,c only one could be true")
				}
				return nil
			},
		}
		f := cmd.Flags()
		f.BoolP("a", "a", false, "a,b,c only one could be true")
		f.BoolP("b", "b", false, "a,b,c only one could be true")
		f.Bool("c", false, "a,b,c only one could be true")
		return cmd
	}

	cmd := getCMD()

	cmd.SetArgs([]string{
		"testcmd",
		"--a=true",
	})
	require.NoError(t, cmd.Execute())

	// This call to cmd.SetArgs is expected to set only the 'b' flag. However, due to the bug, the 'a' flag remains set from the previous call to cmd.SetArgs, leading to an error.
	cmd.SetArgs([]string{
		"testcmd",
		"--b=true",
	})
	require.True(t, cmd.Flags().Changed("a"))
	require.Error(t, cmd.Execute())

	// This call to cmd.SetArgs is expected to set only the 'c' flag. However, the 'a' and 'b' flags remain set from the previous calls, causing an unexpected error.
	cmd.SetArgs([]string{
		"testcmd",
		"--c=true",
	})
	require.Error(t, cmd.Execute())

	// To work around the bug, we must explicitly reset the 'a' and 'b' flags to false, even though we only want to set the 'c' flag to true.
	cmd.SetArgs([]string{
		"testcmd",
		"--a=false",
		"--b=false",
		"--c=true",
	})
	require.NoError(t, cmd.Execute())
}

func TestSetArgsWithWrappedMethod(t *testing.T) {
	getCMD := func() *cobra.Command {
		cmd := &cobra.Command{
			Use: "testcmd",
			RunE: func(cmd *cobra.Command, args []string) error {
				a, _ := cmd.Flags().GetBool("a")
				b, _ := cmd.Flags().GetBool("b")
				c, _ := cmd.Flags().GetBool("c")
				switch {
				case a && b, a && c, b && c:
					return fmt.Errorf("a,b,c only one could be true")
				}
				return nil
			},
		}
		f := cmd.Flags()
		f.BoolP("a", "a", false, "a,b,c only one could be true")
		f.BoolP("b", "b", false, "a,b,c only one could be true")
		f.Bool("c", false, "a,b,c only one could be true")
		return cmd
	}

	cmd := getCMD()

	testutil.SetArgs(cmd, []string{
		"testcmd",
		"--a=true",
	})
	require.NoError(t, cmd.Execute())

	testutil.SetArgs(cmd, []string{
		"testcmd",
		"--b=true",
	})
	require.True(t, cmd.Flags().Changed("a"))
	require.NoError(t, cmd.Execute())

	testutil.SetArgs(cmd, []string{
		"testcmd",
		"--c=true",
	})
	require.NoError(t, cmd.Execute())

	testutil.SetArgs(cmd, []string{
		"testcmd",
		"--a=false",
		"--b=false",
		"--c=true",
	})
	require.NoError(t, cmd.Execute())
}
