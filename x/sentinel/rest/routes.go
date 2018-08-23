package rest

import (
	"github.com/cosmos/cosmos-sdk/client/context"
	sentinel "github.com/cosmos/cosmos-sdk/x/sentinel"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/gorilla/mux"
)

func ServiceRoutes(ctx context.CoreContext, r *mux.Router, cdc *wire.Codec) {

	r.HandleFunc(
		"/register/vpn",       /// service provider
		registervpnHandlerFn(ctx, cdc),
	).Methods("POST")

	r.HandleFunc(
		"/register/master",   // master node
		registermasterdHandlerFn(ctx, cdc),
	).Methods("POST")

	r.HandleFunc(
		"/refund",      	// client
		RefundHandleFn(ctx, cdc),
	).Methods("POST")

	r.HandleFunc(
		"/master",  	// owner  or by vote
		deleteMasterHandlerFn(ctx, cdc),
	).Methods("DELETE")

	r.HandleFunc(
		"/vpn",       // master node deletes service provider
		deleteVpnHandlerFn(ctx, cdc),
	).Methods("DELETE")
	r.HandleFunc(
		"/vpn/pay",   // client
		PayVpnServiceHandlerFn(ctx, cdc),
	).Methods("POST")
	r.HandleFunc(
		"/send-sign",                     // Off-chain  Tx (client to service provider)
		SendSignHandlerFn(ctx, cdc),
	).Methods("POST")
	r.HandleFunc(
		"/vpn/getpayment",    // service provider to chain (from kv store)
		GetVpnPaymentHandlerFn(ctx, cdc),
	).Methods("POST")

}

func RegisterRoutes(ctx context.CoreContext, r *mux.Router, cdc *wire.Codec, keeper sentinel.Keeper) {

	ServiceRoutes(ctx, r, cdc)
	QueryRoutes(ctx, r, cdc, keeper)

}
