package server

type DynamicConfig interface {
	Get(string) any
	GetString(string) string
}
