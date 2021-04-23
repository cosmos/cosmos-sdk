package exported

import (
	"sort"
	"unicode"

	"github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

// ProtocolVersionSetter defines the interface fulfilled by BaseApp
// which allows setting it's appVersion field.
type ProtocolVersionSetter interface {
	SetProtocolVersion(uint64)
}

// Sorting methods to sort slices of ModuleConsensusVersion
// by module name in alphabetical order
type ModuleVersionSlice []*types.ModuleConsensusVersion

func Sort(s []*types.ModuleConsensusVersion) []*types.ModuleConsensusVersion {
	var t ModuleVersionSlice = s
	sort.Sort(t)
	s = t
	return s
}

func (m ModuleVersionSlice) Len() int {
	return len(m)
}

func (m ModuleVersionSlice) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}

func (m ModuleVersionSlice) Less(i, j int) bool {
	iRunes := []rune(m[i].Module)
	jRunes := []rune(m[j].Module)

	max := len(iRunes)
	if max > len(jRunes) {
		max = len(jRunes)
	}

	for idx := 0; idx < max; idx++ {
		ir := iRunes[idx]
		jr := jRunes[idx]

		lir := unicode.ToLower(ir)
		ljr := unicode.ToLower(jr)

		if lir != ljr {
			return lir < ljr
		}

		// the lowercase runes are the same, so compare the original
		if ir != jr {
			return ir < jr
		}
	}

	// If the strings are the same up to the length of the shortest string,
	// the shorter string comes first
	return len(iRunes) < len(jRunes)
}
