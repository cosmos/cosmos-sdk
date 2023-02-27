package log

import "testing"

// reuse the same logger across all tests
var _testingLogger Logger

func NewTestingLogger() Logger {
	if _testingLogger != nil {
		return _testingLogger
	}

	if testing.Verbose() {
		_testingLogger = NewLoggerWithKV(ModuleKey, "test")
	} else {
		_testingLogger = NewNopLogger()
	}

	return _testingLogger
}
