package rest

import (
	"encoding/hex"
	"net/http"

	"github.com/gorilla/mux"

	abci "github.com/tendermint/abci/types"
	sdk "github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/client/commands"
	"github.com/cosmos/cosmos-sdk/errors"
	"github.com/cosmos/cosmos-sdk/modules/base"
	"github.com/cosmos/cosmos-sdk/modules/nonce"
	"github.com/cosmos/cosmos-sdk/modules/roles"
	"github.com/tendermint/tmlibs/common"
)

// RoleInput encapsulates the fields needed to create a role
type RoleInput struct {
	// Role is a hex encoded string of the role name
	// for example, instead of "role" as the name, its
	// hex encoded version "726f6c65".
	Role string `json:"role" validate:"required,min=2"`

	MinimumSigners uint32 `json:"min_sigs" validate:"required,min=1"`

	Signers []sdk.Actor `json:"signers" validate:"required,min=1"`

	// Sequence is the user defined field whose purpose is to
	// prevent replay attacks when creating a role, since it
	// ensures that for a successful role creation, the previous
	// sequence number should have been looked up by the caller.
	Sequence uint32 `json:"seq" validate:"required,min=1"`
}

func decodeRoleHex(roleInHex string) ([]byte, error) {
	parsedRole, err := hex.DecodeString(common.StripHex(roleInHex))
	if err != nil {
		err = errors.WithMessage("invalid hex", err, abci.CodeType_EncodingError)
		return nil, err
	}
	return parsedRole, nil
}

// mux.Router registrars

// RegisterCreateRole is a mux.Router handler that exposes POST
// method access on route /build/create_role to create a role.
func RegisterCreateRole(r *mux.Router) error {
	r.HandleFunc("/build/create_role", doCreateRole).Methods("POST")
	return nil
}

func doCreateRole(w http.ResponseWriter, r *http.Request) {
	ri := new(RoleInput)
	if err := common.ParseRequestAndValidateJSON(r, ri); err != nil {
		common.WriteError(w, err)
		return
	}

	parsedRole, err := decodeRoleHex(ri.Role)
	if err != nil {
		common.WriteError(w, err)
		return
	}

	// Note the ordering of Tx wrapping matters:
	// 1. NonceTx
	tx := (nonce.Tx{}).Wrap()
	tx = nonce.NewTx(ri.Sequence, ri.Signers, tx)

	// 2. CreateRoleTx
	tx = roles.NewCreateRoleTx(parsedRole, ri.MinimumSigners, ri.Signers)

	// 3. ChainTx
	tx = base.NewChainTx(commands.GetChainID(), 0, tx)

	common.WriteSuccess(w, tx)
}

// End of mux.Router registrars
