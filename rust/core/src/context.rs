use core::mem::transmute;
use allocator_api2::alloc::Allocator;
use crate::error::Error;
use crate::message::Message;
use ixc_message_api::handler::{HandlerErrorCode, HostBackend};
use ixc_message_api::header::{MessageHeader, MessageSelector, MESSAGE_HEADER_SIZE};
use ixc_message_api::packet::MessagePacket;
use ixc_message_api::AccountID;
use ixc_schema::codec::Codec;
use ixc_schema::mem::MemoryManager;
use ixc_schema::value::OptionalValue;

/// Context wraps a single message request (and possibly response as well) along with
/// the router callbacks necessary for making nested message calls.
pub struct Context<'a> {
    mem: MemoryManager,
    message_packet: &'a MessagePacket,
    backend: &'a dyn HostBackend,
}

impl<'a> Context<'a> {
    /// Create a new context from a message packet and host callbacks.
    pub fn new(message_packet: &'a MessagePacket, host_callbacks: &'a dyn HostBackend) -> Self {
        Self {
            mem: MemoryManager::new(),
            message_packet,
            backend: host_callbacks,
        }
    }

    /// This is the address of the account that is getting called.
    /// In a receiving account, this is the account's own address.
    pub fn account_id(&self) -> AccountID {
        self.message_packet.header().account
    }

    /// This is the address of the account which is making the message call.
    pub fn caller(&self) -> AccountID {
        self.message_packet.header().sender_account
    }

    // /// Returns a new response with the given value.
    // pub fn ok<R: ResponseValue, E: ResponseValue>(&self, res: <R as ResponseValue>::MaybeBorrowed<'a>) -> Response<'a, R, E> {
    //     Ok(res.to_owned())
    // }

    // /// Dynamically invokes a module message.
    // /// Static module client instances should be preferred wherever possible,
    // /// so that static dependency analysis can be performed.
    // pub fn dynamic_invoke_module<'b, M: Message<'b, true>>(&self, message: M) -> Response<M::Response, M::Error>
    // {
    //     unimplemented!()
    // }

    /// Dynamically invokes an account message.
    /// Static account client instances should be preferred wherever possible,
    /// so that static dependency analysis can be performed.
    pub unsafe fn dynamic_invoke<'b, M: Message<'b>>(&'a mut self, account: AccountID, message: M)
        -> crate::Result<<M::Response<'a> as OptionalValue<'a>>::Value, M::Error> {
        // encode the message body
        let msg_body = M::Codec::encode_value(&message, &self.mem as &dyn Allocator).
            map_err(|_| Error::KnownHandlerError(HandlerErrorCode::EncodingError))?;

        // create the message packet and fill in call details
        let packet = self.mem.allocate_packet(0)
            .map_err(|_| Error::KnownHandlerError(HandlerErrorCode::EncodingError))?;
        let header = packet.header_mut();
        header.sender_account = self.message_packet.header().account;
        header.account = account;
        header.in_pointer1.set_slice(msg_body);
        header.message_selector = M::SELECTOR;

        // invoke the message
        let res = self.backend.invoke(packet, &self.mem)
            .map_err(|_| todo!());

        match res {
            Ok(_) => {
                let res = M::Response::<'a>::decode_value::<M::Codec>(packet, &self.mem).
                    map_err(|_| Error::KnownHandlerError(HandlerErrorCode::EncodingError))?;
                Ok(res)
            }
            Err(_) => {
                todo!()
            }
        }
    }

    // /// Dynamically invokes a message that does not modify state.
    // pub fn dynamic_invoke_readonly<'b, M: Message<'b>>(&self, account: &AccountID, message: M) -> Response<M::Response, M::Error> {
    //     todo!()
    // }
    //
    // /// Dynamically invokes a message that does not read or write state.
    // pub fn dynamic_invoke_pure<'b, M: Message<'b, false>>(&self, account: &AccountID, message: M) -> Response<M::Response, M::Error> {
    //     todo!()
    // }

    /// Get the host backend.
    pub unsafe fn host_backend(&self) -> &dyn HostBackend {
        self.backend
    }

    /// Get the memory manager.
    pub fn memory_manager(&self) -> &MemoryManager {
        &self.mem
    }

    // /// Get the address of the module implementing the given trait, client type or module message, if any.
    // pub fn get_module_address<T: ModuleAPI>(&self) -> Response<Address> {
    //     unimplemented!()
    // }
    //
    // /// Create a new account with the given initialization data.
    // pub fn new_account<H: AccountHandler>(&mut self, init: H::Init) -> Result<<<H as AccountAPI>::ClientFactory as AccountClientFactory>::Client, ()> {
    //     unimplemented!()
    // }
    //
    // /// Create a temporary account with the given initialization data.
    // /// Its address will be empty from the perspective of all observers,
    // /// and it will not be persisted.
    // pub fn new_temp_account<H: AccountHandler>(&mut self, init: H::Init) -> Result<<<H as AccountAPI>::ClientFactory as AccountClientFactory>::Client, ()> {
    //     unimplemented!()
    // }
    //
    // /// Returns a deterministic ID unique to the message call which this context pertains to.
    // /// Such IDs can be used to generate unique IDs for state objects.
    // /// The index parameter can be used to generate up to 256 such unique IDs per message call.
    // pub fn unique_id(&mut self, index: u8) -> Result<u128, ()> {
    //     unimplemented!()
    // }
}

