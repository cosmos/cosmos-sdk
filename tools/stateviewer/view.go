package stateviewer

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"

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

			db, _, err := ReadDB(home, readDBOpts...)
			if err != nil {
				return err
			}
			defer db.Close()

			return db.Print()
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
		RunE:  view,
	}

	cmd.Flags().String(FlagDBBackend, "", "The application database backend (if none specified, fallback to application config)")
	cmd.Flags().Uint(FlagNearest, 0, "Returns the value of the nearest keys to the one specified (if it doesn't exist)")

	return cmd
}

type KV struct {
	Key   string `json:"key"`
	Value []byte `json:"value"`
	Pos   int    `json:"pos,omitempty"`
}

func view(cmd *cobra.Command, args []string) error {
	readDBOpts := []ReadDBOption{}
	if backend := cmd.Flag(FlagDBBackend).Value.String(); cmd.Flag(FlagDBBackend).Changed && backend != "" {
		readDBOpts = append(readDBOpts, ReadDBOptionWithBackend(backend))
	}

	db, keyFormat, err := ReadDB(args[0], readDBOpts...)
	if err != nil {
		return err
	}
	defer db.Close()

	inputKey := args[1]
	key, err := keyFormat(inputKey)
	if err != nil {
		return err
	}

	var result []KV
	val, err := db.Get(key)
	if err != nil {
		return err
	}
	result = append(result, KV{Key: string(key), Value: val, Pos: 0})

	if bound := cmd.Flag(FlagNearest).Value.String(); cmd.Flag(FlagNearest).Changed && bound != "" {
		bound, err := strconv.Atoi(bound)
		if err != nil {
			return fmt.Errorf("invalid nearest value: %w", err)
		}

		nearestItems, err := getNearItems(db, key, bound)
		if err != nil {
			return err
		}

		result = append(result, nearestItems...)
	}

	if len(result) == 0 {
		cmd.Printf("key %q not found\n", key)
		return nil
	}

	cmd.Printf("found %d items\n", len(result))

	// sort the result by position
	sort.Slice(result, func(i, j int) bool {
		return result[i].Pos < result[j].Pos
	})

	out, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}

	cmd.Println(string(out))
	return nil
}

// getNearItems gets the nearest item to the one specified
func getNearItems(db ReadOnlyDB, key []byte, bound int) ([]KV, error) {
	result := []KV{}
	itr, err := db.Iterator(key, nil)
	if err != nil {
		return nil, err
	}
	defer itr.Close()

	revItr, err := db.ReverseIterator(key, nil)
	if err != nil {
		return nil, err
	}
	defer revItr.Close()

	for i := 0; i < bound; i++ {
		if i == 0 { // skip the first item since we look only for the nearest ones
			itr.Next()
			revItr.Next()
		}

		if itr.Valid() {
			result = append(result, KV{Key: string(itr.Key()), Value: itr.Value(), Pos: i + 1})
			itr.Next()
		}

		if revItr.Valid() {
			result = append(result, KV{Key: string(revItr.Key()), Value: revItr.Value(), Pos: -i - 1})
			revItr.Next()
		}

		if !itr.Valid() && !revItr.Valid() {
			break
		}
	}

	return result, nil
}
