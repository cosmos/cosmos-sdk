package rest

import (
	common "github.com/tendermint/tendermint/libs/common"
)

type MsgRegisterVpnService struct {
	Ip            string `json:"ip"`
	UploadSpeed   int64  `json:"upload_speed"`
	DownloadSpeed int64  `json:"download_speed"`
	Ppgb          int64  `json:"price_per_gb"`
	EncMethod     string `json:"enc_method"`
	Latitude      int64  `json:"location_latitude"`
	Longitude     int64  `json:"location_longitude"`
	City          string `json:"location_city"`
	Country       string `json:"location_country"`
	NodeType      string `json:"node_type"`
	Version       string `json:"version"`
	Localaccount  string `json:"name"`
	Password      string `json:"password"`
	Gas           int64  `json:"gas"`
}
type MsgRegisterMasterNode struct {
	Name     string `json:"name"`
	Gas      int64  `json:"gas"`
	Password string `json:"password"`
}

type MsgDeleteVpnUser struct {
	Address  string `json:"address", omitempty`
	Name     string `json:"name"`
	Password string `json:"password"`
	Gas      int64  `json:"gas"`
}
type MsgDeleteMasterNode struct {
	Address  string `json:"address", omitempty`
	Name     string `json:"name"`
	Password string `json:"password"`
	Gas      int64  `json:"gas"`
}
type MsgPayVpnService struct {
	Coins        string `json:"amount", omitempty`
	Vpnaddr      string `json:"vaddress", omitempty`
	Localaccount string `json:"name"`
	Password     string `json:"password"`
	Gas          int64  `json:"gas"`
	SigName      string `json:"sig_name"`
	SigPassword  string `json:"sig_password"`
}

type MsgGetVpnPayment struct {
	Coins        string `json:"amount"`
	Sessionid    string `json:"session_id"`
	Counter      int64  `json:"counter"`
	Localaccount string `json:"name"`
	Gas          int64  `json:"gas"`
	IsFinal      bool   `json:"isfinal"`
	Password     string `json:"password"`
	Signature    string `json:"sign"`
}

type MsgRefund struct {
	Name      string `json:"name"`
	Password  string `json:"password"`
	Sessionid string `json:"session_id", omitempty`
	Gas       int64  `json:"gas"`
}

type ClientSignature struct {
	Coins        string `json:"amount"`
	Sessionid    string `json:"session_id"`
	Counter      int64  `json:"counter"`
	IsFinal      bool   `json:"isfinal"`
	Localaccount string `json:"name"`
	Password     string `json:"password"`
}

type Response struct {
	Success bool            `json:"success"`
	Hash    string          `json:"hash"`
	Height  int64           `json:"height"`
	Data    []byte          `json:"data"`
	Tags    []common.KVPair `json:"tags"`
}

type SendTokens struct {
	Name      string `json:"name"`
	Password  string `json:"password"`
	ToAddress string `json:"to"`
	Coins     string `json:"amount"`
	Gas       int64  `json:"gas"`
}