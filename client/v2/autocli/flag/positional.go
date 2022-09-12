package flag

type positionalArg interface {
	Set(...string)
	FieldValueBinder
}
