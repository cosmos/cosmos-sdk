use core::cell::Cell;
use ixc_message_api::handler::HostBackend;
use ixc_message_api::AccountID;
use ixc_message_api::header::ContextInfo;
use ixc_schema::mem::MemoryManager;

/// Context wraps a single message request (and possibly response as well) along with
/// the router callbacks necessary for making nested message calls.
pub struct Context<'a> {
    pub(crate) mem: MemHandle<'a>,
    pub(crate) backend: &'a dyn HostBackend,
    pub(crate) context_info: ContextInfo,
    gas_consumed: Cell<u64>,
}

enum MemHandle<'a> {
    Borrowed(&'a MemoryManager),
    Owned(MemoryManager),
}

impl<'a> Context<'a> {
    /// Create a new context from a message packet and host callbacks.
    pub fn new(context_info: ContextInfo, host_callbacks: &'a dyn HostBackend) -> Self {
        Self {
            mem: MemHandle::Owned(MemoryManager::new()),
            backend: host_callbacks,
            context_info,
            gas_consumed: Cell::new(0),
        }
    }

    /// Create a new context from a message packet and host callbacks with a pre-allocated memory manager.
    pub fn new_with_mem(context_info: ContextInfo, host_callbacks: &'a dyn HostBackend, mem: &'a MemoryManager) -> Self {
        Self {
            mem: MemHandle::Borrowed(mem),
            backend: host_callbacks,
            context_info,
            gas_consumed: Cell::new(0),
        }
    }

    /// This is the address of the account that is getting called.
    /// In a receiving account, this is the account's own address.
    pub fn self_account_id(&self) -> AccountID {
        self.context_info.account
    }

    /// This is the address of the account which is making the message call.
    pub fn caller(&self) -> AccountID {
        self.context_info.caller
    }

    /// Get the host backend.
    pub unsafe fn host_backend(&self) -> &dyn HostBackend {
        self.backend
    }

    /// Get the memory manager.
    pub fn memory_manager(&self) -> &MemoryManager {
        &self.mem.get()
    }
}

impl <'a> MemHandle<'a> {
    pub fn get(&self) -> &MemoryManager {
        match self {
            MemHandle::Borrowed(mem) => mem,
            MemHandle::Owned(mem) => mem,
        }
    }
}