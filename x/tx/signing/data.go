package signing

import txv1beta1 "cosmossdk.io/api/cosmos/tx/v1beta1"

type TxData struct {
	Tx                         *txv1beta1.Tx
	TxRaw                      *txv1beta1.TxRaw
	BodyHasUnknownNonCriticals bool
}
