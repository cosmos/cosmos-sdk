package container

type Scope struct {
	name string
}

func NewScope(name string) Scope {
	return Scope{name: name}
}

func (s Scope) Name() string {
	return s.name
}
