package tests

import (
	"fmt"
	"io"

	"cosmossdk.io/schema/logutil"
)

type prettyLogger struct {
	out io.Writer
}

func (l prettyLogger) Info(msg string, keyVals ...interface{}) {
	l.write("INFO", msg, keyVals...)
}

func (l prettyLogger) Warn(msg string, keyVals ...interface{}) {
	l.write("WARN", msg, keyVals...)
}

func (l prettyLogger) Error(msg string, keyVals ...interface{}) {
	l.write("ERROR", msg, keyVals...)
}

func (l prettyLogger) Debug(msg string, keyVals ...interface{}) {
	l.write("DEBUG", msg, keyVals...)
}

func (l prettyLogger) write(level, msg string, keyVals ...interface{}) {
	_, err := fmt.Fprintf(l.out, "%s: %s\n", level, msg)
	if err != nil {
		panic(err)
	}

	for i := 0; i < len(keyVals); i += 2 {
		_, err = fmt.Fprintf(l.out, "  %s: %v\n", keyVals[i], keyVals[i+1])
		if err != nil {
			panic(err)
		}
	}
}

var _ logutil.Logger = &prettyLogger{}
