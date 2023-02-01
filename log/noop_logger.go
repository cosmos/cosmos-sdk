package log

var _ Service = &NoOp{}

type NoOp struct{}

func NewNoOpLogger() *NoOp {
	return &NoOp{}
}

func (l NoOp) Debug(msg string, keyvals ...interface{}) {}
func (l NoOp) Info(msg string, keyvals ...interface{})  {}
func (l NoOp) Error(msg string, keyvals ...interface{}) {}

func (l NoOp) With(i ...interface{}) Service {
	return l
}
