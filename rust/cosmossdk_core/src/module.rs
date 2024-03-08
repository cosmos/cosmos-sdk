use crate::router::Router;

pub trait Module: Router {
    fn describe(descriptor: &mut crate::types::cosmos::core::v1alpha1::bundle::ModuleInitDescriptor) -> zeropb::Result<()>;
}

pub trait ModuleBundle: Router {
    fn describe(descriptor: &mut crate::types::cosmos::core::v1alpha1::bundle::ModuleBundleDescriptor) -> zeropb::Result<()>;
}

