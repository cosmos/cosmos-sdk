#[repr(C)]
pub struct MsgSend {
    pub from: ::zeropb::Bytes,
    pub to: ::zeropb::Bytes,
    pub denom: ::zeropb::Str,
    pub amount: ::zeropb::Bytes,
}
unsafe impl zeropb::ZeroCopy for MsgSend {}
#[repr(C)]
pub struct MsgSendResponse {}
unsafe impl zeropb::ZeroCopy for MsgSendResponse {}
#[repr(C)]
pub struct QueryBalance {
    pub address: ::zeropb::Bytes,
    pub denom: ::zeropb::Str,
}
unsafe impl zeropb::ZeroCopy for QueryBalance {}
#[repr(C)]
pub struct QueryBalanceResponse {
    pub balance: ::zeropb::Bytes,
}
unsafe impl zeropb::ZeroCopy for QueryBalanceResponse {}
pub trait MsgServer {
    fn send(
        &self,
        ctx: &mut ::cosmossdk_core::Context,
        req: &MsgSend,
    ) -> ::cosmossdk_core::Result<MsgSendResponse>;
}
impl ::cosmossdk_core::Router for dyn MsgServer {}
pub struct MsgClient {
    connection: zeropb::Connection,
    service_id: u64,
}
impl MsgClient {
    pub fn send(
        &self,
        ctx: &mut ::cosmossdk_core::Context,
        req: ::zeropb::Root<MsgSend>,
    ) -> ::cosmossdk_core::Result<MsgSendResponse> {
        todo!()
    }
}
impl ::zeropb::Client for MsgClient {
    fn service_name(&self) -> &'static str {
        "example.bank.v1.Msg"
    }
}
pub trait QueryServer {
    fn balance(
        &self,
        ctx: &mut ::cosmossdk_core::Context,
        req: &QueryBalance,
    ) -> ::cosmossdk_core::Result<QueryBalanceResponse>;
}
impl ::cosmossdk_core::Router for dyn QueryServer {}
pub struct QueryClient {
    connection: zeropb::Connection,
    service_id: u64,
}
impl QueryClient {
    pub fn balance(
        &self,
        ctx: &mut ::cosmossdk_core::Context,
        req: ::zeropb::Root<QueryBalance>,
    ) -> ::cosmossdk_core::Result<QueryBalanceResponse> {
        todo!()
    }
}
impl ::zeropb::Client for QueryClient {
    fn service_name(&self) -> &'static str {
        "example.bank.v1.Query"
    }
}
