use crate::{Code, Context};

pub trait Module {
    type Config;

    // fn describe(descriptor: &mut crate::types::cosmos::core::v1alpha1::bundle::ModuleInitDescriptor) -> zeropb::Result<()>;
    fn route(&self, route_id: u64, ctx: &mut Context, req: *mut u8, res: *mut *mut u8) -> Code;
}

