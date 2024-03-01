use num_enum::TryFromPrimitive;
use crate::{Context, Root, ZeroCopy};
use crate::root::RawRoot;
use crate::result::{Result};

/// cbindgen:ignore
#[cfg(target_arch = "wasm32")]
#[link(wasm_import_module = "ZeroPB")]
extern "C" {
    fn host_invoke(ctx: u64, method_id: u64, req: *const u8, res: *mut *mut u8) -> u32;
}

pub type Connection = extern "C" fn(u64, u64, *const u8, *mut *mut u8) -> u32;

pub fn connection_invoke<I: ZeroCopy, O: ZeroCopy>(connection: Connection, service_id: u64, ctx: &Context, req: Root<I>) -> Result<O> {
    unsafe {
        let mut res_ptr: *mut u8 = core::ptr::null_mut();
        let mut code = 0;
        if connection as *const () == core::ptr::null() {
            #[cfg(target_arch = "wasm32")]
            {
                code = host_invoke(service_id, ctx.id, req.unsafe_unwrap(), &mut res_ptr);
            }
            #[cfg(not(target_arch = "wasm32"))]
            {
                return Err(crate::Error {
                    code: crate::Code::Internal,
                    msg: Root::empty(),
                });
            }
        } else {
            code = connection(service_id, ctx.id, req.unsafe_unwrap(), &mut res_ptr);
        }
        let code = crate::Code::try_from_primitive(code as u8).map_err(|_| crate::Error {
            code: crate::Code::Internal,
            msg: Root::empty(),
        })?;
        match code {
            crate::Code::Ok => {
                Ok(Root::unsafe_wrap(res_ptr))
            }
            _ => Err(crate::Error {
                code,
                msg: Root::unsafe_wrap(res_ptr),
            }),
        }
    }
}
