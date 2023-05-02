package genesis

// GenesisTxHandler is an interface that modules can implement to provide genesis state transitions
type GenesisTxHandler interface {
	ExecuteGenesisTx([]byte) error
}
