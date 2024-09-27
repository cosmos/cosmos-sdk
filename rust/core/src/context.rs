use bump_scope::BumpScope;
use crate::handler::{AccountAPI, AccountClientFactory, AccountHandler, ModuleAPI};
use crate::message::Message;
use crate::response::Response;
use ixc_message_api::AccountID;
use ixc_message_api::handler::HostBackend;
use ixc_message_api::header::MessageHeader;
use ixc_message_api::packet::MessagePacket;
use ixc_schema::codec::Codec;
use ixc_schema::mem::MemoryManager;

/// Context wraps a single message request (and possibly response as well) along with
/// the router callbacks necessary for making nested message calls.
pub struct Context<'a> {
    message_packet: &'a mut MessagePacket,
    host_callbacks: &'a dyn HostBackend,
    memory_manager: &'a MemoryManager<'a, 'a>,
}

impl<'a> Context<'a> {
    /// This is the address of the account that is getting called.
    /// In a receiving account, this is the account's own address.
    pub fn account_id(&self) -> &AccountID {
        unimplemented!()
    }

    /// This is the address of the account which is making the message call.
    pub fn caller(&self) -> &AccountID {
        unimplemented!()
    }

    // /// Returns a new response with the given value.
    // pub fn ok<R: ResponseValue, E: ResponseValue>(&self, res: <R as ResponseValue>::MaybeBorrowed<'a>) -> Response<'a, R, E> {
    //     Ok(res.to_owned())
    // }

    /// Dynamically invokes a module message.
    /// Static module client instances should be preferred wherever possible,
    /// so that static dependency analysis can be performed.
    pub fn dynamic_invoke_module<'b, M: Message<'b, true>>(&self, message: M) -> Response<M::Response, M::Error>
    {
        unimplemented!()
    }

    /// Dynamically invokes an account message.
    /// Static account client instances should be preferred wherever possible,
    /// so that static dependency analysis can be performed.
    pub fn dynamic_invoke_account<'b, M: Message<'b, false>>(&self, account: &AccountID, message: M) -> Response<M::Response, M::Error> {
        // TODO allocate packet
        let mut guard = self.memory_manager.scope().scope_guard();
        let new_scope = guard.scope();
        let mut header = new_scope.alloc_default::<MessageHeader>();
        // let mut packet = unsafe { MessagePacket::new(header.as_mut_ptr(), 0) };
        // let new_mem_mgr = MemoryManager::new(&new_scope);
        // let msg_body = M::Codec::encode_value(&message, new_mem_mgr.scope())?;
        // packet.in1().set_slice(msg_body);
        // self.host_callbacks.invoke(&mut packet);
        todo!()
        // TODO call self.host_callbacks.invoke
        // let code = self.host_callbacks.invoke(&mut packet);
        // if code != Code::Ok {
        //    // TODO decode error
        // } else {
        //    // TODO decode response
        // }
        unimplemented!()
    }

    /// Get the address of the module implementing the given trait, client type or module message, if any.
    pub fn get_module_address<T: ModuleAPI>(&self) -> Response<Address> {
        unimplemented!()
    }

    /// Create a new account with the given initialization data.
    pub fn new_account<H: AccountHandler>(&mut self, init: H::Init) -> Result<<<H as AccountAPI>::ClientFactory as AccountClientFactory>::Client, ()> {
        unimplemented!()
    }

    /// Create a temporary account with the given initialization data.
    /// Its address will be empty from the perspective of all observers,
    /// and it will not be persisted.
    pub fn new_temp_account<H: AccountHandler>(&mut self, init: H::Init) -> Result<<<H as AccountAPI>::ClientFactory as AccountClientFactory>::Client, ()> {
        unimplemented!()
    }

    /// Returns a deterministic ID unique to the message call which this context pertains to.
    /// Such IDs can be used to generate unique IDs for state objects.
    /// The index parameter can be used to generate up to 256 such unique IDs per message call.
    pub fn unique_id(&mut self, index: u8) -> Result<u128, ()> {
        unimplemented!()
    }
}

