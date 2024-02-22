package mock

import (
	"cosmossdk.io/log"
)

type logger struct{}

func (l logger) Info(msg string, keyVals ...any) {}

func (l logger) Error(msg string, keyVals ...any) {}

func (l logger) Debug(msg string, keyVals ...any) {}

func (l logger) Warn(msg string, keyVals ...any) {}

func (l logger) With(keyVals ...any) log.Logger { return l }

func (l logger) Impl() any { return l }
