package maps

import (
	"fmt"
	"maps"
	"strings"
)

type genericMapValueOptions[K comparable, V any] struct {
	keyParser   func(string) (K, error)
	valueParser func(string) (V, error)
	genericType string
}

type genericMapValue[K comparable, V any] struct {
	value   *map[K]V
	changed bool
	Options genericMapValueOptions[K, V]
}

func newGenericMapValue[K comparable, V any](val map[K]V, p *map[K]V) *genericMapValue[K, V] {
	ssv := new(genericMapValue[K, V])
	ssv.value = p
	*ssv.value = val
	return ssv
}

func (gm *genericMapValue[K, V]) Set(val string) error {
	ss := strings.Split(val, ",")
	out := make(map[K]V, len(ss))
	for _, pair := range ss {
		key, val, found := strings.Cut(pair, "=")
		if !found {
			return fmt.Errorf("%s must be formatted as key=value", pair)
		}
		parsedKey, err := gm.Options.keyParser(key)
		if err != nil {
			return err
		}
		out[parsedKey], err = gm.Options.valueParser(val)
		if err != nil {
			return err
		}
	}
	if !gm.changed {
		*gm.value = out
	} else {
		maps.Copy(*gm.value, out)
	}
	gm.changed = true
	return nil
}

func (gm *genericMapValue[K, V]) Type() string {
	return gm.Options.genericType
}

func (gm *genericMapValue[K, V]) String() string {
	return ""
}
