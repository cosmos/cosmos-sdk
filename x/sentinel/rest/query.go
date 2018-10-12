package rest

import (
	"encoding/json"
	"net/http"

	"github.com/cosmos/cosmos-sdk/client/context"
	sent "github.com/cosmos/cosmos-sdk/x/sentinel"
	senttype "github.com/cosmos/cosmos-sdk/x/sentinel/types"
	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/wire"
)

const (
	storeName = "sentinel"
)

/**
* @api {get} /session/{sessionId} To get session data.
* @apiName getSessionData
* @apiGroup Sentinel-Tendermint
* @apiSuccessExample Response:
*{
*    "name": "vpn",{
*    "TotalLockedCoins": [
*        {
*            "denom": "sentinel",
*            "amount": "10000000000"
*        }
*    ],
*    "ReleasedCoins": [
*        {
*            "denom": "sentinel",
*            "amount": "5000000000"
*        }
*    ],
*    "Counter": 1,
*    "Timestamp": 1537361017,
*    "VpnPubKey": [2,97,15,10,206,154,217,19,35,137,55,116,142,249,18,94,82,184,186,222,255,183,15,37,229,108,32,62,209,252,247,182,145],
*    "CPubKey": [3,157,182,213,107,56,95,22,24,197,116,75,236,23,60,131,180,160,198,244,216,103,74,189,19,147,141,25,242,109,176,252,39],
*    "CAddress": "cosmosaccaddr130q3n8kkpa9flav0sa5lefjunmruhchg5z6pzd",
*	    "Status": 1
*
*}
 */

func querySessionHandlerFn(cdc *wire.Codec, ctx context.CoreContext, k sent.Keeper) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		vars := mux.Vars(r)
		var clientSession senttype.Session
		sessionId := vars["sessionId"]
		res, err := ctx.QueryStore([]byte(sessionId), storeName)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("couldn't query session."))
			return
		}
		err = cdc.UnmarshalBinary(res, &clientSession)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("couldn't marshall query result. Error: "))
			return
		}
		bz, err := json.Marshal(clientSession)
		w.Write(bz)
	}
}
