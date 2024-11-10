package types

const (
	// ModuleName defines the name of the x/consensus module.
	ModuleName = "consensus"

	// StoreKey defines the module's store key.
	StoreKey = ModuleName

	// block params key
	BlockParamsKey       = ModuleName + "block_params"
	ValidatorKeyTypesKey = ModuleName + "validator_params"
	EvidenceKeyTypesKey  = ModuleName + "evidence_params"
)

var (
	ByteBlockParamsKey       = []byte(BlockParamsKey)
	ByteValidatorKeyTypesKey = []byte(ValidatorKeyTypesKey)
	ByteEvidenceKeyTypesKey  = []byte(EvidenceKeyTypesKey)
)
