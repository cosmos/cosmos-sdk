package signing

type HasCustomSigners interface {
	GetCustomSigners() ([][]byte, error)
}
