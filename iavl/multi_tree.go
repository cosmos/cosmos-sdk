package iavlx

type MultiTree struct {
	trees       []*Tree        // always ordered by tree name
	treesByName map[string]int // index of the trees by name
}

func (t *MultiTree) TreeByName(name string) *Tree {
	idx, ok := t.treesByName[name]
	if !ok {
		return nil
	}
	return t.trees[idx]
}
