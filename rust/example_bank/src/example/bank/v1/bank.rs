#[repr(C)]
pub struct MsgSend {
    pub from: ::zeropb::Str,
    pub to: ::zeropb::Str,
    pub denom: ::zeropb::Str,
    pub amount: u64,
}
unsafe impl zeropb::ZeroCopy for MsgSend {}
#[repr(C)]
pub struct MsgSendResponse {}
unsafe impl zeropb::ZeroCopy for MsgSendResponse {}
#[repr(C)]
pub struct QueryBalance {
    pub address: ::zeropb::Str,
    pub denom: ::zeropb::Str,
}
unsafe impl zeropb::ZeroCopy for QueryBalance {}
#[repr(C)]
pub struct QueryBalanceResponse {
    pub balance: u64,
}
unsafe impl zeropb::ZeroCopy for QueryBalanceResponse {}
trait MsgServer {
    fn send(
        &self,
        ctx: &mut ::zeropb::Context,
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
        ctx: &mut ::zeropb::Context,
        req: *mut u8,
        res: *mut *mut u8,
    ) -> ::zeropb::Code {
        unsafe {
            let result: ::zeropb::RawResult<*mut u8> = match method_id {
                1u64 => {
                    self.send(ctx, &*(req as *const MsgSend))
                        .map(|res| res.unsafe_unwrap())
                }
                _ => return ::zeropb::Code::Unimplemented,
            };
            match result {
                Ok(ptr) => {
                    *res = ptr;
                    ::zeropb::Code::Ok
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
struct MsgClient {
    connection: zeropb::Connection,
    service_id: u64,
}
impl MsgClient {
    fn send(
        &self,
        ctx: &mut zeropb::Context,
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
trait QueryServer {
    fn balance(
        &self,
        ctx: &mut ::zeropb::Context,
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
        ctx: &mut ::zeropb::Context,
        req: *mut u8,
        res: *mut *mut u8,
    ) -> ::zeropb::Code {
        unsafe {
            let result: ::zeropb::RawResult<*mut u8> = match method_id {
                1u64 => {
                    self.balance(ctx, &*(req as *const QueryBalance))
                        .map(|res| res.unsafe_unwrap())
                }
                _ => return ::zeropb::Code::Unimplemented,
            };
            match result {
                Ok(ptr) => {
                    *res = ptr;
                    ::zeropb::Code::Ok
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
struct QueryClient {
    connection: zeropb::Connection,
    service_id: u64,
}
impl QueryClient {
    fn balance(
        &self,
        ctx: &mut zeropb::Context,
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
