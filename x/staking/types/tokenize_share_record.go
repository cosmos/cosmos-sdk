package types

import (
	"fmt"
	"strconv"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	"github.com/pkg/errors"
)

func (r TokenizeShareRecord) GetModuleAddress() sdk.AccAddress {
	// NOTE: The module name is intentionally hard coded so that, if this
	// function were to move to a different module in future SDK version,
	// it would not break all the address lookups
	moduleName := "lsm"
	return address.Module(moduleName, []byte(r.ModuleAccount))
}

func (r TokenizeShareRecord) GetShareTokenDenom() string {
	return fmt.Sprintf("%s/%d", strings.ToLower(r.Validator), r.Id)
}

func ParseShareTokenDenom(denom string) (TokenizeShareRecord, error) {
	record := TokenizeShareRecord{}

	denomParts := strings.Split(denom, "/")
	if partsLen := len(denomParts); partsLen != 2 {
		err := errors.Errorf("wrong number of segments in share token denom: %d (expected 2)", partsLen)
		return record, err
	}

	valAddress, err := sdk.ValAddressFromBech32(denomParts[0])
	if err != nil {
		err = errors.Wrap(err, "failed to parse val address part")
		return record, err
	}

	recordId, err := strconv.ParseUint(denomParts[1], 10, 64)
	if err != nil {
		err = errors.Wrap(err, "failed to parse recordId part")
		return record, err
	}

	record.Id = uint64(recordId)
	record.Validator = valAddress.String()
	record.ModuleAccount = fmt.Sprintf("%s%d", TokenizeShareModuleAccountPrefix, recordId)

	return record, nil
}
