//! Handler traits for account and module handlers.
use ixc_message_api::AccountID;
use ixc_message_api::handler::RawHandler;
use ixc_schema::codec::Codec;
use ixc_schema::SchemaValue;
use ixc_schema::value::OptionalValue;
use crate::Context;
use crate::resource::{InitializationError, Resources};
use crate::routes::Router;

/// Handler trait for account and module handlers.
pub trait Handler: RawHandler + Router + Resources + ClientFactory {
    /// The name of the handler.
    const NAME: &'static str;
    /// The parameter used for initializing the handler.
    type Init<'a>: OptionalValue<'a>;
    /// The codec used for initializing the handler.
    type InitCodec: Codec + Default;
}

/// Account API trait.
pub trait HandlerAPI: Router {
    /// Account client factory type.
    type ClientFactory: ClientFactory;
}

/// Account factory trait.
pub trait ClientFactory {
    /// Account client type.
    type Client: Client;

    /// Create a new account client with the given address.
    fn new_client(account_id: AccountID) -> Self::Client;
}

/// Account client trait.
pub trait Client {
    /// Get the address of the account.
    fn account_id(&self) -> AccountID;
}

// /// Module handler trait.
// pub trait ModuleHandler: ModuleAPI + Handler {}
//
// /// Module API trait.
// pub trait ModuleAPI {
//     /// Module client type.
//     type Client: Resource;
// }
//
/// Mixes in an account handler into another account handler.
pub struct Mixin<H: Handler>(H);

// unsafe impl<H: Handler> Resource for Mixin<H> {
//     unsafe fn new(initializer: &mut Initializer) -> Result<Self, InitializationError> {
//         todo!()
//     }
// }

impl<H: Handler> core::ops::Deref for Mixin<H> {
    type Target = H;

    fn deref(&self) -> &Self::Target {
        &self.0
    }
}

//
// /// Mixes in a module handler into another module handler.
// pub struct ModuleMixin<H: ModuleHandler>(H);
//
// unsafe impl<H: ModuleHandler> Resource for ModuleMixin<H> {
//     unsafe fn new(initializer: &mut Initializer) -> Result<Self, InitializationError> {
//         todo!()
//     }
// }
//
// impl<H: ModuleHandler> core::ops::Deref for ModuleMixin<H> {
//     type Target = H;
//
//     fn deref(&self) -> &Self::Target {
//         &self.0
//     }
// }