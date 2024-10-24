//! Rust Cosmos SDK RFC 003 hypervisor implementation.
use ixc_core_macros::message_selector;
use ixc_message_api::code::{ErrorCode};
use ixc_message_api::handler::{Allocator, HostBackend, RawHandler};
use ixc_message_api::packet::MessagePacket;
use ixc_message_api::AccountID;
use ixc_vm_api::{HandlerID, VM};
use std::cell::RefCell;
use std::collections::HashMap;
use std::ops::DerefMut;
use std::sync::Arc;
use ixc_message_api::code::ErrorCode::SystemCode;
use ixc_message_api::code::SystemCode::{AccountNotFound, FatalExecutionError, HandlerNotFound, InvalidHandler, MessageNotHandled, UnauthorizedCallerAccess};

/// Rust Cosmos SDK RFC 003 hypervisor implementation.
pub struct Hypervisor<ST: StateHandler> {
    vmdata: Arc<VMData>,
    state_handler: ST,
}

impl<ST: StateHandler + Default> Default for Hypervisor<ST> {
    fn default() -> Self {
        Self::new(ST::default())
    }
}

struct VMData {
    vms: HashMap<String, Box<dyn VM>>,
    default_vm: Option<String>,
}

impl<ST: StateHandler> Hypervisor<ST> {
    /// Create a new hypervisor with the given state handler.
    pub fn new(state_handler: ST) -> Self {
        Self {
            vmdata: Arc::from(VMData {
                vms: HashMap::new(),
                default_vm: None,
            }),
            state_handler,
        }
    }

    /// This is a hack until we figure out a better way to reference handler IDs.
    pub fn set_default_vm(&mut self, name: &str) -> Result<(), ()> {
        let mut vmdata = Arc::get_mut(&mut self.vmdata).ok_or(())?;
        vmdata.default_vm = Some(name.to_string());
        Ok(())
    }

    /// Register a VM with the hypervisor.
    pub fn register_vm(&mut self, name: &str, vm: Box<dyn VM>) -> Result<(), ()> {
        let mut vmdata = Arc::get_mut(&mut self.vmdata).ok_or(())?;
        vmdata.vms.insert(name.to_string(), vm);
        Ok(())
    }

    /// Invoke a message packet.
    pub fn invoke(&mut self, message_packet: &mut MessagePacket, allocator: &dyn Allocator) -> Result<(), ErrorCode> {
        let tx = self.state_handler.new_transaction(message_packet.header().context_info.caller, true).
            map_err(|_| SystemCode(FatalExecutionError))?;
        let mut exec_context = ExecContext {
            vmdata: self.vmdata.clone(),
            tx: RefCell::new(tx),
        };
        let res = exec_context.invoke(message_packet, allocator);
        let tx = exec_context.tx.into_inner();
        if res.is_ok() {
            self.state_handler.commit(tx)
                .map_err(|_| SystemCode(FatalExecutionError))?;
        }

        res
    }
}

/// The state handler traits the hypervisor expects.
pub trait StateHandler {
    /// The transaction type.
    type Tx: Transaction;
    /// Create a new transaction.
    fn new_transaction(&self, account_id: AccountID, volatile: bool) -> Result<Self::Tx, NewTxError>;
    /// Commit a transaction.
    fn commit(&mut self, tx: Self::Tx) -> Result<(), CommitError>;
}

/// An error when creating a new transaction.
pub struct NewTxError;

/// An error when committing a transaction.
pub enum CommitError {
    /// Attempted to commit when the call stack was not empty.
    UnfinishedCallStack,
}

/// A transaction.
pub trait Transaction {
    /// Initialize the account storage and push a new frame for the newly initialized storage.
    fn init_account_storage(&mut self, account: AccountID) -> Result<(), PushFrameError>;
    /// Push a new execution frame.
    fn push_frame(&mut self, account: AccountID, volatile: bool) -> Result<(), PushFrameError>;
    /// Pop the current execution frame.
    fn pop_frame(&mut self, commit: bool) -> Result<(), PopFrameError>;
    /// Get the active account.
    fn active_account(&self) -> AccountID;
    /// Removes the data for the active account.
    fn self_destruct_account(&mut self) -> Result<(), ()>;
    /// Directly read a key from the account's KV store.
    fn raw_kv_get(&self, account_id: AccountID, key: &[u8]) -> Option<Vec<u8>>;
    /// Directly write a key to the account's raw KV store.
    fn raw_kv_set(&self, account_id: AccountID, key: &[u8], value: &[u8]);
    /// Directly delete a key from the account's raw KV store.
    fn raw_kv_delete(&self, account_id: AccountID, key: &[u8]);
    /// Handle a message packet.
    fn handle(&self, message_packet: &mut MessagePacket, allocator: &dyn Allocator) -> Result<(), ErrorCode>;
}

/// A push frame error.
#[non_exhaustive]
pub enum PushFrameError {
    /// Tried to push a volatile frame on top of a non-volatile frame.
    VolatileAccessError,
}

/// A pop frame error.
#[non_exhaustive]
pub enum PopFrameError {
    /// No frames to pop.
    NoFrames,
}

struct ExecContext<TX: Transaction> {
    vmdata: Arc<VMData>,
    tx: RefCell<TX>,
}

impl<'a, TX: Transaction> ExecContext<TX> {
    fn get_account_handler_id(&self, account_id: AccountID) -> Option<HandlerID> {
        let id: u64 = account_id.into();
        let key = format!("h:{}", id);
        let value = self.tx.borrow().raw_kv_get(HYPERVISOR_ACCOUNT, key.as_bytes())?;
        self.parse_handler_id(&value)
    }

    fn init_next_account(&self, handler_id: &HandlerID) -> Result<AccountID, PushFrameError> {
        let mut tx = self.tx.borrow_mut();
        let id = tx.raw_kv_get(HYPERVISOR_ACCOUNT, b"next_account_id").map_or(ACCOUNT_ID_NON_RESERVED_START, |v| {
            u64::from_le_bytes(v.try_into().unwrap())
        });
        // we push a new storage frame here because if initialization fails all of this gets rolled back
        tx.init_account_storage(AccountID::new(id))?;
        tx.raw_kv_set(HYPERVISOR_ACCOUNT, b"next_account_id", &(id + 1).to_le_bytes());
        tx.raw_kv_set(HYPERVISOR_ACCOUNT, format!("h:{}", id).as_bytes(), format_handler_id(handler_id).as_bytes());
        Ok(AccountID::new(id))
    }

    fn destroy_current_account_data(&self) -> Result<(), ()> {
        let current_account = self.tx.borrow().active_account();
        let id: u64 = current_account.into();
        let key = format!("h:{}", id);
        self.tx.borrow().raw_kv_delete(HYPERVISOR_ACCOUNT, key.as_bytes());
        self.tx.borrow_mut().self_destruct_account()
    }

    fn parse_handler_id(&self, value: &[u8]) -> Option<HandlerID> {
        parse_handler_id(value, &self.vmdata.default_vm)
    }
}

const ACCOUNT_ID_NON_RESERVED_START: u64 = u16::MAX as u64 + 1;

const HYPERVISOR_ACCOUNT: AccountID = AccountID::new(1);
const STATE_ACCOUNT: AccountID = AccountID::new(2);

impl<TX: Transaction> HostBackend for ExecContext<TX> {
    fn invoke(&self, message_packet: &mut MessagePacket, allocator: &dyn Allocator) -> Result<(), ErrorCode> {
        // get the mutable transaction from the RefCell
        // check if the caller matches the active account
        let account = self.tx.borrow().active_account();
        if message_packet.header().context_info.caller != account {
            return Err(SystemCode(UnauthorizedCallerAccess));
        }
        // TODO support authorization middleware

        let target_account = message_packet.header().context_info.account;
        // check if the target account is a system account
        match target_account {
            HYPERVISOR_ACCOUNT => return self.handle_system_message(message_packet, allocator),
            STATE_ACCOUNT => return self.tx.borrow().handle(message_packet, allocator),
            _ => {}
        }

        // find the account's handler ID and retrieve its VM
        let handler_id = self.get_account_handler_id(target_account).
            ok_or(SystemCode(AccountNotFound))?;
        let vm = self.vmdata.vms.get(&handler_id.vm).
            ok_or(SystemCode(HandlerNotFound))?;

        // push an execution frame for the target account
        self.tx.borrow_mut().push_frame(target_account, false). // TODO add volatility support
            map_err(|_| SystemCode(InvalidHandler))?;
        // run the handler
        let res = vm.run_handler(&handler_id.vm_handler_id, message_packet, self, allocator);
        // pop the execution frame
        self.tx.borrow_mut().pop_frame(res.is_ok()).
            map_err(|_| SystemCode(InvalidHandler))?;

        res
    }
}

impl<TX: Transaction> ExecContext<TX> {
    fn handle_system_message(&self, message_packet: &mut MessagePacket, allocator: &dyn Allocator) -> Result<(), ErrorCode> {
        match message_packet.header().message_selector {
            CREATE_SELECTOR => unsafe {
                // get the input data
                let create_header = message_packet.header_mut();
                let handler_id = create_header.in_pointer1.get(message_packet);
                let init_data = create_header.in_pointer2.get(message_packet);

                // resolve the handler ID and retrieve the VM
                let handler_id = self.parse_handler_id(handler_id).
                    ok_or(SystemCode(HandlerNotFound))?;
                let vm = self.vmdata.vms.get(&handler_id.vm).
                    ok_or(SystemCode(HandlerNotFound))?;
                let desc = vm.describe_handler(&handler_id.vm_handler_id).
                    ok_or(SystemCode(HandlerNotFound))?;

                // get the next account ID and initialize the account storage
                let id = self.init_next_account(&handler_id).
                    map_err(|_| SystemCode(InvalidHandler))?;

                // create a packet for calling on_create
                let mut on_create_packet = MessagePacket::allocate(allocator, 0).
                    map_err(|_| SystemCode(FatalExecutionError))?;
                let mut on_create_header = on_create_packet.header_mut();
                // TODO: how do we specify a selector that can only be called by the system?
                on_create_header.context_info.account = id;
                on_create_header.context_info.caller = create_header.context_info.caller;
                on_create_header.message_selector = ON_CREATE_SELECTOR;
                on_create_header.in_pointer1.set_slice(init_data);

                let res = vm.run_handler(&handler_id.vm_handler_id, &mut on_create_packet, self, allocator);
                let is_ok = match res {
                    Ok(_) => true,
                    Err(SystemCode(ixc_message_api::code::SystemCode::MessageNotHandled)) => true,
                    _ => false,
                };
                self.tx.borrow_mut().pop_frame(is_ok).
                    map_err(|_| SystemCode(FatalExecutionError))?;

                if is_ok {
                    create_header.in_pointer1.set_u64(id.into());
                    Ok(())
                } else {
                    res
                }
            },
            SELF_DESTRUCT_SELECTOR => {
                self.destroy_current_account_data().
                    map_err(|_| SystemCode(FatalExecutionError))
            }
            _ => {
                Err(SystemCode(MessageNotHandled))
            }
        }
    }
}

const CREATE_SELECTOR: u64 = message_selector!("ixc.account.v1.create");
const ON_CREATE_SELECTOR: u64 = message_selector!("ixc.account.v1.on_create");
const SELF_DESTRUCT_SELECTOR: u64 = message_selector!("ixc.account.v1.self_destruct");

fn parse_handler_id(value: &[u8], default_vm: &Option<String>) -> Option<HandlerID> {
    let str = String::from_utf8(value.to_vec()).ok()?;
    let mut parts = str.split(':');
    let mut vm = parts.next()?;
    let vm_handler_id = if let Some(handler_id) = parts.next() {
        handler_id
    } else {
        let handler_id = vm;
        if let Some(dvm) = default_vm {
            vm = dvm;
        } else {
            return None;
        }
        handler_id
    };
    Some(HandlerID {
        vm: vm.to_string(),
        vm_handler_id: vm_handler_id.to_string(),
    })
}

fn format_handler_id(HandlerID { vm, vm_handler_id }: &HandlerID) -> String {
    format!("{}:{}", vm, vm_handler_id)
}

#[cfg(test)]
mod tests {
    #[test]
    fn test_parse_handler_id() {
        let value = b"vm1:handler1";
        let handler_id = super::parse_handler_id(value, &None).unwrap();
        assert_eq!(handler_id.vm, "vm1");
        assert_eq!(handler_id.vm_handler_id, "handler1");

        let value = b"handler1";
        let opt_handler_id = super::parse_handler_id(value, &None);
        assert!(opt_handler_id.is_none());

        let value = b"handler1";
        let handler_id = super::parse_handler_id(value, &Some("default".into())).unwrap();
        assert_eq!(handler_id.vm, "default");
        assert_eq!(handler_id.vm_handler_id, "handler1");
    }
}