package simulation

import (
	"os"
	"sort"
	"strconv"
	"strings"
)

type SummaryEntry struct {
	Key   string
	Count int
}

func ToSortedEntries(items map[string]int) []SummaryEntry {
	entries := make([]SummaryEntry, 0, len(items))
	for k, c := range items {
		entries = append(entries, SummaryEntry{Key: k, Count: c})
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Count == entries[j].Count {
			return entries[i].Key < entries[j].Key
		}
		return entries[i].Count > entries[j].Count
	})
	return entries
}

func SummaryTopNFromEnv() int {
	v := strings.TrimSpace(os.Getenv("SIMAPP_SUMMARY_TOP_N"))
	if v == "" {
		return 0
	}
	n, err := strconv.Atoi(v)
	if err != nil || n < 0 {
		return 0
	}
	return n
}
