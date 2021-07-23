package container

import "reflect"

type config struct {
	autoGroupTypes   map[reflect.Type]bool
	onePerScopeTypes map[reflect.Type]bool
}
