use allocator_api2::alloc::Allocator;
use imbl::{HashMap, OrdMap, Vector};
use ixc::message_selector;
use ixc_hypervisor::{CommitError, NewTxError, PopFrameError, PushFrameError, StateHandler, Transaction};
use ixc_message_api::header::MessageSelector;
use ixc_message_api::packet::MessagePacket;
use ixc_message_api::AccountID;
use std::alloc::Layout;
use std::cell::RefCell;
use thiserror::Error;
use ixc_message_api::code::ErrorCode;
use ixc_message_api::code::ErrorCode::{HandlerCode, SystemCode};
use ixc_message_api::code::SystemCode::{FatalExecutionError, InvalidHandler};

#[derive(Default, Clone)]
pub struct VersionedMultiStore {
    versions: Vector<MultiStore>,
}

impl StateHandler for VersionedMultiStore {
    type Tx = Tx;

    fn new_transaction(&self, account_id: AccountID, volatile: bool) -> Result<Self::Tx, NewTxError> {
        let latest = self.versions.last().map(|s| s.clone()).unwrap_or_default();
        Ok(Tx {
            call_stack: vec![],
            current_frame: RefCell::new(Frame {
                store: latest,
                account: account_id,
                changes: vec![],
                volatile,
                user_tx: true,
            }),
        })
    }

    fn commit(&mut self, tx: Self::Tx) -> Result<(), CommitError> {
        if !tx.call_stack.is_empty() {
            return Err(CommitError::UnfinishedCallStack);
        }
        let current_frame = tx.current_frame.into_inner();
        self.versions.push_back(current_frame.store);
        Ok(())
    }
}

#[derive(Default, Clone, Debug)]
pub struct MultiStore {
    stores: HashMap<AccountID, Store>,
}

#[derive(Default, Clone, Debug)]
pub struct Store {
    kv_store: OrdMap<Vec<u8>, Vec<u8>>,
    accumulator_store: OrdMap<Vec<u8>, u128>,
}

#[derive(Clone)]
struct Update {
    account: AccountID,
    key: Vec<u8>,
    operation: Operation,
}

#[derive(Clone)]
enum Operation {
    Set(Vec<u8>),
    Remove,
    Add(u128),
    SafeSub(u128),
}

type ChangeSet = Vec<Update>;

pub struct Tx {
    call_stack: Vec<Frame>,
    current_frame: RefCell<Frame>,
}

const HAS_SELECTOR: MessageSelector = message_selector!("ixc.store.v1.has");
const GET_SELECTOR: MessageSelector = message_selector!("ixc.store.v1.get");
const SET_SELECTOR: MessageSelector = message_selector!("ixc.store.v1.set");
const DELETE_SELECTOR: MessageSelector = message_selector!("ixc.store.v1.delete");

impl Transaction for Tx {
    fn init_account_storage(&mut self, account: AccountID, _storage_params: &[u8]) -> Result<(), PushFrameError> {
        self.push_frame(account, true)
    }

    fn push_frame(&mut self, account: AccountID, volatile: bool) -> Result<(), PushFrameError> {
        if !self.current_frame.borrow().volatile && volatile {
            return Err(PushFrameError::VolatileAccessError);
        }
        let next_frame = Frame {
            store: self.current_frame.borrow().store.clone(),
            account,
            changes: vec![],
            volatile,
            user_tx: false,
        };
        self.call_stack.push(self.current_frame.borrow().clone());
        self.current_frame = RefCell::new(next_frame);
        Ok(())
    }

    fn pop_frame(&mut self, commit: bool) -> Result<(), PopFrameError> {
        if let Some(mut previous_frame) = self.call_stack.pop() {
            if commit {
                let current_frame = self.current_frame.borrow();
                previous_frame.store = current_frame.store.clone();
                previous_frame.changes.append(&mut current_frame.changes.clone());
            }
            self.current_frame = RefCell::new(previous_frame);
            Ok(())
        } else {
            Err(PopFrameError::NoFrames)
        }
    }

    fn active_account(&self) -> AccountID {
        self.current_frame.borrow().account
    }

    fn self_destruct_account(&mut self) -> Result<(), ()> {
        let mut current_frame = self.current_frame.borrow_mut();
        let account = current_frame.account;
        current_frame.store.stores.remove(&account);
        Ok(())
    }

    fn raw_kv_get(&self, account_id: AccountID, key: &[u8]) -> Option<Vec<u8>> {
        let current_frame = self.current_frame.borrow();
        current_frame.store.stores.get(&account_id).and_then(|s| s.kv_store.get(key).cloned())
    }

    fn raw_kv_set(&self, account_id: AccountID, key: &[u8], value: &[u8]) {
        let mut current_frame = self.current_frame.borrow_mut();
        let mut store = current_frame.get_kv_store(account_id);
        store.kv_store.insert(key.to_vec(), value.to_vec());
        current_frame.changes.push(Update {
            account: account_id,
            key: key.to_vec(),
            operation: Operation::Set(value.to_vec()),
        });
    }

    fn raw_kv_delete(&self, account_id: AccountID, key: &[u8]) {
        let mut current_frame = self.current_frame.borrow_mut();
        let mut store = current_frame.get_kv_store(account_id);
        store.kv_store.remove(key);
        current_frame.changes.push(Update {
            account: account_id,
            key: key.to_vec(),
            operation: Operation::Remove,
        });
    }

    fn handle(&self, message_packet: &mut MessagePacket, allocator: &dyn Allocator) -> Result<(), ErrorCode> {
        unsafe {
            let header = message_packet.header();
            match header.message_selector {
                HAS_SELECTOR => self.has(message_packet),
                GET_SELECTOR => self.get(message_packet, allocator),
                SET_SELECTOR => self.set(message_packet),
                DELETE_SELECTOR => self.delete(message_packet),
                _ => Err(todo!())
            }
        }
    }
}

enum Access {
    Read,
    Write,
}

impl Tx {
    unsafe fn has(&self, packet: &mut MessagePacket) -> Result<(), ErrorCode> {
        let key = packet.header().in_pointer1.get(packet);
        // self.track_access(key, Access::Read)?;
        todo!()
    }

    unsafe fn get(&self, packet: &mut MessagePacket, allocator: &dyn Allocator) -> Result<(), ErrorCode> {
        let key = packet.header().in_pointer1.get(packet);
        self.track_access(key, Access::Read)
            .map_err(|_| SystemCode(InvalidHandler))?;
        let mut current_frame = self.current_frame.borrow_mut();
        let account = current_frame.account;
        let current_store = current_frame.get_kv_store(account);
        match current_store.kv_store.get(key) {
            None => unsafe {
                return Err(HandlerCode(0)) // KV-stores should use handler code 0 to indicate not found
            }
            Some(value) => unsafe {
                let out = allocator.allocate(Layout::from_size_align_unchecked(value.len(), 16)).
                    map_err(|_| SystemCode(FatalExecutionError))?;
                let out_slice = core::slice::from_raw_parts_mut(out.as_ptr() as *mut u8, value.len());
                out_slice.copy_from_slice(value.as_slice());
                packet.header_mut().out_pointer1.set_slice(out_slice);
            }
        }
        Ok(())
    }

    unsafe fn set(&self, packet: &mut MessagePacket) -> Result<(), ErrorCode> {
        let key = packet.header().in_pointer1.get(packet);
        let value = packet.header().in_pointer2.get(packet);
        self.track_access(key, Access::Write)
            .map_err(|_| SystemCode(InvalidHandler))?;
        let mut current_frame = self.current_frame.borrow_mut();
        let account = current_frame.account;
        let mut current_store = current_frame.get_kv_store(account);
        current_store.kv_store.insert(key.to_vec(), value.to_vec());
        current_frame.changes.push(Update {
            account,
            key: key.to_vec(),
            operation: Operation::Set(value.to_vec()),
        });
        Ok(())
    }

    unsafe fn delete(&self, packet: &mut MessagePacket) -> Result<(), ErrorCode> {
        todo!()
    }

    fn track_access(&self, key: &[u8], access: Access) -> Result<(), AccessError> {
        // TODO track reads and writes for parallel execution
        Ok(())
    }
}

#[derive(Debug, Error)]
enum Error {
    #[error("allocation error")]
    AllocError(#[from] allocator_api2::alloc::AllocError),
    #[error("access error")]
    AccessError(#[from] AccessError),
}

#[derive(Debug, Error)]
#[error("access error")]
struct AccessError;

#[derive(Clone)]
pub struct Frame {
    store: MultiStore,
    account: AccountID,
    changes: ChangeSet,
    volatile: bool,
    user_tx: bool,
    // TODO events
}

impl Frame {
    fn get_kv_store(&mut self, account_id: AccountID) -> &mut Store {
        if self.store.stores.contains_key(&account_id) {
            self.store.stores.get_mut(&account_id).unwrap()
        } else {
            self.store.stores.insert(account_id, Store::default());
            self.store.stores.get_mut(&account_id).unwrap()
        }
    }
}