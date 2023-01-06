package confix

import (
	"fmt"
	"io"
	"sort"

	"github.com/creachadair/tomledit"
	"github.com/creachadair/tomledit/parser"
	"github.com/creachadair/tomledit/transform"
	"golang.org/x/exp/maps"
)

const (
	Section = "S"
	Mapping = "M"
)

type Diff struct {
	Type    string // "section" or "mapping"
	Deleted bool

	Key   string
	Value string
}

// DiffDocs diffs the keyspaces of the TOML documents in files lhs and rhs.
// Comments, order, and values are ignored for comparison purposes.
func DiffDocs(lhs, rhs *tomledit.Document) []Diff {
	diff := diffSections(lhs.Global, rhs.Global)

	lsec, rsec := lhs.Sections, rhs.Sections
	transform.SortSectionsByName(lsec)
	transform.SortSectionsByName(rsec)

	i, j := 0, 0
	for i < len(lsec) && j < len(rsec) {
		if lsec[i].Name.Before(rsec[j].Name) {
			diff = append(diff, Diff{Type: Section, Deleted: true, Key: lsec[i].Name.String()})
			for key, value := range allKVs(lsec[i]) {
				diff = append(diff, Diff{Type: Mapping, Deleted: true, Key: key, Value: value})
			}
			i++
		} else if rsec[j].Name.Before(lsec[i].Name) {
			diff = append(diff, Diff{Type: Section, Key: rsec[j].Name.String()})
			for key, value := range allKVs(rsec[j]) {
				diff = append(diff, Diff{Type: Mapping, Key: key, Value: value})
			}
			j++
		} else {
			diff = append(diff, diffSections(lsec[i], rsec[j])...)
			i++
			j++
		}
	}
	for ; i < len(lsec); i++ {
		diff = append(diff, Diff{Type: Section, Deleted: true, Key: lsec[i].Name.String()})
		for key, value := range allKVs(lsec[i]) {
			diff = append(diff, Diff{Type: Mapping, Deleted: true, Key: key, Value: value})
		}
	}
	for ; j < len(rsec); j++ {
		diff = append(diff, Diff{Type: Section, Key: rsec[j].Name.String()})
		for key, value := range allKVs(rsec[j]) {
			diff = append(diff, Diff{Type: Mapping, Key: key, Value: value})
		}
	}

	return diff
}

func allKVs(s *tomledit.Section) map[string]string {
	keys := map[string]string{}
	s.Scan(func(key parser.Key, entry *tomledit.Entry) bool {
		keys[key.String()] = entry.Value.String()
		return true
	})
	return keys
}

func diffSections(lhs, rhs *tomledit.Section) []Diff {
	return diffKeys(allKVs(lhs), allKVs(rhs))
}

func diffKeys(lhs, rhs map[string]string) []Diff {
	diff := []Diff{}

	lhsKeys := maps.Keys(lhs)
	rhsKeys := maps.Keys(rhs)
	sort.Strings(lhsKeys)
	sort.Strings(rhsKeys)

	i, j := 0, 0
	for i < len(lhs) && j < len(rhs) {
		if lhsKeys[i] < rhsKeys[j] {
			diff = append(diff, Diff{Type: Mapping, Deleted: true, Key: lhsKeys[i], Value: lhs[lhsKeys[i]]})
			i++
		} else if lhsKeys[i] > rhsKeys[j] {
			diff = append(diff, Diff{Type: Mapping, Key: rhsKeys[j], Value: rhs[rhsKeys[j]]})
			j++
		} else {
			i++
			j++
		}
	}
	for ; i < len(lhsKeys); i++ {
		diff = append(diff, Diff{Type: Mapping, Deleted: true, Key: lhsKeys[i], Value: lhs[lhsKeys[i]]})
	}
	for ; j < len(rhsKeys); j++ {
		diff = append(diff, Diff{Type: Mapping, Key: rhsKeys[j], Value: rhs[rhsKeys[j]]})
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
			fmt.Fprintln(w, fmt.Sprintf("-%s", diff.Type), fmt.Sprintf("%s=%s", diff.Key, diff.Value))
		} else {
			fmt.Fprintln(w, fmt.Sprintf("+%s", diff.Type), fmt.Sprintf("%s=%s", diff.Key, diff.Value))
		}
	}
}
