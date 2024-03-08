use crate::router::Router;

pub trait ModuleService: Router {
    fn describe(descriptor: &mut crate::types::cosmos::core::v1alpha1::bundle::ModuleOutput) -> zeropb::Result<()>;
}