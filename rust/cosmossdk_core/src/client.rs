use crate::cosmos::core::v1alpha1::bundle::Input;

pub trait Client {
    fn describe(descriptor: &mut Input) -> zeropb::Result<()>;
}

impl Client for dyn zeropb::Client {
    fn describe(descriptor: &mut Input) -> zeropb::Result<()> {
        todo!()
    }
}