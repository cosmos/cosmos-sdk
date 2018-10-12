package rest

import (
	"encoding/json"
	"fmt"

	"net/http"
	"reflect"
	"strconv"
	"strings"

	ioutill "io/ioutil"

	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/libs/common"

	"github.com/cosmos/cosmos-sdk/client/context"
	ckeys "github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/crypto/keys"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	"github.com/cosmos/cosmos-sdk/x/sentinel"
	senttype "github.com/cosmos/cosmos-sdk/x/sentinel/types"
)

/**
* @api {get} /keys/seed To get seeds for generate keys.
* @apiName getSeeds
* @apiGroup Sentinel-Tendermint
* @apiSuccessExample Response:
{
* garden sunset night final child popular fall ostrich amused diamond lift stool useful brisk very half rice evil any behave merge shift ring chronic
* }
*/
/**
* @api {post} /keys To get account.
* @apiName getAccount
* @apiGroup Sentinel-Tendermint
* @apiParam {String} name Name Account holder name.
* @apiParam {String} password Password password for account.
* @apiParam {String} seed Seed seed words to get account.
* @apiError AccountAlreadyExists AccountName is  already exists
* @apiError AccountSeedsNotEnough Seed words are not enough
* @apiErrorExample AccountAlreadyExists-Response:
* {
*   Account with name XXXXX... already exists.
* }
* @apiErrorExample AccountSeedsNotEnough-Response:
* {
*  recovering only works with XXX word (fundraiser) or 24 word mnemonics, got: XX words
* }
* @apiSuccessExample Response:
*{
*    "name": "vpn",
*    "type": "local",
*    "address": "cosmosaccaddr1udntgzszesn7z3xm64hafvjlegrh38ukzw9m7g",
*    "pub_key": "cosmosaccpub1addwnpepqfjqadxwa9p8tvwhydsakyvkajxgyd0ycanv25u7yff7lqtkwuk8vqcy5cg",
*    "seed": "hour cram bike donor script fragile together derive capital joy glance morning athlete special hint scrub guitar view popular dream idle inquiry transfer often"
*}
 */
/**
* @api {post} /register/vpn To register VPN service provider.
* @apiName registerVPN
* @apiGroup Sentinel-Tendermint
* @apiParam {String} ip Ip address of VPN service provider.
* @apiParam {Number} upload_speed Upload Net speed of VPN service.
* @apiParam {Number} download_speed Download Net speed of VPN service.
* @apiParam {Number} price_per_gb Price per GB.
* @apiParam {String} enc_method Encryption method.
* @apiParam {Number} location_latitude  Latitude Location of service provider.
* @apiParam {Number} location_longitude  Longiude Location of service provider.
* @apiParam {String} location_city  City Location of service provider.
* @apiParam {String} location_country  Country Location of service provider.
* @apiParam {String} node_type  Node type.
* @apiParam {String} version version.
* @apiParam {String} name Account name of service provider.
* @apiParam {string} password password of account.
* @apiParam {Number} gas Gas value.
* @apiError AccountAlreadyExists VPN service provider already exists
* @apiError NetSpeedInvalidError Netspeed is Invalid
* @apiError IpAddressInvalidError IP address is Invalid
* @apiError Price_per_GBInvalidError Price per GB is Invalid
* @apiErrorExample AccountAlreadyExists-Response:
*{
 * checkTx failed: (1245197) Msg 0 failed: === ABCI Log ===
* Codespace: 19
* Code:      13
* ABCICode:  1245197
* Error:     --= Error =--
* Data: common.FmtError{format:"Address already Registered as VPN node", args:[]interface {}(nil)}
* Msg Traces:
* --= /Error =--
*
*=== /ABCI Log ===
*}
* @apiErrorExample NetSpeedInvalidError-Response:
*{
 * checkTx failed: (1245197) Msg 0 failed: === ABCI Log ===
* Codespace: 19
* Code:      13
* ABCICode:  1245197
* Error:     --= Error =--
* Data: common.FmtError{format:"NetSpeed is not Valid", args:[]interface {}(nil)}
* Msg Traces:
* --= /Error =--
*
*=== /ABCI Log ===
*}
* @apiErrorExample IpAddressInvalidError-Response:
*{
 * "  invalid Ip address."
*}
* @apiErrorExample Price_per_GBInvalidError-Response:
*{
 * checkTx failed: (1245197) Msg 0 failed: === ABCI Log ===
* Codespace: 19
* Code:      13
* ABCICode:  1245197
* Error:     --= Error =--
* Data: common.FmtError{format:"Price per GB is not Valid", args:[]interface {}(nil)}
* Msg Traces:
* --= /Error =--
*
*=== /ABCI Log ===
*}
* @apiSuccessExample Response:
*{
*   "Success": true,
*   "Hash": "CF8E073D624F7FA6A41C3CAD9B4A1DB693234225",
*   "Height": 343,
*   "Data": "eyJ0eXBlIjoic2VudGluZWwvcmVnaXN0ZXJ2cG4iLCJ2YWx1ZSI6eyJGc3BlZWQiOiIxMiIsIlBwZ2IiOiIyMyIsIkxvY2F0aW9uIjoiaHlkIn19",
*    "Tags": [
*        {
*            "key": "dnBuIHJlZ2lzdGVyZWQgYWRkcmVzcw==",
*            "value": "Y29zbW9zYWNjYWRkcjFlZ3RydjdxdGU0NnY2cXEzN3p0YzB2dzRuMmhrejZuempycDVhZQ=="
*        }
*		    ]
*}
*/

func NewResponse(success bool, hash string, height int64, data []byte, tags []common.KVPair) Response {
	//var res Response
	return Response{
		Success: success,
		Height:  height,
		Hash:    hash,
		Data:    data,
		Tags:    tags,
	}
}
func registervpnHandlerFn(ctx context.CoreContext, cdc *wire.Codec) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var a int64
		msg := MsgRegisterVpnService{}
		body, err := ioutill.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}
		err = json.Unmarshal(body, &msg)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Invalid  Msg Unmarshal function Request"))
			return
		}

		if !validateIp(msg.Ip) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("invalid Ip address."))
			return

		}
		if reflect.TypeOf(msg.Ppgb) != reflect.TypeOf(a) || msg.Ppgb < 0 {

			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(" entered invalid amount of price per Gb"))
			return
		}
		if msg.UploadSpeed <= 0 || msg.DownloadSpeed <= 0 {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(" entered invalid net speed details"))
			return
		}
		if msg.Latitude <= -90*10000 || msg.Longitude <= -180*10000 || msg.Latitude > 90*10000 || msg.Longitude > 180*10000 || msg.City == "" || msg.Country == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(" entered invalid  Location details"))
			return
		}
		if msg.NodeType == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(" Node type is required"))
			return
		}
		if msg.Version == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(" Version is required"))
			return
		}
		ctx = ctx.WithGas(msg.Gas)
		ctx = ctx.WithFromAddressName(msg.Localaccount)
		addr, err := ctx.GetFromAddress()
		if err != nil {
			sdk.ErrInvalidAddress("The given Address is Invalid")
		}
		ctx = ctx.WithDecoder(authcmd.GetAccountDecoder(cdc))
		acc, err := ctx.GetAccountNumber(addr)
		seq, err := ctx.NextSequence(addr)
		ctx = ctx.WithSequence(seq)
		ctx = ctx.WithAccountNumber(acc)
		if err != nil {
			w.Write([]byte("account number error"))
			w.Write([]byte(string(acc)))

		}

		msg1 := sentinel.NewMsgRegisterVpnService(addr, msg.Ip, msg.Ppgb, msg.UploadSpeed, msg.DownloadSpeed, msg.EncMethod, msg.Latitude, msg.Longitude, msg.City, msg.Country, msg.NodeType, msg.Version)

		txBytes, err := ctx.SignAndBuild(msg.Localaccount, msg.Password, []sdk.Msg{msg1}, cdc)

		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(err.Error()))
			return
		}

		res, err := ctx.BroadcastTx(txBytes)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		respon := NewResponse(true, res.Hash.String(), res.Height, res.DeliverTx.Data, res.DeliverTx.Tags)
		data, err := json.MarshalIndent(respon, "", " ")
		w.Write(data)
	}
	return nil
}

/**
* @api {post} /register/master To register Master Node.
* @apiName registerMasterNode
* @apiGroup Sentinel-Tendermint
* @apiParam {String} name  Account name of Master Node.
* @apiParam {Number} gas Gas value.
* @apiParam {string} password password of account.
* @apiError AccountAlreadyExists Master Node already exists
* @apiErrorExample AccountAlreadyExists-Response:
*{
* checkTx failed: (1245197) Msg 0 failed: === ABCI Log ===
* Codespace: 19
* Code:      13
* ABCICode:  1245197
* Error:     --= Error =--
* Data: common.FmtError{format:"Address already Registered as VPN node", args:[]interface {}(nil)}
* Msg Traces:
* --= /Error =--
*
*=== /ABCI Log ===
*}
* @apiSuccessExample Response:
{
*{
 *   "Success": true,
*    "Hash": "CF8E073D624F7FA6A41C3CAD9B4A1DB693234225",
*    "Height": 343,
*    "Data": "eyJ0eXBlIjoic2VudGluZWwvcmVnaXN0ZXJ2cG4iLCJ2YWx1ZSI6eyJGc3BlZWQiOiIxMiIsIlBwZ2IiOiIyMyIsIkxvY2F0aW9uIjoiaHlkIn19==",
*    "Tags": [
*        {
*             "key": "dnBuIHJlZ2lzdGVyZWQgYWRkcmVzcw==",
*             "value": "Y29zbW9zYWNjYWRkcjFlZ3RydjdxdGU0NnY2cXEzN3p0YzB2dzRuMmhrejZuempycDVhZQ=="
*         }
*             ]
* }
*/
func registermasterdHandlerFn(ctx context.CoreContext, cdc *wire.Codec) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		msg := MsgRegisterMasterNode{}
		var err error
		body, err := ioutill.ReadAll(r.Body)
		if err != nil {
			return
		}

		json.Unmarshal(body, &msg)
		ctx = ctx.WithFromAddressName(msg.Name)
		ctx = ctx.WithGas(msg.Gas)
		addr, err := ctx.GetFromAddress()
		if err != nil {
			sdk.ErrInvalidAddress("The given Address is Invalid")
		}
		ctx = ctx.WithGas(msg.Gas)
		ctx = ctx.WithDecoder(authcmd.GetAccountDecoder(cdc))

		acc, err := ctx.GetAccountNumber(addr)
		seq, err := ctx.NextSequence(addr)
		ctx = ctx.WithSequence(seq)
		ctx = ctx.WithAccountNumber(acc)
		if err != nil {
			w.Write([]byte("account number error"))
			w.Write([]byte(string(acc)))

		}

		msg1 := sentinel.NewMsgRegisterMasterNode(addr)

		txBytes, err := ctx.SignAndBuild(msg.Name, msg.Password, []sdk.Msg{msg1}, cdc)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(err.Error()))
			return
		}

		res, err := ctx.BroadcastTx(txBytes)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		respon := NewResponse(true, res.Hash.String(), res.Height, res.DeliverTx.Data, res.DeliverTx.Tags)
		data, err := json.MarshalIndent(respon, "", " ")
		w.Write(data)
	}
	return nil
}

/**
* @api {delete} /vpn To Delete VPN Node.
* @apiName  deleteVpnNode
* @apiGroup Sentinel-Tendermint
* @apiParam {String} address  Address of VPN Node which we want to delete.
* @apiParam {String} name AccountName of the person who is deleting the VPN node.
* @apiParam {string} password password of account.
* @apiParam {Number} gas Gas value.
* @apiError AccountNotExists VPN Node not exists
* @apiErrorExample AccountNotExists-Response:
*{
* checkTx failed: (1245197) Msg 0 failed: === ABCI Log ===
* Codespace: 19
* Code:      13
* ABCICode:  1245197
* Error:     --= Error =--
* Data: common.FmtError{format:"Account is not exist", args:[]interface {}(nil)}
* Msg Traces:
* --= /Error =--
*
*=== /ABCI Log ===
*}
* @apiSuccessExample Response:
{
 *   "Success": true,
 *   "Hash": "32EF9DFB6BC24D3159A8310F1AE438BED479466E",
 *   "Height": 3698,
 *   "Data": "FRTjZrQKAswn4UTeyJ0eXBlIjoic2VudGluZWWQiOiIxMiIsIlBwZ2IiOiIyMyIsIkxvY2F0aW9uIjoiaHlkIn19b1W/Usl/KB3iflg==",
 *   "Tags": [
 *       {
 *           "key": "ZGVsZXRlZCBWcG4gYWRkcmVzcw==",
 *           "value": "42a0CgLMJ+FE29Vv1LJfygd4n5Y="
 *      }
 *  ]
}
*/
func deleteVpnHandlerFn(ctx context.CoreContext, cdc *wire.Codec) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var msg MsgDeleteVpnUser
		var err error
		body, err := ioutill.ReadAll(r.Body)
		if err != nil {
			return
		}
		json.Unmarshal(body, &msg)
		if msg.Address == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(" entered invalid address."))
			return
		}
		Vaddr, err := sdk.AccAddressFromBech32(msg.Address)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}
		ctx = ctx.WithGas(msg.Gas)
		ctx = ctx.WithFromAddressName(msg.Name)
		addr, err := ctx.GetFromAddress()
		if err != nil {
			panic(err)
		}
		ctx = ctx.WithDecoder(authcmd.GetAccountDecoder(cdc))
		acc, err := ctx.GetAccountNumber(addr)
		seq, err := ctx.NextSequence(addr)
		ctx = ctx.WithSequence(seq)
		ctx = ctx.WithAccountNumber(acc)
		if err != nil {
			w.Write([]byte("account number error"))
			w.Write([]byte(string(acc)))

		}
		msg1 := sentinel.NewMsgDeleteVpnUser(addr, Vaddr)
		txBytes, err := ctx.SignAndBuild(msg.Name, msg.Password, []sdk.Msg{msg1}, cdc)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(err.Error()))
			return
		}

		res, err := ctx.BroadcastTx(txBytes)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		respon := NewResponse(true, res.Hash.String(), res.Height, res.DeliverTx.Data, res.DeliverTx.Tags)
		data, err := json.MarshalIndent(respon, "", " ")
		w.Write(data)
	}
	return nil
}

/**
* @api {delete} /master To Delete Master Node.
* @apiName  deleteMasterNode
* @apiGroup Sentinel-Tendermint
* @apiParam {String} address  Address of Master Node which we want to delete.
* @apiParam {String} name AccountName of the person who is deleting the Master node.
* @apiParam {string} password password of account.
* @apiParam {Number} gas Gas value.
* @apiError AccountNotExists Master Node not exists
* @apiErrorExample AccountNotExists-Response:
*{
* checkTx failed: (1245197) Msg 0 failed: === ABCI Log ===
* Codespace: 19
* Code:      13
* ABCICode:  1245197s
* Error:     --= Error =--
* Data: common.FmtError{format:"Account is not exist", args:[]interface {}(nil)}
* Msg Traces:
* --= /Error =--
*
*=== /ABCI Log ===
*}
* @apiSuccessExample Response:
{
 *   "Success": true,
 *   "Hash": "32EF9DFB6BC24D3159A8310F1AE438BED479466E",
 *   "Height": 3698,
 *   "Data": "FRTjZrQKAswn4UeyJ0eXBlIwZ2IiOiIyMyIsIkxvY2F0aW9uIjoiaHlkIn19Tb1W/Usl/KB3iflg==",
 *   "Tags": [
 *       {
 *           "key": "ZGVsZXRlZCBWcG4gYWRkcmVzcw==",
 *           "value": "42a0CgLMJ+FE29Vv1LJfygd4n5Y="
 *      }
 *  ]
}
*/
func deleteMasterHandlerFn(ctx context.CoreContext, cdc *wire.Codec) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var msg MsgDeleteMasterNode
		var err error
		body, err := ioutill.ReadAll(r.Body)
		if err != nil {
			return
		}
		json.Unmarshal(body, &msg)
		if msg.Address == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(" entered invalid address."))
			return
		}
		Maddr, err := sdk.AccAddressFromBech32(msg.Address)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}
		ctx = ctx.WithGas(msg.Gas)
		ctx = ctx.WithFromAddressName(msg.Name)
		addr, err := ctx.GetFromAddress()
		if err != nil {
			sdk.ErrInvalidAddress("The given Address is Invalid")
		}
		ctx = ctx.WithDecoder(authcmd.GetAccountDecoder(cdc))
		acc, err := ctx.GetAccountNumber(addr)
		seq, err := ctx.NextSequence(addr)
		ctx = ctx.WithSequence(seq)
		ctx = ctx.WithAccountNumber(acc)
		msg1 := sentinel.NewMsgDeleteMasterNode(addr, Maddr)
		txBytes, err := ctx.SignAndBuild(msg.Name, msg.Password, []sdk.Msg{msg1}, cdc)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(err.Error()))
			return
		}
		res, err := ctx.BroadcastTx(txBytes)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		respon := NewResponse(true, res.Hash.String(), res.Height, res.DeliverTx.Data, res.DeliverTx.Tags)
		data, err := json.MarshalIndent(respon, "", " ")
		w.Write(data)
	}
	return nil
}

func validateIp(host string) bool {
	parts := strings.Split(host, ".")

	if len(parts) < 4 {
		return false
	}

	for _, x := range parts {
		if i, err := strconv.Atoi(x); err == nil {
			if i < 0 || i > 255 {
				return false
			}
		} else {
			return false
		}

	}
	return true
}

/**
* @api {post} /vpn/pay To Pay for VPN service.
* @apiName  payVPN service
* @apiGroup Sentinel-Tendermint
* @apiParam {String} amount  Amount to pay for vpn service.
* @apiParam {String} vaddress Address of the vpn service provider.
* @apiParam {String} name Account name of Client
* @apiParam {string} password password of account.
* @apiParam {Number} gas Gas value.
* @apiParam {String} sig_name NewAccountName.
* @apiParam {String} sig_password NewAccountPassword.
* @apiError AccountNotExists VPN Node not exists
* @apiError AccountNameAlreadyExists The new account name is already exist
* @apiError InsufficientFunds Funds are less than 100
* @apiErrorExample AccountVPNNotExists-Response:
*{
 * checkTx failed: (1245197) Msg 0 failed: === ABCI Log ===
* Codespace: 1
* Code:      9
* ABCICode:  65545
* Error:     --= Error =--
* Data: common.FmtError{format:"VPN address is not registered", args:[]interface {}(nil)}
* Msg Traces:
* --= /Error =--
*
*=== /ABCI Log ===
*}
* @apiErrorExample AccountNameAlreadyExists-Response:
*{
	* " Account with name XXXXXX already exists."
*}
* @apiErrorExample InsufficientFunds-Response:
*{
	* "Funds must be Greaterthan or equals to 100"
*}
* @apiSuccessExample Response:
{
*   "Success": true,
*   "Hash": "D2C58CAFC580CC39A4CFAB4325991A9378AFE77D",
*   "Height": 1196,
*   "Data": "IjNwWGdHazB5MnBGceyJ0eXBlIjoic2VudGluZWwvcmVnaXN0ZXJ2cG4iLCJ2YWx1ZSI6eyJGc3BlZWQiOiIxMiIsIlBwZ2IiOiIyMyIsIkxvY2F0aW9uIjoiaHlkIn19TdZdWIwak5xIg==",
*   "Tags": [
*      {
*       "key": "c2VuZGVyIGFkZHJlc3M=",
*       "value": "Y29zbW9zYWNjYWRkcjFuY3hlbGpjcjRnOWhzdmw3amRuempkazNyNzYyamUzenk4bXU5MA=="
*      },
*     {
*      "key": "c2Vlc2lvbiBpZA==",
*      "value": "M3BYZ0drMHkycEZxN1l1YjBqTnE="
*     }
*          ]
}
*/
func PayVpnServiceHandlerFn(ctx context.CoreContext, cdc *wire.Codec) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var kb keys.Keybase
		msg := MsgPayVpnService{}
		body, err := ioutill.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}
		err = json.Unmarshal(body, &msg)

		kb, err = ckeys.GetKeyBase()
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
			return
		}

		seed := getSeed(keys.Secp256k1)
		if msg.SigName == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(" Enter the Name."))
			return
		}
		if msg.SigPassword == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(" Enter  the Password."))
			return
		}

		if msg.Coins == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(" invalid address."))
			return
		}
		if msg.Vpnaddr == "" {

			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(" entered invalid vpn address"))
			return
		}
		vaddr, err := sdk.AccAddressFromBech32(msg.Vpnaddr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}
		coins, err := sdk.ParseCoins(msg.Coins)
		if err != nil {

			sdk.ErrInternal("Parse Coins Failed")
		}
		coin := sdk.Coins{sdk.NewCoin(coins[0].Denom, 99)}
		if !coins.Minus(coin).IsPositive() {
			w.Write([]byte("Funds must be Greater than or equals to 100"))
			return
		}
		infos, err := kb.List()
		for _, i := range infos {
			if i.GetName() == msg.SigName {
				w.WriteHeader(http.StatusConflict)
				w.Write([]byte(fmt.Sprintf("Account with name %s already exists.", msg.SigName)))
				return
			}
		}
		info, err := kb.CreateKey(msg.SigName, seed, msg.SigPassword)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		ctx = ctx.WithFromAddressName(msg.Localaccount)
		ctx = ctx.WithGas(msg.Gas)
		ctx = ctx.WithDecoder(authcmd.GetAccountDecoder(cdc))
		addr, err := ctx.GetFromAddress()
		if err != nil {
			sdk.ErrInvalidAddress("The given Address is Invalid")
			return
		}
		acc, err := ctx.GetAccountNumber(addr)
		seq, err := ctx.NextSequence(addr)
		ctx = ctx.WithSequence(seq)
		ctx = ctx.WithAccountNumber(acc)
		msg1 := sentinel.NewMsgPayVpnService(coins, vaddr, addr, info.GetPubKey())
		txBytes, err := ctx.SignAndBuild(msg.Localaccount, msg.Password, []sdk.Msg{msg1}, cdc)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(err.Error()))
			return
		}

		res, err := ctx.BroadcastTx(txBytes)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		respon := NewResponse(true, res.Hash.String(), res.Height, res.DeliverTx.Data, res.DeliverTx.Tags)
		data, err := json.MarshalIndent(respon, "", " ")
		w.Write(data)
	}
	return nil
}

//To create client signature....... This is not a transaction......

/**
* @api {post} /send-sign To Create sigature of the client.
* @apiName  CreateSignature
* @apiGroup Sentinel-Tendermint
* @apiParam {String} name AccountName of the client.
* @apiParam {string} password password of account.
* @apiParam {String} session_id session-id.
* @apiParam {String} amount Amount to create signature.
* @apiParam {Number} counter Counter value of the sigature.
*@apiParam {Boolean} isfial boolean value for is this final signature or not.
* @apiSuccessExample Response:
* 10lz2f928xpzsyggqhc9mu80qj59vx0rc6sedxmsfhca8ysuhhtgqypar3h4ty0pgftwqygp6vm54drttw5grlz4p5n238cvzxe2vpxmu6hhnqvt0uxstg7et4vdqhm4v
 */

func SendSignHandlerFn() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var kb keys.Keybase
		msg := ClientSignature{}
		var err error
		body, err := ioutill.ReadAll(r.Body)
		if err != nil {
			return
		}
		err = json.Unmarshal(body, &msg)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Invalid  Msg Unmarshal function Request"))
			return
		}
		kb, err = ckeys.GetKeyBase()
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
			return
		}
		if msg.Localaccount == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(" invalid Account Name."))
			return
		}
		if msg.Password == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(" invalid Password."))
			return
		}
		coins, err := sdk.ParseCoins(msg.Coins)
		if err != nil {
			sdk.ErrInternal("Parse Coins Failed")
		}

		bz := senttype.ClientStdSignBytes(coins, []byte(msg.Sessionid), msg.Counter, msg.IsFinal)
		sign, _, err := kb.Sign(msg.Localaccount, msg.Password, bz)
		if err != nil {
			w.Write([]byte(" Signature failed"))
			return
		}
		s, err := senttype.GetBech32Signature(sign)
		if err != nil {
			w.Write([]byte("signature marshal failed"))
		}
		w.Write([]byte(s))
	}
	return nil
}

/**
* @api {post} /refund To Refund the balance of client.
* @apiName  Refund
* @apiGroup Sentinel-Tendermint
* @apiParam {String} name AccountName of the client.
* @apiParam {string} password password of account.
* @apiParam {String} session_id session-id.
* @apiParam {Number} gas Gas value.
* @apiError TimeInvalidError Time is not more than 24 hours
* @apiError InvalidSessionIdError SessionId is invalid
* @apiErrorExample TimeInvalidError-Response:
*{
 * checkTx failed: (1245197) Msg 0 failed: === ABCI Log ===
* Codespace: 19
* Code:      2
* ABCICode:  6551245
* Error:     --= Error =--
* Data: common.FmtError{format:"time is less than 24 hours  or the balance is negative or equal to zero", args:[]interface {}(nil)}
* Msg Traces:
* --= /Error =--
*
*=== /ABCI Log ===
*}
* @apiErrorExample InvalidSessionIdError-Response:
*{
 * checkTx failed: (1245197) Msg 0 failed: === ABCI Log ===
* Codespace: 19
* Code:      6
* ABCICode:  124545
* Error:     --= Error =--
* Data: common.FmtError{format:"Invalid SessionId", args:[]interface {}(nil)}
* Msg Traces:
* --= /Error =--
*
*=== /ABCI Log ===
*}
* @apiSuccessExample Response:
*{
*	{
 *   "Success": true,
 *   "Hash": "868B602828FA48F1D4A03D9D066EB42DEC483AA0",
 *   "Height": 1092,
 *   "Data": "Qwi/dQ1h0GcdrppVOeyJ0eXBlIjoic2VudGluZWwvcmVnaXN0yJGc3BlZWQiOiIxMiIsIlBwZ2IiOiIyMyIsIkxvY2F0aW9uIjoiaHlkIn192hhGfJVl3g=",
 *   "Tags": [
 *{
 *           "key": "Y2xpZW50IFJlZnVuZCBBZGRyZXNzOg==",
 *           "value": "Y29zbW9zYWNjYWRkcjFndnl0N2FnZHY4Z3h3OGR3bmYybms2cnByOGU5dDltY3hkeGV3cA=="
 *       }
 *   ]
*}
* }
*/

func RefundHandleFn(ctx context.CoreContext, cdc *wire.Codec) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		msg := MsgRefund{}
		var err error
		body, err := ioutill.ReadAll(r.Body)
		if err != nil {
			return
		}
		err = json.Unmarshal(body, &msg)
		if err != nil {
			sentinel.ErrUnMarshal("Unmarshal of Given Message Type is failed")

		}
		ctx = ctx.WithFromAddressName(msg.Name)
		ctx = ctx.WithGas(msg.Gas)
		addr, err := ctx.GetFromAddress()
		if err != nil {
			sdk.ErrInvalidAddress("The given Address is Invalid")
		}
		ctx = ctx.WithDecoder(authcmd.GetAccountDecoder(cdc))
		acc, err := ctx.GetAccountNumber(addr)
		seq, err := ctx.NextSequence(addr)
		ctx = ctx.WithSequence(seq)
		ctx = ctx.WithAccountNumber(acc)
		msg1 := sentinel.NewMsgRefund(addr, []byte(msg.Sessionid))
		txBytes, err := ctx.SignAndBuild(msg.Name, msg.Password, []sdk.Msg{msg1}, cdc)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(err.Error()))
			return
		}

		res, err := ctx.BroadcastTx(txBytes)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		respon := NewResponse(true, res.Hash.String(), res.Height, res.DeliverTx.Data, res.DeliverTx.Tags)
		data, err := json.MarshalIndent(respon, "", " ")
		w.Write(data)
	}
}

/**
* @api {post} /vpn/getpayment To get payment of vpn service
* @apiName  GetVPNPayment
* @apiGroup Sentinel-Tendermint
* @apiParam {String} amount Amount to send VPN node.
* @apiParam {String} session_id session-id.
* @apiParam {Number} counter Counter value.
* @apiParam {String} name Account name of client.
* @apiParam {Number} gas gas value.
* @apiParam {Boolean} isfinal is this final signature or not.
* @apiParam {string} password password of account.
* @apiParam {string} sign signature of the client.
* @apiError InvalidSessionId  SessionId is invalid
* @apiError SignatureVerificationFailed  Invalid signature
* @apiErrorExample InvalidSessionId-Response:
*{
 * checkTx failed: (1245197) Msg 0 failed: === ABCI Log ===
* Codespace: 19
* Code:      6
* ABCICode:  65545
* Error:     --= Error =--
* Data: common.FmtError{format:"Invalid session Id", args:[]interface {}(nil)}
* Msg Traces:
* --= /Error =--
*
*=== /ABCI Log ===
*}
* @apiErrorExample SignatureVerificationFailed-Response:
*{
 * checkTx failed: (1245197) Msg 0 failed: === ABCI Log ===
* Codespace: 19
* Code:      6
* ABCICode:  65545
* Error:     --= Error =--
* Data: common.FmtError{format:"signature verification failed", args:[]interface {}(nil)}
* Msg Traces:
* --= /Error =--
*
*=== /ABCI Log ===
*}
* @apiSuccessExample Response:
*{
*    "Success": true,
*    "Hash": "629F4603A5A4DE598B58DC494CCC38DB9FD96604",
*    "Height": 353,
*    "Data":"eyJ0eXBlIjoic2VudGluZWwvcmVnaXN0ZXJ2cG4iLCJ2YWx1ZSI6eyJGc3BlZWQiOiIxMiIsIlBwZ2IiOiyJ0eXBlIjoic2VudGluZWwvcmVnaXN0ZXJ2cG4iLCJ2YW9==",
*    "Tags": [
*        {
*            "key": "VnBuIFByb3ZpZGVyIEFkZHJlc3M6",
*            "value": "Y29zbW9zYWNjYWRkcjF1ZG50Z3pzemVzbjd6M3htNjRoYWZ2amxlZ3JoMzh1a3p3OW03Zw=="
*        },
*        {
*            "key": "c2Vlc2lvbklk",
*            "value": "WVZJRW81Y0dIczdkb09UVzRDTk4="
*        }
*    ]
*}
*/

func GetVpnPaymentHandlerFn(ctx context.CoreContext, cdc *wire.Codec) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		msg := MsgGetVpnPayment{}
		body, err := ioutill.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}
		err = json.Unmarshal(body, &msg)
		if err != nil {
			sentinel.ErrUnMarshal("UnMarshal of MessageType is failed")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if msg.Signature == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(" invalid signature ."))
			return
		}
		if msg.Coins == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(" invalid amount."))
			return
		}
		if msg.Sessionid == "" {

			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(" Session Id is wrong"))
			return
		}
		if msg.Counter < 0 {

			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Invalid Counter"))
			return
		}
		coins, err := sdk.ParseCoins(msg.Coins)
		if err != nil {
			sdk.ErrInternal("Parse Coins failed")
		}

		var sig crypto.Signature
		sig, err = senttype.GetBech64Signature(msg.Signature)
		cdc.UnmarshalBinaryBare([]byte(msg.Signature), &sig)
		//sig,err=cryptoAmino.SignatureFromBytes([]byte(msg.Signature))
		if err != nil {
			w.Write([]byte("Signature from string conversion failed"))
		}

		ctx = ctx.WithFromAddressName(msg.Localaccount)
		ctx = ctx.WithGas(msg.Gas)
		ctx = ctx.WithDecoder(authcmd.GetAccountDecoder(cdc))
		addr, err := ctx.GetFromAddress()
		if err != nil {
			w.Write([]byte("Invalid Address"))
			return
		}
		acc, err := ctx.GetAccountNumber(addr)
		seq, err := ctx.NextSequence(addr)
		ctx = ctx.WithSequence(seq)
		ctx = ctx.WithAccountNumber(acc)
		msg1 := sentinel.NewMsgGetVpnPayment(coins, []byte(msg.Sessionid), msg.Counter, addr, sig, msg.IsFinal)
		txBytes, err := ctx.SignAndBuild(msg.Localaccount, msg.Password, []sdk.Msg{msg1}, cdc)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(err.Error()))
			return
		}
		res, err := ctx.BroadcastTx(txBytes)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		respon := NewResponse(true, res.Hash.String(), res.Height, res.DeliverTx.Data, res.DeliverTx.Tags)
		data, err := json.MarshalIndent(respon, "", " ")
		w.Write(data)
	}
	return nil
}

/**
* @api {post} /send To send money to account.
* @apiName sendTokens
* @apiGroup Sentinel-Tendermint
* @apiParam {String} name Name Account holder name.
* @apiParam {String} password Password password for account.
* @apiParam {String} to To address.
* @apiParam {String} amount Amount to send.
* @apiParam {Number} gas gas value.
*
* @apiSuccessExample Response:
*{
*   "Success": true,
*   "Hash": "CF8E073D624F7FA6A41C3CAD9B4A1DB693234225",
*   "Height": 343,
*   "Data": "eyJ0eXBlIjoic2VudGluZWwvcmVnaXN0ZXJ2cG4iLCJ2YWx1ZSI6eyJGc3BlZWQiOiIxMiIsIlBwZ2IiOiIyMyIsIkxvY2F0aW9uIjoiaHlkIn19",
*    "Tags": [
*        {
*            "key": "dnBuIHJlZ2lzdGVyZWQgYWRkcmVzcw==",
*            "value": "Y29zbW9zYWNjYWRkcjFlZ3RydjdxdGU0NnY2cXEzN3p0YzB2dzRuMmhrejZuempycDVhZQ=="
*        }
*		    ]
*}
 */

func SendTokenHandlerFn(ctx context.CoreContext, cdc *wire.Codec) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		msg := SendTokens{}
		body, err := ioutill.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}
		err = json.Unmarshal(body, &msg)
		if err != nil {
			sentinel.ErrUnMarshal("UnMarshal of MessageType is failed")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if msg.Coins == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(" invalid amount."))
			return
		}
		coins, err := sdk.ParseCoins(msg.Coins)
		if err != nil {
			sdk.ErrInternal("Parse Coins failed")
		}
		ctx = ctx.WithFromAddressName(msg.Name)
		ctx = ctx.WithGas(msg.Gas)
		addr, err := ctx.GetFromAddress()
		if err != nil {
			sdk.ErrInvalidAddress("The given Address is Invalid")
		}
		to, err := sdk.AccAddressFromBech32(msg.ToAddress)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}
		ctx = ctx.WithDecoder(authcmd.GetAccountDecoder(cdc))
		acc, err := ctx.GetAccountNumber(addr)
		seq, err := ctx.NextSequence(addr)
		ctx = ctx.WithSequence(seq)
		ctx = ctx.WithAccountNumber(acc)
		msg1 := sentinel.NewMsgSendTokens(addr, coins, to)
		txBytes, err := ctx.SignAndBuild(msg.Name, msg.Password, []sdk.Msg{msg1}, cdc)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(err.Error()))
			return
		}
		res, err := ctx.BroadcastTx(txBytes)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		respon := NewResponse(true, res.Hash.String(), res.Height, res.DeliverTx.Data, res.DeliverTx.Tags)
		data, err := json.MarshalIndent(respon, "", " ")
		w.Write(data)

	}
}
