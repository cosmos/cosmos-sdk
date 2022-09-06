package v2

const Paramspace = "baseapp"

// Parameter store keys for all the consensus parameter types.
var (
	ParamStoreKeyBlockParams     = []byte("BlockParams")
	ParamStoreKeyEvidenceParams  = []byte("EvidenceParams")
	ParamStoreKeyValidatorParams = []byte("ValidatorParams")
)
