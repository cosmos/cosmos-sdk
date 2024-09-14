use interchain_message_api::Address;
use crate::resource::{InitializationError, Initializer, Resource};

pub trait Handler {
    type Init;
}

pub trait AccountHandler: AccountAPI + Handler {
}

pub trait AccountAPI {
    type Factory: AccountFactory;
}

pub trait AccountFactory: Resource {
    type Ref;
    fn new_client(address: &Address) -> Self::Ref;
}

pub trait AccountRef {
    fn address(&self) -> &Address;
}

pub trait ModuleHandler: ModuleAPI + Handler {
}

pub trait ModuleAPI {
    type Ref: Resource;
}

pub struct AccountMixin<H: AccountHandler>(H);

unsafe impl<H: AccountHandler> Resource for AccountMixin<H> {
    unsafe fn new(initializer: &mut Initializer) -> Result<Self, InitializationError> {
        todo!()
    }
}

impl<H: AccountHandler> core::ops::Deref for AccountMixin<H> {
    type Target = H;

    fn deref(&self) -> &Self::Target {
        &self.0
    }
}

pub struct ModuleMixin<H: ModuleHandler>(H);

unsafe impl<H: ModuleHandler> Resource for ModuleMixin<H> {
    unsafe fn new(initializer: &mut Initializer) -> Result<Self, InitializationError> {
        todo!()
    }
}

impl<H: ModuleHandler> core::ops::Deref for ModuleMixin<H> {
    type Target = H;

    fn deref(&self) -> &Self::Target {
        &self.0
    }
}