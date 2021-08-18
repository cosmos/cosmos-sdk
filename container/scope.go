package container

// Scope is a special type used to define a provider scope.
//
// Special scoped constructors can be used with Provide by declaring a
// constructor with its first input parameter of type Scope. These constructors
// should construct an unique value for each dependency based on scope and will
// be called at most once per scope.
//
// Constructors passed to ProvideWithScope can also declare an input parameter
// of type Scope to retrieve their scope.
type Scope interface {
	isScope()

	// Name returns the name of the scope which is unique within a container.
	Name() string
}

// NewScope creates a new scope with the provided name. Only one scope with a
// given name can be created per container.
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
