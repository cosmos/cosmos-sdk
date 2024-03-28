package stf

import (
	// "fmt"
	"unsafe"
)

// Container is a map that stores objects in any
type Container struct {
	m map[string]any
}

func (c Container) Set(prefix []byte, value any) {
	_, exists := c.m[unsafeString(prefix)]
	if exists {
		c.m[unsafeString(prefix)] = value
	}
	c.m[string(prefix)] = value
}

func (c Container) Get(prefix []byte) (value any, ok bool) {
	value, ok = c.m[unsafeString(prefix)]
	return
}

func (c Container) Remove(prefix []byte) {
	// tempMap := make(map[string]any)
	// for key := range c.m {
	// 	if key == unsafeString(prefix) {
	// 		continue
	// 	}
	// 	v, _ := c.m[key]
	// 	tempMap[key] = v
	// }
	// fmt.Println("tempMap", tempMap)
	// c.m = tempMap
	// fmt.Println("c", c)

	delete(c.m, unsafeString(prefix))
}

// ModuleContainer is a map that stores module objects
// key is module name
type ModuleContainer struct {
	m map[string]Container
}

func NewModuleContainer() ModuleContainer {
	return ModuleContainer{
		m: make(map[string]Container),
	}
}

// Get container of a module from module name in bytes
func (m ModuleContainer) GetContainer(address []byte) Container {
	v, ok := m.m[unsafeString(address)]
	if ok {
		return v
	}
	kc := Container{m: make(map[string]any)}
	m.m[string(address)] = kc
	return kc
}

func unsafeString(b []byte) string { return *(*string)(unsafe.Pointer(&b)) }
