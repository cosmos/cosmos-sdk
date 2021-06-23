package db

import (
	"bytes"
	"os"
)

func cp(bz []byte) (ret []byte) {
	ret = make([]byte, len(bz))
	copy(ret, bz)
	return ret
}

// Returns a slice of the same length (big endian)
// except incremented by one.
// Returns nil on overflow (e.g. if bz bytes are all 0xFF)
// CONTRACT: len(bz) > 0
func cpIncr(bz []byte) (ret []byte) {
	if len(bz) == 0 {
		panic("cpIncr expects non-zero bz length")
	}
	ret = cp(bz)
	for i := len(bz) - 1; i >= 0; i-- {
		if ret[i] < byte(0xFF) {
			ret[i]++
			return
		}
		ret[i] = byte(0x00)
		if i == 0 {
			// Overflow
			return nil
		}
	}
	return nil
}

// See DB interface documentation for more information.
func IsKeyInDomain(key, start, end []byte) bool {
	if bytes.Compare(key, start) < 0 {
		return false
	}
	if end != nil && bytes.Compare(end, key) <= 0 {
		return false
	}
	return true
}

func FileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}

// Encapsulates valid and current versions
type VersionManager struct {
	Versions []uint64
	current  uint64
	// initial uint64
}

func NewVersionManager(versions []uint64) *VersionManager {
	ret := &VersionManager{Versions: versions}
	// Set current working version
	if len(versions) == 0 {
		ret.current = 1
	} else {
		ret.current = versions[len(versions)-1] + 1
	}
	return ret
}

func (vm *VersionManager) Valid(version uint64) bool {
	if version == vm.Current() {
		return true
	}
	// todo: maybe use map to avoid linear search
	for _, ver := range vm.Versions {
		if ver == version {
			return true
		}
	}
	return false
}
func (vm *VersionManager) Initial() uint64 {
	if len(vm.Versions) == 0 {
		return 1
	}
	return vm.Versions[0]
}
func (vm *VersionManager) Current() uint64 {
	return vm.current
}
func (vm *VersionManager) Save() uint64 {
	id := vm.Current()
	vm.current = id + 1
	vm.Versions = append(vm.Versions, id)
	return id
}
