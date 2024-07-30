package indexer

import (
	"fmt"
	"sort"

	"cosmossdk.io/schema/appdata"
)

type FilterConfig struct {
	// ExcludeState specifies that the indexer will not receive state updates.
	ExcludeState bool `json:"exclude_state"`

	// ExcludeEvents specifies that the indexer will not receive events.
	ExcludeEvents bool `json:"exclude_events"`

	// ExcludeTxs specifies that the indexer will not receive transaction's.
	ExcludeTxs bool `json:"exclude_txs"`

	// ExcludeBlockHeaders specifies that the indexer will not receive block headers,
	// although it will still receive StartBlock and Commit callbacks, just without
	// the header data.
	ExcludeBlockHeaders bool `json:"exclude_block_headers"`

	Modules ModuleFilterConfig `json:"modules"`
}

func (f FilterConfig) Validate() error {
	return f.Modules.Validate()
}

func (f FilterConfig) Apply(listener appdata.Listener) appdata.Listener {
	listener = f.Modules.Apply(listener)

	if f.ExcludeBlockHeaders && listener.StartBlock != nil {
		cb := listener.StartBlock
		listener.StartBlock = func(data appdata.StartBlockData) error {
			data.HeaderBytes = nil
			data.HeaderJSON = nil
			return cb(data)
		}
	}

	if f.ExcludeTxs {
		listener.OnTx = nil
	}

	if f.ExcludeEvents {
		listener.OnEvent = nil
	}

	return listener
}

type ModuleFilterConfig struct {
	// Include specifies a list of modules whose state the indexer will
	// receive state updates for.
	// Only one of include or exclude modules should be specified.
	Include []string `json:"include"`

	// Exclude specifies a list of modules whose state the indexer will not
	// receive state updates for.
	// Only one of include or exclude modules should be specified.
	Exclude []string `json:"exclude"`
}

func (f ModuleFilterConfig) Validate() error {
	if len(f.Exclude) != 0 && len(f.Include) != 0 {
		return fmt.Errorf("only one of exclude or include can be set for module filter")
	}
	return nil
}

// Apply applies the module filter config to the listener. It will panic if the config is invalid.
// Callers should check the config with Validate before calling Apply.
func (f ModuleFilterConfig) Apply(listener appdata.Listener) appdata.Listener {
	if f := f.ToFunction(); f != nil {
		listener = appdata.ModuleFilter(listener, f)
	}

	return listener
}

func (f ModuleFilterConfig) ToFunction() func(string) bool {
	if err := f.Validate(); err != nil {
		panic(err)
	}

	if len(f.Exclude) != 0 {
		excluded := filterListToMap(f.Exclude)
		return func(moduleName string) bool {
			return !excluded[moduleName]
		}

	} else if len(f.Include) != 0 {
		included := filterListToMap(f.Include)
		return func(moduleName string) bool {
			return included[moduleName]
		}
	} else {
		return nil
	}
}

func combineModuleFilters(filters []ModuleFilterConfig) ModuleFilterConfig {
	var exclusionFilter map[string]bool
	allIncludedModules := map[string]bool{}
	for _, filter := range filters {
		if len(filter.Include) != 0 {
			for _, moduleName := range filter.Include {
				// we create a UNION of all included modules
				allIncludedModules[moduleName] = true
			}
		} else if len(filter.Exclude) != 0 {
			if exclusionFilter == nil {
				// if this is the first exclude filter then just use it
				exclusionFilter = filterListToMap(filter.Exclude)
			} else {
				// if this is a new exclude filter we must INTERSECT it with the existing filter
				for _, name := range filter.Exclude {
					if !exclusionFilter[name] {
						delete(exclusionFilter, name)
					}
				}
			}
		}
	}

	if len(exclusionFilter) != 0 {
		// if we're excluding modules then this takes priority over include filters, but we
		// must clear included keys from the exclusion list
		for k := range allIncludedModules {
			delete(exclusionFilter, k)
		}
		if len(exclusionFilter) == 0 {
			// if we're not excluding anything now, we return an empty filter
			return ModuleFilterConfig{}
		}
		return ModuleFilterConfig{Exclude: filterMapToList(exclusionFilter)}
	} else if len(allIncludedModules) != 0 {
		// if we're just including modules, we return that list
		return ModuleFilterConfig{Include: filterMapToList(allIncludedModules)}
	} else {
		// we have nothing to filter so empty
		return ModuleFilterConfig{}
	}
}

func filterListToMap(names []string) map[string]bool {
	res := map[string]bool{}
	for _, name := range names {
		res[name] = true
	}
	return res
}

func filterMapToList(names map[string]bool) []string {
	res := make([]string, 0, len(names))
	for name := range names {
		res = append(res, name)
	}
	sort.Strings(res)
	return res
}
