package simulation

import (
	"fmt"
	"io"
	"sort"
)

type eventStats map[string]uint

func newEventStats() eventStats {
	events := make(map[string]uint)
	return events
}

func (es eventStats) tally(eventDesc string) {
	es[eventDesc]++
}

// Pretty-print events as a table
func (es eventStats) Print(w io.Writer) {
	var keys []string
	for key := range es {
		keys = append(keys, key)
	}

	sort.Strings(keys)
	fmt.Fprintf(w, "Event statistics: \n")

	for _, key := range keys {
		fmt.Fprintf(w, "  % 60s => %d\n", key, es[key])
	}
}
