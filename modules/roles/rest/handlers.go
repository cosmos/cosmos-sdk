package rest

import (
	"encoding/hex"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/spf13/viper"

	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/basecoin"
	"github.com/tendermint/basecoin/client/commands"
	"github.com/tendermint/basecoin/client/commands/query"
	"github.com/tendermint/basecoin/errors"
	"github.com/tendermint/basecoin/modules/base"
	"github.com/tendermint/basecoin/modules/nonce"
	"github.com/tendermint/basecoin/modules/roles"
	"github.com/tendermint/basecoin/stack"
	"github.com/tendermint/tmlibs/common"
)

// RoleInput encapsulates the fields needed to create a role
type RoleInput struct {
	// Role is a hex encoded string of the role name
	// for example, instead of "role" as the name, its
	// hex encoded version "726f6c65".
	Role string `json:"role" validate:"required,min=2"`

	MinimumSigners uint32 `json:"min_sigs" validate:"required,min=1"`

	Signers []basecoin.Actor `json:"signers" validate:"required,min=1"`

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

func doQueryRole(w http.ResponseWriter, r *http.Request) {
	args := mux.Vars(r)
	roleInHex := args["role"]
	parsedRole, err := decodeRoleHex(roleInHex)
	if err != nil {
		common.WriteError(w, err)
		return
	}

	recvRole := new(roles.Role)
	key := stack.PrefixedKey(roles.NameRole, parsedRole)
	prove := !viper.GetBool(commands.FlagTrustNode)
	height, err := query.GetParsed(key, recvRole, prove)
	if err != nil {
		common.WriteError(w, err)
		return
	}

	if err := query.OutputProof(recvRole, height); err != nil {
		common.WriteError(w, err)
		return
	}
	common.WriteSuccess(w, recvRole)
}

// RegisterQueryRole is a mux.Router handler that exposes GET
// method access on route /query/role/{theRole} to query for a role.
func RegisterQueryRole(r *mux.Router) error {
	r.HandleFunc("/query/role/{role}", doQueryRole).Methods("GET")
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
