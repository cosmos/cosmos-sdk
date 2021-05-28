package authn

type Credential interface {
	Address() []byte
}
