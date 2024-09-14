use interchain_message_api::Address;
use crate::message::Message;
use crate::Response;

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

    pub fn dynamic_invoke_module<M: Message<true>>(&self, message: M) -> Response<M::Response, M::Error>
    {
        unimplemented!()
    }

    pub fn dynamic_invoke_account<M: Message<false>>(&self, account: &Address, message: M) -> Response<M::Response, M::Error>
    {
        unimplemented!()
    }

    pub fn get_module_address<T>(&self) -> Response<&Address> {
        unimplemented!()
    }
}

