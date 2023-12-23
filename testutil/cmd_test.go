package testutil_test

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
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

	// although we called cmd.SetArgs and expected that only b was set, but a has already set by last call `cmd.SetArgs`
	cmd.SetArgs([]string{
		"testcmd",
		"--b=true",
	})
	require.True(t, cmd.Flags().Changed("a"))
	require.Error(t, cmd.Execute())

	// although we called cmd.SetArgs and expected that only c was set, but a,b has already set by last two call `cmd.SetArgs`
	cmd.SetArgs([]string{
		"testcmd",
		"--c=true",
	})
	require.Error(t, cmd.Execute())

	// we should explicitly set all args that has changed before, so that it works as we expected
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
