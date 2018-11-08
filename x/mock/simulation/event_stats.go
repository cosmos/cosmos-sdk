package simulation

import (
	"fmt"
	"sort"
)

type eventStats map[string]uint

func newEventStats() eventStats {
	events := make(map[string]uint)
	return events
	event := func(what string) {
		events[what]++
	}
}

func (es *eventStats) tally(eventDesc string) {
	es[eventDesc]++
}

// Pretty-print events as a table
func (es eventStats) Print() {
	var keys []string
	for key := range es {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	fmt.Printf("Event statistics: \n")
	for _, key := range keys {
		fmt.Printf("  % 60s => %d\n", key, es[key])
	}
}
