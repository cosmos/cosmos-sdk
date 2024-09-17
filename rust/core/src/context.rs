use interchain_message_api::Address;
use crate::handler::{AccountAPI, AccountFactory, AccountHandler, AccountClient, ModuleAPI};
use crate::message::Message;
use crate::response::Response;

/// Context wraps a single message request (and possibly response as well) along with
/// the router callbacks necessary for making nested message calls.
pub struct Context {}

impl Context {
    /// This is the address of the account that is getting called.
    /// In a receiving account, this is the account's own address.
    pub fn address(&self) -> &Address {
        unimplemented!()
    }

    /// This is the address of the account which is making the message call.
    pub fn caller(&self) -> &Address {
        unimplemented!()
    }

    /// Dynamically invokes a module message.
    /// Static module client instances should be preferred wherever possible,
    /// so that static dependency analysis can be performed.
    pub fn dynamic_invoke_module<M: Message<true>>(&self, message: M) -> Response<M::Response, M::Error>
    {
        unimplemented!()
    }

    /// Dynamically invokes an account message.
    /// Static account client instances should be preferred wherever possible,
    /// so that static dependency analysis can be performed.
    pub fn dynamic_invoke_account<M: Message<false>>(&self, account: &Address, message: M) -> Response<M::Response, M::Error> {
        unimplemented!()
    }

    /// Get the address of the module implementing the given trait, client type or module message, if any.
    pub fn get_module_address<T: ModuleAPI>(&self) -> Response<&Address> {
        unimplemented!()
    }

    /// Create a new account with the given initialization data.
    pub fn new_account<H: AccountHandler>(&mut self, init: H::Init) -> Response<<<H as AccountAPI>::Factory as AccountFactory>::Client> {
        unimplemented!()
    }

    /// Create a temporary account with the given initialization data.
    /// Its address will be empty from the perspective of all observers,
    /// and it will not be persisted.
    pub fn new_temp_account<H: AccountHandler>(&mut self, init: H::Init) -> Response<<<H as AccountAPI>::Factory as AccountFactory>::Client> {
        unimplemented!()
    }
}

