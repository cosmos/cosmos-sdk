package container

type Scope interface {
	isScope()

	Name() string
}

func NewScope(name string) Scope {
	return &scope{name: name}
}

type scope struct {
	name string
}

func (s *scope) isScope() {}

func (s *scope) Name() string {
	return s.name
}
