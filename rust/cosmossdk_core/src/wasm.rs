#![no_std]

/// cbindgen:ignore
#[link(wasm_import_module = "CosmosSDK")]
extern "C" {
    fn resolve_service_method(name: *const u8) -> u32;
    fn register_unary_method(name: *const u8, method_id: u32);
    fn invoke_unary_method(ctx: u32, method_id: u32, req: *const u8, res: *mut u8) -> i32;
    fn store_get(ctx: u32, key: *const u8, len: u32) -> i64;
    fn store_set(ctx: u32, key: *const u8, len: u32, value: *const u8, vlen: u32);
    fn store_delete(ctx: u32, key: *const u8, len: u32);
    fn store_iter(ctx: u32, start: *const u8, len: u32, end: *const u8, elen: usize, iter_buf: *mut u8);
    fn store_iter_next(iter: *mut u8) -> i32;
    fn store_iter_release(iter: *mut u8);
}

/// cbindgen:ignore
#[link(wasm_import_module = "CosmosSDK")]
type ModuleInitFn = unsafe extern "C" fn(init_data: *const ModuleInitData) -> i32;

/// cbindgen:ignore
#[link(wasm_import_module = "CosmosSDK")]
type UnaryMethodHandler = unsafe extern "C" fn(ctx: u32, req: *const u8, res: *mut u8) -> i32;

/// cbindgen:ignore
#[link(wasm_import_module = "CosmosSDK")]
#[repr(C)]
struct ModuleInitData {
    config: *const u8,
    config_len: u32,
    register_unary_method: unsafe extern "C" fn(name: *const u8, handler: UnaryMethodHandler),
}

/// cbindgen:ignore
#[link(wasm_import_module = "CosmosSDK")]
#[repr(C)]
struct IterBuf {
    key: *mut u8,
    len: usize,
    value: *mut u8,
    vlen: usize,
}

pub fn test1() {
    unsafe {
        invoke_unary(0, 0, 0 as *const u8, 0 as *mut u8);
    }
}

struct Client {}

impl ClientConn<i32, i32> for Client {
    fn resolve_unary(&self, method: &str) -> zeropb::Handler<'_, i32, i32> {
        let target = unsafe { resolve_service_method(method.as_ptr()) };
        |ctx, req, resp| {
            unsafe {
                invoke_unary(ctx, target, req as *const u8, 0 as *mut u8);
                Ok(())
            }
            Ok(())
        }
    }
}

trait Module {
    type Config
        where
            Self: ZeroCopy;

    fn init(
        config: Config,
        client: Client,
        service_registry: &mut zeropb::ServiceRegistry,
    ) -> anyhow::Result<()>;
}

struct ModuleSet {}
