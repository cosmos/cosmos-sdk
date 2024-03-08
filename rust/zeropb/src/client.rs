use num_enum::TryFromPrimitive;
use cosmossdk_core::{Code, Context};
use crate::{Root, ZeroCopy};
use crate::module_id::ModuleID;
use crate::root::RawRoot;
use crate::result::{Result};

pub trait Client {
    fn service_name(&self) -> &'static str;
}

/// cbindgen:ignore
#[cfg(target_arch = "wasm32")]
#[link(wasm_import_module = "ZeroPB")]
extern "C" {
    fn host_invoke(ctx: u64, method_id: u64, req: *const u8, res: *mut *mut u8) -> u64;
}

pub type Connection = extern "C" fn(u64, u64, usize, usize) -> u64;

pub fn connection_invoke<I: ZeroCopy, O: ZeroCopy>(connection: Connection, method_id: u64, ctx: &mut Context, req: Root<I>) -> Result<O> {
    unsafe {
        let mut res_ptr: *mut u8 = core::ptr::null_mut();
        let mut code = 0;
        if connection as *const () == core::ptr::null() {
            #[cfg(target_arch = "wasm32")]
            {
                code = host_invoke(method_id, ctx.id, req.unsafe_unwrap(), &mut res_ptr);
            }
            #[cfg(not(target_arch = "wasm32"))]
            {
                return Err(crate::Error {
                    code: Code::Internal,
                    msg: Root::empty(),
                });
            }
        } else {
            code = connection(method_id, ctx.id, req.unsafe_unwrap() as usize, res_ptr as usize);
        }
        let code = Code::try_from_primitive(code as u8).map_err(|_| crate::Error {
            code: Code::Internal,
            msg: Root::empty(),
        })?;
        match code {
            Code::Ok => {
                Ok(Root::unsafe_wrap(res_ptr))
            }
            _ => Err(crate::Error {
                code,
                msg: Root::unsafe_wrap(res_ptr),
            }),
        }
    }
}
