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
    ) -> ::zeropb::Result<MsgSendResponse>;
}
impl ::zeropb::Server for dyn MsgServer {
    fn service_name(&self) -> &'static str {
        "example.bank.v1.Msg"
    }
    fn route(
        &self,
        method_id: u64,
        ctx: &mut ::cosmossdk_core::Context,
        req: *mut u8,
        res: *mut *mut u8,
    ) -> ::cosmossdk_core::Code {
        unsafe {
            let result: ::zeropb::RawResult<*mut u8> = match method_id {
                1u64 => {
                    self.send(ctx, &*(req as *const MsgSend))
                        .map(|res| res.unsafe_unwrap())
                }
                _ => return ::cosmossdk_core::Code::Unimplemented,
            };
            match result {
                Ok(ptr) => {
                    *res = ptr;
                    ::cosmossdk_core::Code::Ok
                }
                Err(err) => {
                    let ptr = err.msg.unsafe_unwrap();
                    if ptr != core::ptr::null_mut() {
                        *res = ptr;
                    }
                    err.code
                }
            }
        }
    }
}
pub struct MsgClient {
    connection: zeropb::Connection,
    service_id: u64,
}
impl MsgClient {
    pub fn send(
        &self,
        ctx: &mut cosmossdk_core::Context,
        req: zeropb::Root<MsgSend>,
    ) -> zeropb::Result<MsgSendResponse> {
        ::zeropb::connection_invoke(self.connection, 1u64, ctx, req)
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
    ) -> ::zeropb::Result<QueryBalanceResponse>;
}
impl ::zeropb::Server for dyn QueryServer {
    fn service_name(&self) -> &'static str {
        "example.bank.v1.Query"
    }
    fn route(
        &self,
        method_id: u64,
        ctx: &mut ::cosmossdk_core::Context,
        req: *mut u8,
        res: *mut *mut u8,
    ) -> ::cosmossdk_core::Code {
        unsafe {
            let result: ::zeropb::RawResult<*mut u8> = match method_id {
                1u64 => {
                    self.balance(ctx, &*(req as *const QueryBalance))
                        .map(|res| res.unsafe_unwrap())
                }
                _ => return ::cosmossdk_core::Code::Unimplemented,
            };
            match result {
                Ok(ptr) => {
                    *res = ptr;
                    ::cosmossdk_core::Code::Ok
                }
                Err(err) => {
                    let ptr = err.msg.unsafe_unwrap();
                    if ptr != core::ptr::null_mut() {
                        *res = ptr;
                    }
                    err.code
                }
            }
        }
    }
}
pub struct QueryClient {
    connection: zeropb::Connection,
    service_id: u64,
}
impl QueryClient {
    pub fn balance(
        &self,
        ctx: &mut cosmossdk_core::Context,
        req: zeropb::Root<QueryBalance>,
    ) -> zeropb::Result<QueryBalanceResponse> {
        ::zeropb::connection_invoke(self.connection, 1u64, ctx, req)
    }
}
impl ::zeropb::Client for QueryClient {
    fn service_name(&self) -> &'static str {
        "example.bank.v1.Query"
    }
}
