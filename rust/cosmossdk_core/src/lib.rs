#![no_std]
#![cfg(target_arch = "wasm32")]
mod wasm;

use zeropb;
use zeropb::{ClientConn, ZeroCopy};

#[link(wasm_import_module = "CosmosSDK")]
extern "C" {
    fn invoke_unary(ctx: u32, target: u32, req: *const u8, res: *mut u8) -> i32;
    fn resolve_service_method(name: *const u8) -> u32;
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
