package depinject

import (
	"fmt"
	"reflect"
	"strings"
)

type varRef string

func (v varRef) emit() string {
	return string(v)
}

type funCall struct {
	loc  Location
	args []expr
}

func (f funCall) emit() string {
	var args []string
	for _, arg := range f.args {
		if arg == nil {
			args = append(args, "nil")
		} else {
			args = append(args, arg.emit())
		}
	}
	return fmt.Sprintf("%s(%s)", f.loc.Name(), strings.Join(args, ", "))
}

type zeroValue struct {
	reflect.Type
}

func (z zeroValue) emit() string {
	return reflect.Zero(z).String()
}

type castType struct {
	typ reflect.Type
	e   expr
}

func (c castType) emit() string {
	// TODO import type
	return fmt.Sprintf("%s(%s)", c.typ.String(), c.e.emit())
}

type methodCall struct {
	receiver expr
	method   string
	args     []expr
}

func (m methodCall) emit() string {
	var args []string
	for _, arg := range m.args {
		args = append(args, arg.emit())
	}
	return fmt.Sprintf("%s.%s(%s)", m.receiver.emit(), m.method, strings.Join(args, ", "))
}

type fieldRef struct {
	e         expr
	fieldName string
}

func (f fieldRef) emit() string {
	return fmt.Sprintf("%s.%s", f.e.emit(), f.fieldName)
}

type stringLit string

func (s stringLit) emit() string {
	return fmt.Sprintf("%q", s)
}

type expr interface {
	emit() string
}

var _, _, _, _, _, _, _ expr = varRef(""), funCall{}, zeroValue{}, castType{}, methodCall{}, fieldRef{}, stringLit("")

func (c *container) createVar(namePrefix string) varRef {
	return c.doCreateVar(namePrefix, nil)
}

func (c *container) doCreateVar(namePrefix string, handle interface{}) varRef {
	v := varRef(namePrefix)
	i := 2
	for {
		_, ok := c.vars[v]
		if !ok {
			c.vars[v] = handle
			if handle != nil {
				c.reverseVars[handle] = v
			}
			return v
		}

		v = varRef(fmt.Sprintf("%s%d", namePrefix, i))
		i++
	}
}

func (c *container) getOrCreateVar(namePrefix string, handle interface{}) (v varRef, created bool) {
	if v, ok := c.reverseVars[handle]; ok {
		return v, false
	}

	return c.doCreateVar(namePrefix, handle), true
}

func (c *container) codegenWrite(v ...interface{}) {
	for _, x := range v {
		switch x := x.(type) {
		case expr:
			_, _ = fmt.Fprint(c.codegenOut, x.emit())
		default:
			_, _ = fmt.Fprint(c.codegenOut, x)
		}
	}
}

func (c *container) codegenWriteln(v ...interface{}) {
	c.codegenWrite(v...)
	_, _ = fmt.Fprintln(c.codegenOut)
}
