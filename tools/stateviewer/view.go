package stateviewer

import (
	"github.com/spf13/cobra"
)

func RawViewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "raw-view [home]",
		Short: "Dump the entire state of an application database to stdout",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			home := args[0]
			readDBOpts := []ReadDBOption{}
			if backend := cmd.Flag(FlagDBBackend).Value.String(); cmd.Flag(FlagDBBackend).Changed && backend != "" {
				readDBOpts = append(readDBOpts, ReadDBOptionWithBackend(backend))
			}

			db, err := ReadDB(home, readDBOpts...)
			if err != nil {
				return err
			}
			defer db.Close()

			cmd.Println(db.Print())
			return nil
		},
	}

	cmd.Flags().String(FlagDBBackend, "", "The application database backend (if none specified, fallback to application config)")

	return cmd
}

func ViewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "view [home] [key]",
		Short: "View a specific key in an application database",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return view(cmd, args)
		},
	}

	cmd.Flags().String(FlagDBBackend, "", "The application database backend (if none specified, fallback to application config)")
	cmd.Flags().Uint(FlagNear, 0, "Returns the value of the nearest keys to the one specified (if it doesn't exist)")

	return cmd
}

func view(cmd *cobra.Command, args []string) error {
	home, key := args[0], args[1]
	readDBOpts := []ReadDBOption{}
	if backend := cmd.Flag(FlagDBBackend).Value.String(); cmd.Flag(FlagDBBackend).Changed && backend != "" {
		readDBOpts = append(readDBOpts, ReadDBOptionWithBackend(backend))
	}

	db, err := ReadDB(home, readDBOpts...)
	if err != nil {
		return err
	}
	defer db.Close()

	result, err := db.Get([]byte(key))
	if err != nil {
		return err
	}

	if result == nil {
		cmd.Printf("key %q not found\n", key)
		return nil
	}

	cmd.Println(string(result))

	return nil
}
