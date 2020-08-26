package anycompress

import (
	"errors"
	"fmt"
        "sync"
)

// This trie is explicitly for use with prefix matching of typeURLs in a
// types.Any registry, for use with types.Any compression.

var errNoMatch = errors.New("no match found")

// trie is a prefix matcher with O(k) aka constant time searches and insertions, where k
// is the length of the subject string.
type trie struct {
        mu sync.RWMutex
	leaves [('Z' - 'A') + 1 + ('z' - 'a') + 1 + ('9' - '0') + 1 + len(".") + len("/") + len("_") + len("-")]*trie
	value  []byte
}

func newTrie() *trie {
	return new(trie)
}

func trieIndex(r byte) int {
	switch {
	case r >= 'A' && r <= 'Z':
		return int(r - 'A')

	case r >= 'a' && r <= 'z':
		return int((r - 'a') + ('Z' - 'A'))

	case r >= '0' && r <= '9':
		return int(('z' - 'a') + ('Z' - 'A') + (r - '0'))

	case r == '.':
		return int(('z' - 'a') + ('Z' - 'A') + ('9' - '0') + 1)

	case r == '/':
		return int(('z' - 'a') + ('Z' - 'A') + ('9' - '0') + 2)

	case r == '_':
		return int(('z' - 'a') + ('Z' - 'A') + ('9' - '0') + 3)

	case r == '-':
		return int(('z' - 'a') + ('Z' - 'A') + ('9' - '0') + 4)

	default:
		return -1
	}
}

func (t *trie) set(key, value []byte) (prev []byte, err error) {
        t.mu.Lock()
        defer t.mu.Unlock()

	ni := t
	for i := range key {
		r := key[i]
		ti := trieIndex(r)
		if ti < 0 {
			err = fmt.Errorf("non-trie rune: %c", r)
			return
		}

		if ni.leaves[ti] == nil {
			ni.leaves[ti] = newTrie()
		}
		ni = ni.leaves[ti]
	}

	if ni == nil {
		err = errors.New("could not insert the key")
		return
	}

	prev = ni.value
	ni.value = value

	return
}

func (t *trie) longestPrefix(key []byte) (ni *trie, i int) {
        t.mu.RLock()
        defer t.mu.RUnlock()

	ni = t
	for i = 0; i < len(key) && ni != nil; i++ {
		r := key[i]
		ti := trieIndex(r)
		if ti < 0 {
			return
		}

		if ni.leaves[ti] == nil {
			break
		}
		ni = ni.leaves[ti]
	}
	return
}

func (t *trie) get(key []byte) ([]byte, error) {
        t.mu.RLock()
        defer t.mu.RUnlock()

	ni, i := t.longestPrefix(key)
	if i != len(key) {
		return nil, errNoMatch
	}
	return ni.value, nil
}
