package indexer

func combineIncludeExcludeFilters(include, exclude map[string]bool) func(string) bool {
	if exclude != nil {
		for k := range include {
			delete(exclude, k)
		}
		if len(exclude) == 0 {
			return nil
		}
		return func(s string) bool {
			return !exclude[s]
		}
	} else if include != nil {
		return func(s string) bool {
			return include[s]
		}
	} else {
		return nil
	}
}

func filterIntersection(target, x map[string]bool) map[string]bool {
	if target == nil {
		target = map[string]bool{}
		for k := range x {
			target[k] = true
		}
	} else {
		for k := range target {
			if !x[k] {
				delete(target, k)
			}
		}
	}
	return target
}
