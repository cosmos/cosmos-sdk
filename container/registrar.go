package container

type Registrar interface {
	RegisterProvider(fn interface{}) error
}
