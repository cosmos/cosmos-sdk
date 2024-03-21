package multisig

import (
	"context"
	"errors"
	"fmt"

	v1 "cosmossdk.io/api/cosmos/accounts/defaults/multisig/v1"
	"cosmossdk.io/collections"
	"cosmossdk.io/core/address"
	"cosmossdk.io/core/header"
	"cosmossdk.io/x/accounts/accountstd"
	aa_interface_v1 "cosmossdk.io/x/accounts/interfaces/account_abstraction/v1"
	"cosmossdk.io/x/tx/signing"
	"github.com/cosmos/cosmos-sdk/codec"
)

var (
	ParticipansPrefix = collections.NewPrefix(0)
	SequencePrefix    = collections.NewPrefix(1)
	ConfigPrefix      = collections.NewPrefix(2)
)

// Compile-time type assertions
var (
	_ accountstd.Interface = (*Account)(nil)
)

type Account struct {
	Participants collections.Map[[]byte, uint64]
	Sequence     collections.Sequence
	Config       collections.Item[*v1.Config]

	addrCodec address.Codec
	hs        header.Service

	signingHandlers *signing.HandlerMap
	customAlgos     map[string]SignatureHandler
}

// SignatureHandler allows custom signatures, it must be able to produce sign bytes
// and also verify the received signatures. It does NOT produce signatures as this
// is done off-chain.
// Note: implementers will probably want to have also a signing method in this handler
// so they can use it in the client.
type SignatureHandler interface {
	VerifySignature(signBytes []byte, pubkeys [][]byte) error

	// ValidatePubKey must error if the provided key does not comply with the
	// established format.
	ValidatePubKey([]byte) error

	// not sure if needed
	RecoverPubKey([]byte) ([]byte, error)

	GetSignBytes(msgs []byte, pubkeys [][]byte) error
}

type Options struct {
	CustomAlgorithms map[string]SignatureHandler
}

func NewAccount(name string, handlerMap *signing.HandlerMap, opts Options) accountstd.AccountCreatorFunc {
	return func(deps accountstd.Dependencies) (string, accountstd.Interface, error) {
		return name, &Account{
			Participants:    collections.NewMap(deps.SchemaBuilder, ParticipansPrefix, "participants", collections.BytesKey, collections.Uint64Value),
			Sequence:        collections.NewSequence(deps.SchemaBuilder, SequencePrefix, "sequence"),
			Config:          collections.NewItem(deps.SchemaBuilder, ConfigPrefix, "config", codec.CollValueV2[v1.Config]()),
			addrCodec:       deps.AddressCodec,
			signingHandlers: handlerMap,
		}, nil
	}
}

func (a *Account) Init(ctx context.Context, msg *v1.MsgInit) (*v1.MsgInitResponse, error) {
	if len(msg.PubKeys) != len(msg.Weights) {
		return nil, errors.New("the number of public keys and weights must be equal")
	}

	isValidAlgo := false
	allSupportedAlgos := []string{} // just to make the error more informative
	supportedModes := a.signingHandlers.SupportedModes()
	for i := range supportedModes {
		if msg.Config.Algo == supportedModes[i].String() {
			isValidAlgo = true
			break
		}
		allSupportedAlgos = append(allSupportedAlgos, supportedModes[i].String())
	}

	if !isValidAlgo {
		for i := range a.customAlgos {
			if msg.Config.Algo == i {
				isValidAlgo = true
				break
			}
			allSupportedAlgos = append(allSupportedAlgos, i)
		}
	}

	if !isValidAlgo {
		return nil, fmt.Errorf("unsupported signing algo: %s, must be one of: %v", msg.Config.Algo, allSupportedAlgos)
	}

	// set participants
	for i := range msg.PubKeys {
		if err := a.Participants.Set(ctx, msg.PubKeys[i], msg.Weights[i]); err != nil {
			return nil, err
		}
	}

	totalWeight := uint64(0)
	for _, weight := range msg.Weights {
		totalWeight += weight
	}

	if err := ValidateConfig(msg.Config, totalWeight); err != nil {
		return nil, err
	}

	return &v1.MsgInitResponse{}, nil
}

// Authenticate implements the authentication flow of an abstracted base account.
func (a Account) Authenticate(ctx context.Context, msg *aa_interface_v1.MsgAuthenticate) (*aa_interface_v1.MsgAuthenticateResponse, error) {
	if !accountstd.SenderIsAccountsModule(ctx) {
		return nil, errors.New("unauthorized: only accounts module is allowed to call this")
	}

	pubKey, signerData, err := a.computeSignerData(ctx)
	if err != nil {
		return nil, err
	}

	txData, err := a.getTxData(msg)
	if err != nil {
		return nil, err
	}

	gotSeq := msg.Tx.AuthInfo.SignerInfos[msg.SignerIndex].Sequence
	if gotSeq != signerData.Sequence {
		return nil, fmt.Errorf("unexpected sequence number, wanted: %d, got: %d", signerData.Sequence, gotSeq)
	}

	signMode, err := parseSignMode(msg.Tx.AuthInfo.SignerInfos[msg.SignerIndex].ModeInfo)
	if err != nil {
		return nil, fmt.Errorf("unable to parse sign mode: %w", err)
	}

	signature := msg.Tx.Signatures[msg.SignerIndex]

	signBytes, err := a.signingHandlers.GetSignBytes(ctx, signMode, signerData, txData)
	if err != nil {
		return nil, err
	}

	if !pubKey.VerifySignature(signBytes, signature) {
		return nil, errors.New("signature verification failed")
	}

	return &aa_interface_v1.MsgAuthenticateResponse{}, nil
}

func ValidateConfig(cfg *v1.Config, totalWeight uint64) error {
	// check for zero values
	if cfg.Threshold == 0 || cfg.Quorum == 0 || cfg.VotingPeriod == 0 {
		return errors.New("threshold, quorum and voting period must be greater than zero")
	}

	// threshold must be less than or equal to the total weight
	if totalWeight < uint64(cfg.Threshold) {
		return errors.New("threshold must be less than or equal to the total weight")
	}

	// quota must be less than or equal to the total weight
	if totalWeight < uint64(cfg.Quorum) {
		return errors.New("quorum must be less than or equal to the total weight")
	}

	return nil
}

func (a Account) QuerySequence(ctx context.Context, _ *v1.QuerySequence) (*v1.QuerySequenceResponse, error) {
	seq, err := a.Sequence.Peek(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.QuerySequenceResponse{Sequence: seq}, nil
}

// RegisterExecuteHandlers implements implementation.Account.
func (a *Account) RegisterExecuteHandlers(builder *accountstd.ExecuteBuilder) {
	accountstd.RegisterExecuteHandler(builder, a.Authenticate) // account abstraction
}

// RegisterInitHandler implements implementation.Account.
func (a *Account) RegisterInitHandler(builder *accountstd.InitBuilder) {
	accountstd.RegisterInitHandler(builder, a.Init)
}

// RegisterQueryHandlers implements implementation.Account.
func (a *Account) RegisterQueryHandlers(builder *accountstd.QueryBuilder) {
	accountstd.RegisterQueryHandler(builder, a.QuerySequence)
}
