package container

import (
	"fmt"
	"reflect"
)

type config struct {
	autoGroupTypes   map[reflect.Type]bool
	onePerScopeTypes map[reflect.Type]bool
	loggers          []func(string)
	indentStr        string
}

func newConfig() *config {
	return &config{
		autoGroupTypes:   map[reflect.Type]bool{},
		onePerScopeTypes: map[reflect.Type]bool{},
	}
}

func (c *config) indentLogger() {
	c.indentStr = c.indentStr + " "
}

func (c *config) dedentLogger() {
	if len(c.indentStr) > 0 {
		c.indentStr = c.indentStr[1:]
	}
}

func (c config) logf(format string, args ...interface{}) {
	s := fmt.Sprintf(c.indentStr+format, args...)
	for _, logger := range c.loggers {
		logger(s)
	}
}
