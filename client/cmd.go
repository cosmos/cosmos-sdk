package client

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// ValidateCmd returns unknown command error or Help display if help flag set
func ValidateCmd(cmd *cobra.Command, args []string) error {
	var cmds []string
	var help bool

	// construct array of commands and search for help flag
	for _, arg := range args {
		if arg == "--help" || arg == "-h" {
			help = true
		} else if len(arg) > 0 && !(arg[0] == '-') {
			cmds = append(cmds, arg)
		}
	}

	if !help && len(cmds) > 0 {
		err := fmt.Sprintf("unknown command \"%s\" for \"%s\"", cmds[0], cmd.CalledAs())

		// build suggestions for unknown argument
		if suggestions := cmd.SuggestionsFor(cmds[0]); len(suggestions) > 0 {
			err += "\n\nDid you mean this?\n"
			for _, s := range suggestions {
				err += fmt.Sprintf("\t%v\n", s)
			}
		}
		return errors.New(err)
	}

	return cmd.Help()
}
