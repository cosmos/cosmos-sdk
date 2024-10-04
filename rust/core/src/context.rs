use core::cell::Cell;
use ixc_message_api::handler::HostBackend;
use ixc_message_api::AccountID;
use ixc_message_api::header::ContextInfo;
use ixc_schema::mem::MemoryManager;

/// Context wraps a single message request (and possibly response as well) along with
/// the router callbacks necessary for making nested message calls.
pub struct Context<'a> {
    pub(crate) mem: MemoryManager,
    pub(crate) backend: &'a dyn HostBackend,
    pub(crate) context_info: ContextInfo,
    gas_consumed: Cell<u64>,
}

impl<'a> Context<'a> {
    /// Create a new context from a message packet and host callbacks.
    pub fn new(context_info: ContextInfo, host_callbacks: &'a dyn HostBackend) -> Self {
        Self {
            mem: MemoryManager::new(),
            backend: host_callbacks,
            context_info,
            gas_consumed: Cell::new(0),
        }
    }

    /// This is the address of the account that is getting called.
    /// In a receiving account, this is the account's own address.
    pub fn account_id(&self) -> AccountID {
        self.context_info.account
    }

    /// This is the address of the account which is making the message call.
    pub fn caller(&self) -> AccountID {
        self.context_info.caller
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

