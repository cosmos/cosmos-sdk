use core::cell::Cell;
use ixc_message_api::handler::HostBackend;
use ixc_message_api::AccountID;
use ixc_schema::mem::MemoryManager;

/// Context wraps a single message request (and possibly response as well) along with
/// the router callbacks necessary for making nested message calls.
pub struct Context<'a> {
    pub(crate) mem: MemHandle<'a>,
    pub(crate) backend: &'a dyn HostBackend,
    pub(crate) account: AccountID, // 16 bytes
    pub(crate) caller: AccountID, // 16 bytes
    gas_left: Cell<u64>,
}

enum MemHandle<'a> {
    Borrowed(&'a MemoryManager),
    Owned(MemoryManager),
}

impl<'a> Context<'a> {
    /// Create a new context from a message packet and host callbacks.
    pub fn new(account: AccountID, caller: AccountID, gas_left: u64, host_callbacks: &'a dyn HostBackend) -> Self {
        Self {
            mem: MemHandle::Owned(MemoryManager::new()),
            backend: host_callbacks,
            account,
            caller,
            gas_left: Cell::new(gas_left),
        }
    }

    /// Create a new context from a message packet and host callbacks with a pre-allocated memory manager.
    pub fn new_with_mem(account: AccountID, caller: AccountID, gas_left: u64, host_callbacks: &'a dyn HostBackend, mem: &'a MemoryManager) -> Self {
        Self {
            mem: MemHandle::Borrowed(mem),
            backend: host_callbacks,
            account,
            caller,
            gas_left: Cell::new(gas_left),
        }
    }

    /// This is the address of the account that is getting called.
    /// In a receiving account, this is the account's own address.
    pub fn self_account_id(&self) -> AccountID {
        self.account
    }

    /// This is the address of the account which is making the message call.
    pub fn caller(&self) -> AccountID {
        self.caller
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