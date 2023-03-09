struct ServiceImpl {
    service: String,
    methods: Vec<MethodImpl>,
}

struct MethodImpl {
    method: String,
    handler: MethodHandler,
}

#[repr(C, u8)]
enum MethodHandler {
    /// req is memory buffer representing the encoded request.
    /// res is a memory buffer that the response should be written to in encoded format. It is
    /// expected to be exactly 64kb in size. In the case of a non-Ok result, the response buffer
    /// should be written with a (possibly empty) null-terminated string representing the error
    /// message.
    Unary(extern fn(ctx: &Context, req: *const u8, res: *mut u8) -> Result<()>),
    /// req and res are the same as with Unary. send should be called every time the server has
    /// a new response to send. If it returns a non-Ok response the server should stop sending
    /// and return. If the request supports client-side cancellation, Cancel can be used to
    /// coordinate this and the server can still return an Ok response. The function should return
    /// when done sending. send is assumed to be blocking and will only return when the client is
    /// ready to receive more data.
    ServerStreaming(extern fn(ctx: &Context, req: *const u8, res: *mut u8, send: extern fn() -> Result<()>) -> Result<()>),
    /// req and res are the same as with Unary. recv should be called every time the server is
    /// ready to process more data. If it returns Ok, the server should expect that a new request
    /// was written to req. The server may cancel reading at any time by returning. If recv returns
    /// a non Ok response, the server should return. recv is assumed to be blocking and will only
    /// return when the client has sent more data.
    ClientStreaming(extern fn(ctx: &Context, req: *const u8, res: *mut u8, recv: extern fn() -> Result<()>) -> Result<()>),
}

#[repr(C)]
/// Context should be considered an opaque handle other than the caller field.
struct Context {
    /// caller will be populated only for internal message calls
    caller: String,
}

#[repr(u8)]
enum ErrorCode {
    Cancelled,
    Unknown,
    InvalidArgument,
    NotFound,
    AlreadyExists,
    PermissionDenied,
    ResourceExhausted,
    FailedPrecondition,
    Aborted,
    OutOfRange,
    Unimplemented,
    Internal,
    Unavailable,
    DataLoss,
    Unauthenticated,
}

#[repr(C, u8)]
enum Result<T> {
    Ok(T),
    Err(ResultCode),
}

struct ServiceClient {
    service: String,
    methods: Vec<MethodClient>,
}

struct MethodClient {
    method: String,
    handler: MethodClientFn,
}

#[repr(C, u8)]
enum MethodClientFn {
    Unary(extern fn(ctx: &Context, req: *const u8) -> Result<*const u8>),
    ServerStreaming(extern fn(ctx: &Context, req: *const u8) -> Result<ClientRecvFn>),
    ClientStreaming(extern fn(ctx: &Context) -> Result<ClientSendFn>),
}

#[repr(u8)]
enum RecvArgs {
    Continue,
    Cancel,
}

#[repr(C, u8)]
enum RecvPacket {
    HaveData(*const u8),
    Done,
}

type ClientRecvFn = extern fn(RecvArgs) -> Result<RecvPacket>;

#[repr(C, u8)]
enum SendPacket {
    HaveData(*const u8),
    Done,
}

#[repr(C, u8)]
enum SendResult {
    Continue,
    Done(*const u8),
}

type ClientSendFn = extern fn(SendPacket) -> Result<SendResult>;
