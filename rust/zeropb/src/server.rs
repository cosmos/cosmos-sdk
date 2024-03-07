use crate::Context;
use crate::module_id::ModuleID;

trait Server<S> {
    fn route<S>(ctx: u64, method_id: u64, ctx: &mut Context, caller: &ModuleID, req: *const u8, res: *mut *mut u8) -> u64;
}