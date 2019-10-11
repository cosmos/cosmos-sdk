package types

type Port struct {
	id string
}

// ID of the port
func (p Port) ID() string {
	return p.id
}

// // bindPort, expected to be called only at init time
// // TODO: is it safe to support runtime bindPort?
// func (man Manager) Port(id string) Port {
// 	if _, ok := man.ports[id]; ok {
// 		panic("port already occupied")
// 	}
// 	man.ports[id] = struct{}{}
// 	return Port{man, id}
// }

// // releasePort
// func (port Port) Release() {
// 	delete(port.channel.ports, port.id)
// }

// func (man Manager) IsValid(port Port) bool {
// 	_, ok := man.ports[port.id]
// 	return ok
// }
