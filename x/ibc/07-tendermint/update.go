package tendermint

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

// CheckValidityAndUpdateState checks if the provided header is valid and updates
// the consensus state if appropriate. It returns an error if:
// - the client or header provided are not parseable to tendermint types
// - the header is invalid
// - header height is lower than the latest client height
// - header valset commit verification fails
//
// Tendermint client validity checking uses the bisection algorithm described
// in the [Tendermint spec](https://github.com/tendermint/spec/blob/master/spec/consensus/light-client.md).
func CheckValidityAndUpdateState(
	clientState clientexported.ClientState, header clientexported.Header, chainID string,
) (clientexported.ClientState, clientexported.ConsensusState, error) {
	tmClientState, ok := clientState.(ClientState)
	if !ok {
		return nil, nil, sdkerrors.Wrap(
			clienttypes.ErrInvalidClientType, "light client is not from Tendermint",
		)
	}

	tmHeader, ok := header.(Header)
	if !ok {
		return nil, nil, sdkerrors.Wrap(
			clienttypes.ErrInvalidHeader, "header is not from Tendermint",
		)
	}

	if err := checkValidity(tmClientState, tmHeader, chainID); err != nil {
		return nil, nil, err
	}

	tmClientState, consensusState := update(tmClientState, tmHeader)
	return tmClientState, consensusState, nil
}

// checkValidity checks if the Tendermint header is valid
//
// CONTRACT: assumes header.Height > consensusState.Height
func checkValidity(clientState ClientState, header Header, chainID string) error {
	if header.GetHeight() < clientState.LatestHeight {
		return sdkerrors.Wrapf(
			clienttypes.ErrInvalidHeader,
			"header height < latest client state height (%d < %d)", header.GetHeight(), clientState.LatestHeight,
		)
	}

	// basic consistency check
	if err := header.ValidateBasic(chainID); err != nil {
		return err
	}

	return header.ValidatorSet.VerifyCommit(header.ChainID, header.Commit.BlockID, header.Height, header.Commit)
}

// update the consensus state from a new header
func update(clientState ClientState, header Header) (ClientState, ConsensusState) {
	clientState.LatestHeight = header.GetHeight()
	consensusState := ConsensusState{
		Root:             commitment.NewRoot(header.AppHash),
		ValidatorSetHash: header.ValidatorSet.Hash(),
	}

	return clientState, consensusState
}
