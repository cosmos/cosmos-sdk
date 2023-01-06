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

type Diff struct {
	Type    string // "section" or "mapping"
	Deleted bool

	Key string
	// TODO store value change as well
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
			for _, key := range allKeys(lsec[i]) {
				diff = append(diff, Diff{Type: Mapping, Deleted: true, Key: key})
			}
			i++
		} else if rsec[j].Name.Before(lsec[i].Name) {
			diff = append(diff, Diff{Type: Section, Key: rsec[j].Name.String()})
			for _, key := range allKeys(rsec[j]) {
				diff = append(diff, Diff{Type: Mapping, Key: key})
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
		for _, key := range allKeys(lsec[i]) {
			diff = append(diff, Diff{Type: Mapping, Deleted: true, Key: key})
		}
	}
	for ; j < len(rsec); j++ {
		diff = append(diff, Diff{Type: Section, Key: rsec[j].Name.String()})
		for _, key := range allKeys(rsec[j]) {
			diff = append(diff, Diff{Type: Mapping, Key: key})
		}
	}

	return diff
}

func allKeys(s *tomledit.Section) []string {
	var keys []string
	s.Scan(func(key parser.Key, _ *tomledit.Entry) bool {
		keys = append(keys, key.String())
		return true
	})
	return keys
}

func diffSections(lhs, rhs *tomledit.Section) []Diff {
	return diffKeys(allKeys(lhs), allKeys(rhs))
}

func diffKeys(lhs, rhs []string) []Diff {
	diff := []Diff{}

	sort.Strings(lhs)
	sort.Strings(rhs)

	i, j := 0, 0
	for i < len(lhs) && j < len(rhs) {
		if lhs[i] < rhs[j] {
			diff = append(diff, Diff{Type: Mapping, Deleted: true, Key: lhs[i]})
			i++
		} else if lhs[i] > rhs[j] {
			diff = append(diff, Diff{Type: Mapping, Key: rhs[j]})
			j++
		} else {
			i++
			j++
		}
	}
	for ; i < len(lhs); i++ {
		diff = append(diff, Diff{Type: Mapping, Deleted: true, Key: lhs[i]})
	}
	for ; j < len(rhs); j++ {
		diff = append(diff, Diff{Type: Mapping, Key: rhs[j]})
	}

	return diff
}

// PrintDiff output prints one line per key that differs:
// -S name    -- section exists in lhs but not rhs
// +S name    -- section exists in rhs but not lhs
// -M name    -- mapping exists in lhs but not f2
// +M name    -- mapping exists in rhs but not lhs
func PrintDiff(w io.Writer, diffs []Diff) {
	for _, diff := range diffs {
		if diff.Deleted {
			fmt.Fprintln(w, fmt.Sprintf("-%s", diff.Type), diff.Key)
		} else {
			fmt.Fprintln(w, fmt.Sprintf("+%s", diff.Type), diff.Key)
		}
	}
}
