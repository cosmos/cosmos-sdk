package db

import (
	"fmt"
)

// VersionManager encapsulates the current valid versions of a DB and computes
// the next version.
type VersionManager struct {
	versions      map[uint64]struct{}
	initial, last uint64
}

var _ VersionSet = (*VersionManager)(nil)

// NewVersionManager creates a VersionManager from a slice of version ids.
func NewVersionManager(versions []uint64) *VersionManager {
	vmap := make(map[uint64]struct{})
	var init, last uint64
	for _, ver := range versions {
		vmap[ver] = struct{}{}
		if init == 0 || ver < init {
			init = ver
		}
		if ver > last {
			last = ver
		}
	}
	return &VersionManager{versions: vmap, initial: init, last: last}
}

// Exists implements VersionSet.
func (vm *VersionManager) Exists(version uint64) bool {
	_, has := vm.versions[version]
	return has
}

// Last implements VersionSet.
func (vm *VersionManager) Last() uint64 {
	return vm.last
}

func (vm *VersionManager) Initial() uint64 {
	return vm.initial
}

func (vm *VersionManager) Save(target uint64) (uint64, error) {
	next := vm.Last() + 1
	if target == 0 {
		target = next
	} else if target < next {
		return 0, fmt.Errorf(
			"target version cannot be less than next sequential version (%v < %v)", target, next)
	}
	if _, has := vm.versions[target]; has {
		return 0, fmt.Errorf("version exists: %v", target)
	}

	vm.versions[target] = struct{}{}
	vm.last = target
	if len(vm.versions) == 1 {
		vm.initial = target
	}
	return target, nil
}

func findLimit(m map[uint64]struct{}, cmp func(uint64, uint64) bool, init uint64) uint64 {
	for x := range m {
		if cmp(x, init) {
			init = x
		}
	}
	return init
}

func (vm *VersionManager) Delete(target uint64) {
	delete(vm.versions, target)
	if target == vm.last {
		vm.last = findLimit(vm.versions, func(x, max uint64) bool { return x > max }, 0)
	}
	if target == vm.initial {
		vm.initial = findLimit(vm.versions, func(x, min uint64) bool { return x < min }, vm.last)
	}
}

type vmIterator struct {
	ch   <-chan uint64
	open bool
	buf  uint64
}

func (vi *vmIterator) Next() bool {
	vi.buf, vi.open = <-vi.ch
	return vi.open
}
func (vi *vmIterator) Value() uint64 { return vi.buf }

// Iterator implements VersionSet.
func (vm *VersionManager) Iterator() VersionIterator {
	ch := make(chan uint64)
	go func() {
		for ver := range vm.versions {
			ch <- ver
		}
		close(ch)
	}()
	return &vmIterator{ch: ch}
}

// Count implements VersionSet.
func (vm *VersionManager) Count() int { return len(vm.versions) }

// Equal implements VersionSet.
func (vm *VersionManager) Equal(that VersionSet) bool {
	if vm.Count() != that.Count() {
		return false
	}
	for it := that.Iterator(); it.Next(); {
		if !vm.Exists(it.Value()) {
			return false
		}
	}
	return true
}

func (vm *VersionManager) Copy() *VersionManager {
	vmap := make(map[uint64]struct{})
	for ver := range vm.versions {
		vmap[ver] = struct{}{}
	}
	return &VersionManager{versions: vmap, initial: vm.initial, last: vm.last}
}
