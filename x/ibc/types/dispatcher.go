package types

import (
	"regexp"
)

// See baseapp/router.go

type Dispatcher interface {
	AddDispatch(string, Handler) Dispatcher
	Dispatch(string) Handler
}

type dispatch struct {
	d string
	h Handler
}

type dispatcher struct {
	dispatches []dispatch
}

func NewDispatcher() Dispatcher {
	return &dispatcher{
		dispatches: make([]dispatch, 0),
	}
}

var isAlpha = regexp.MustCompile(`^[a-zA-Z]+$`).MatchString

func (dis *dispatcher) AddDispatch(d string, h Handler) Dispatcher {
	if !isAlpha(d) {
		panic("dispatch expressions can only contain alphanumeric characters")
	}

	dis.dispatches = append(dis.dispatches, dispatch{d, h})

	return dis
}

func (dis *dispatcher) Dispatch(path string) Handler {
	for _, dispatch := range dis.dispatches {
		if dispatch.d == path {
			return dispatch.h
		}
	}
	return nil
}
