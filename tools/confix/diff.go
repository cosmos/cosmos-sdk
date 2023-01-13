package confix

import (
	"fmt"
	"io"
	"sort"

	"github.com/creachadair/tomledit"
	"github.com/creachadair/tomledit/parser"
	"github.com/creachadair/tomledit/transform"
)

const (
	Section = "S"
	Mapping = "M"
)

type KV struct {
	Key   string
	Value string
	Block []string // comment block
}

type Diff struct {
	Type    string // "section" or "mapping"
	Deleted bool

	KV KV
}

// DiffKeys diffs the keyspaces of the TOML documents in files lhs and rhs.
// Comments, order, and values are ignored for comparison purposes.
func DiffKeys(lhs, rhs *tomledit.Document) []Diff {
	diff := diffSections(lhs.Global, rhs.Global)

	lsec, rsec := lhs.Sections, rhs.Sections
	transform.SortSectionsByName(lsec)
	transform.SortSectionsByName(rsec)

	i, j := 0, 0
	for i < len(lsec) && j < len(rsec) {
		if lsec[i].Name.Before(rsec[j].Name) {
			diff = append(diff, Diff{Type: Section, Deleted: true, KV: KV{Key: lsec[i].Name.String()}})
			for _, kv := range allKVs(lsec[i]) {
				diff = append(diff, Diff{Type: Mapping, Deleted: true, KV: kv})
			}
			i++
		} else if rsec[j].Name.Before(lsec[i].Name) {
			diff = append(diff, Diff{Type: Section, KV: KV{Key: rsec[j].Name.String()}})
			for _, kv := range allKVs(rsec[j]) {
				diff = append(diff, Diff{Type: Mapping, KV: kv})
			}
			j++
		} else {
			diff = append(diff, diffSections(lsec[i], rsec[j])...)
			i++
			j++
		}
	}
	for ; i < len(lsec); i++ {
		diff = append(diff, Diff{Type: Section, Deleted: true, KV: KV{Key: lsec[i].Name.String()}})
		for _, kv := range allKVs(lsec[i]) {
			diff = append(diff, Diff{Type: Mapping, Deleted: true, KV: kv})
		}
	}
	for ; j < len(rsec); j++ {
		diff = append(diff, Diff{Type: Section, KV: KV{Key: rsec[j].Name.String()}})
		for _, kv := range allKVs(rsec[j]) {
			diff = append(diff, Diff{Type: Mapping, KV: kv})
		}
	}

	return diff
}

func allKVs(s *tomledit.Section) []KV {
	keys := []KV{}
	s.Scan(func(key parser.Key, entry *tomledit.Entry) bool {
		keys = append(keys, KV{
			Key:   key.String(),
			Value: entry.Value.String(),
			Block: entry.Block,
		})

		return true
	})
	return keys
}

func diffSections(lhs, rhs *tomledit.Section) []Diff {
	return diffKeys(allKVs(lhs), allKVs(rhs))
}

func diffKeys(lhs, rhs []KV) []Diff {
	diff := []Diff{}

	sort.Slice(lhs, func(i, j int) bool {
		return lhs[i].Key < lhs[j].Key
	})
	sort.Slice(rhs, func(i, j int) bool {
		return rhs[i].Key < rhs[j].Key
	})

	i, j := 0, 0
	for i < len(lhs) && j < len(rhs) {
		if lhs[i].Key < rhs[j].Key {
			diff = append(diff, Diff{Type: Mapping, Deleted: true, KV: lhs[i]})
			i++
		} else if lhs[i].Key > rhs[j].Key {
			diff = append(diff, Diff{Type: Mapping, KV: rhs[j]})
			j++
		} else {
			i++
			j++
		}
	}
	for ; i < len(lhs); i++ {
		diff = append(diff, Diff{Type: Mapping, Deleted: true, KV: lhs[i]})
	}
	for ; j < len(rhs); j++ {
		diff = append(diff, Diff{Type: Mapping, KV: rhs[j]})
	}

	return diff
}

// PrintDiff output prints one line per key that differs:
// -S name    -- section exists in f1 but not f2
// +S name    -- section exists in f2 but not f1
// -M name    -- mapping exists in f1 but not f2
// +M name    -- mapping exists in f2 but not f1
func PrintDiff(w io.Writer, diffs []Diff) {
	for _, diff := range diffs {
		if diff.Deleted {
			fmt.Fprintln(w, fmt.Sprintf("-%s", diff.Type), fmt.Sprintf("%s=%s", diff.KV.Key, diff.KV.Value))
		} else {
			fmt.Fprintln(w, fmt.Sprintf("+%s", diff.Type), fmt.Sprintf("%s=%s", diff.KV.Key, diff.KV.Value))
		}
	}
}
