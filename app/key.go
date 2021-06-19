package app

type ModuleKey interface {
	moduleKey()
	ID() ModuleID
}

type ModuleID interface {
	moduleID()
	Name() string
}

type moduleKey struct {
	*moduleID
}

type moduleID struct {
	name string
}

var _ ModuleKey = &moduleKey{}

func (m *moduleKey) moduleKey() {}

func (m *moduleKey) ID() ModuleID {
	return m.moduleID
}

func (m *moduleID) moduleID() {
}

func (m *moduleID) Name() string {
	return m.name
}
