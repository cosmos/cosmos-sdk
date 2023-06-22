package stateviewer

import "github.com/spf13/cobra"

func StatsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stats [home]",
		Short: "Prints stats about the state of an application database",
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

			stats := db.Stats()
			for key, val := range stats {
				cmd.Printf("%s: %s\n", key, val)
			}

			return nil
		},
	}

	cmd.Flags().String(FlagDBBackend, "", "The application database backend (if none specified, fallback to application config)")

	return cmd
}
