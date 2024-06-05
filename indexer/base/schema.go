package indexerbase

type Schema struct {
	Tables []Table
}

type Table struct {
	Name        string
	KeyFields   []Field
	ValueFields []Field
}

type Field struct {
	Name string
	Type Type
}
