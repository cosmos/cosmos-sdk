package internal

type Pin interface {
	Unpin()
}

type NoopPin struct{}

func (NoopPin) Unpin() {}
