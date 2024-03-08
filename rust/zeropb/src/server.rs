use crate::{Code, Context};

pub trait Server {
    fn route(&self, method_id: u64, ctx: &mut Context, req: *mut u8, res: *mut *mut u8) -> Code;
}
