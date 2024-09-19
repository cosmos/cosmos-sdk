//! Handler traits for account and module handlers.
use ixc_message_api::Address;
use ixc_schema::StructCodec;
use crate::resource::{InitializationError, Initializer, Resource};

/// Handler trait for account and module handlers.
pub trait Handler {
    /// The parameter used for initializing the handler.
    type Init /*: StructCodec*/;
}

/// Account handler trait.
pub trait AccountHandler: AccountAPI + Handler {}

/// Account API trait.
pub trait AccountAPI {
    /// Account client factory type.
    type ClientFactory: AccountClientFactory;
}

/// Account factory trait.
pub trait AccountClientFactory: Resource {
    /// Account client type.
    type Client;

    /// Create a new account client with the given address.
    fn new_client(address: &Address) -> Self::Client;
}

/// Account client trait.
pub trait AccountClient {
    /// Get the address of the account.
    fn address(&self) -> &Address;
}

/// Module handler trait.
pub trait ModuleHandler: ModuleAPI + Handler {}

/// Module API trait.
pub trait ModuleAPI {
    /// Module client type.
    type Client: Resource;
}

/// Mixes in an account handler into another account handler.
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

/// Mixes in a module handler into another module handler.
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