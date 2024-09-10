package server

type DynamicConfig interface {
	GetString(string) string
	UnmarshalSub(string, any) (bool, error)
}
