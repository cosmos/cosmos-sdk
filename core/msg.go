package core

type Msg interface {
	GetSigners() []string
	ValidateBasic() error
}
