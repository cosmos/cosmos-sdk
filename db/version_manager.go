package db

import (
	"fmt"
	"sort"
)

// VersionManager encapsulates the current valid versions of a DB and computes
// the next version.
type VersionManager struct {
	versions []uint64
}

var _ VersionSet = (*VersionManager)(nil)

// NewVersionManager creates a VersionManager from a slice of version ids.
func NewVersionManager(versions []uint64) *VersionManager {
	vs := make([]uint64, len(versions))
	copy(vs, versions)
	sort.Slice(vs, func(i, j int) bool { return vs[i] < vs[j] })
	return &VersionManager{vs}
}

// Exists implements VersionSet.
func (vm *VersionManager) Exists(version uint64) bool {
	_, has := binarySearch(vm.versions, version)
	return has
}

// Last implements VersionSet.
func (vm *VersionManager) Last() uint64 {
	if len(vm.versions) == 0 {
		return 0
	}
	return vm.versions[len(vm.versions)-1]
}

func (vm *VersionManager) Initial() uint64 {
	if len(vm.versions) == 0 {
		return 0
	}
	return vm.versions[0]
}

func (vm *VersionManager) Save(target uint64) (uint64, error) {
	next := vm.Last() + 1
	if target == 0 {
		target = next
	} else if target < next {
		return 0, fmt.Errorf(
			"target version cannot be less than next sequential version (%v < %v)", target, next)
	}
	if vm.Exists(target) {
		return 0, fmt.Errorf("version exists: %v", target)
	}
	vm.versions = append(vm.versions, target)
	return target, nil
}

func (vm *VersionManager) Delete(target uint64) {
	i, has := binarySearch(vm.versions, target)
	if !has {
		return
	}
	vm.versions = append(vm.versions[:i], vm.versions[i+1:]...)
}

func (vm *VersionManager) DeleteAbove(target uint64) {
	var iFrom *int
	for i, v := range vm.versions {
		if iFrom == nil && v > target {
			iFrom = new(int)
			*iFrom = i
		}
	}
	if iFrom != nil {
		vm.versions = vm.versions[:*iFrom]
	}
}

type vmIterator struct {
	vmgr *VersionManager
	i    int
}

func (vi *vmIterator) Next() bool {
	vi.i++
	return vi.i < len(vi.vmgr.versions)
}
func (vi *vmIterator) Value() uint64 { return vi.vmgr.versions[vi.i] }

// Iterator implements VersionSet.
func (vm *VersionManager) Iterator() VersionIterator {
	return &vmIterator{vm, -1}
}

// Count implements VersionSet.
func (vm *VersionManager) Count() int { return len(vm.versions) }

// Equal implements VersionSet.
func (vm *VersionManager) Equal(that VersionSet) bool {
	if vm.Count() != that.Count() {
		return false
	}
	for i, it := 0, that.Iterator(); it.Next(); {
		if vm.versions[i] != it.Value() {
			return false
		}
		i++
	}
	return true
}

func (vm *VersionManager) Copy() *VersionManager {
	vs := make([]uint64, len(vm.versions))
	copy(vs, vm.versions)
	return &VersionManager{vs}
}

// Returns closest index and whether it's a match
func binarySearch(hay []uint64, ndl uint64) (int, bool) {
	var mid int
	from, to := 0, len(hay)-1
	for from <= to {
		mid = (from + to) / 2
		switch {
		case hay[mid] < ndl:
			from = mid + 1
		case hay[mid] > ndl:
			to = mid - 1
		default:
			return mid, true
		}
	}
	return from, false
}
