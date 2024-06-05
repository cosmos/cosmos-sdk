package indexerbase

type Schema struct {
	Tables []Table
}

type Table struct {
	Name         string
	KeyColumns   []Column
	ValueColumns []Column
}

type Column struct {
	Name           string
	Type           Type
	Nullable       bool
	AddressPrefix  string
	EnumDefinition EnumDefinition
}

type EnumDefinition struct {
	Name   string
	Values []string
}
