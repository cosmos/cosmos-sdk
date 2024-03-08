use crate::{Code, Context};

pub trait Server {
    // fn describe(descriptor: &mut crate::types::cosmos::core::v1alpha1::bundle::ModuleOutput) -> zeropb::Result<()>;
    fn route(&self, method_id: u64, ctx: &mut Context, req: *mut u8, res: *mut *mut u8) -> Code;
}

pub trait Client {

}
