package types

import "context"

// ValidatorPubKeyTypesChecker checks that the validator public key types allowed by
// a consensus params update still cover the consensus keys of the chain's validators.
// CometBFT rejects validator updates whose key type is not in
// ConsensusParams.Validator.PubKeyTypes, so removing a type that is still in use halts
// the chain at the next power change of such a validator.
type ValidatorPubKeyTypesChecker interface {
	ValidateValidatorPubKeyTypes(ctx context.Context, pubKeyTypes []string) error
}
