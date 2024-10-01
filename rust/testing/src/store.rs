use allocator_api2::alloc::Allocator;
use imbl::{HashMap, OrdMap, Vector};
use ixc::message_selector;
use ixc_hypervisor::{CommitError, NewTxError, PopFrameError, PushFrameError, StateHandler, Transaction};
use ixc_message_api::code::ErrorCode;
use ixc_message_api::header::MessageSelector;
use ixc_message_api::packet::MessagePacket;
use ixc_message_api::AccountID;
use std::alloc::Layout;
use thiserror::Error;

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
            current_frame: Frame {
                store: latest,
                account: account_id,
                changes: vec![],
                volatile,
                user_tx: true,
            },
            current_store: Store::default(),
        })
    }

    fn commit(&mut self, tx: Self::Tx) -> Result<(), CommitError> {
        if !tx.call_stack.is_empty() {
            return Err(CommitError::UnfinishedCallStack);
        }
        let mut store = tx.current_frame.store;
        store.stores.insert(tx.current_frame.account, tx.current_store);
        self.versions.push_back(store);
        Ok(())
    }
}

#[derive(Default, Clone)]
pub struct MultiStore {
    stores: HashMap<AccountID, Store>,
}

#[derive(Default, Clone)]
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
    current_frame: Frame,
    current_store: Store,
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
        if !self.current_frame.volatile && volatile {
            return Err(PushFrameError::VolatileAccessError);
        }
        self.current_frame.store.stores.insert(self.current_frame.account, self.current_store.clone());
        self.current_store = self.current_frame.store.stores.get(&account).map(|s| s.clone()).unwrap_or_default();
        let next_frame = Frame {
            store: self.current_frame.store.clone(),
            account,
            changes: vec![],
            volatile,
            user_tx: false,
        };
        self.call_stack.push(self.current_frame.clone());
        self.current_frame = next_frame;
        Ok(())
    }

    fn pop_frame(&mut self, commit: bool) -> Result<(), PopFrameError> {
        if let Some(mut previous_frame) = self.call_stack.pop() {
            if commit {
                previous_frame.store.stores.insert(self.current_frame.account, self.current_store.clone());
                previous_frame.changes.append(&mut self.current_frame.changes);
            }
            self.current_frame = previous_frame;
            self.current_store = self.current_frame.store.stores.get(&self.current_frame.account).map(|s| s.clone()).unwrap_or_default();
            Ok(())
        } else {
            Err(PopFrameError::NoFrames)
        }
    }

    fn active_account(&self) -> AccountID {
        self.current_frame.account
    }

    fn raw_kv_get(&self, account_id: AccountID, key: &[u8]) -> Option<Vec<u8>> {
        if account_id == self.current_frame.account {
            self.current_store.kv_store.get(key).cloned()
        } else {
            self.current_frame.store.stores.get(&account_id).and_then(|s| s.kv_store.get(key).cloned())
        }
    }

    fn raw_kv_set(&mut self, account_id: AccountID, key: &[u8], value: &[u8]) {
        if account_id == self.current_frame.account {
            self.current_store.kv_store.insert(key.to_vec(), value.to_vec());
        } else {
            let store = self.current_frame.store.stores.get_mut(&account_id).unwrap();
            store.kv_store.insert(key.to_vec(), value.to_vec());
        }
        self.current_frame.changes.push(Update {
            account: account_id,
            key: key.to_vec(),
            operation: Operation::Set(value.to_vec()),
        });
    }

    fn raw_kv_delete(&mut self, account_id: AccountID, key: &[u8]) {
        if account_id == self.current_frame.account {
            self.current_store.kv_store.remove(key);
        } else {
            let store = self.current_frame.store.stores.get_mut(&account_id).unwrap();
            store.kv_store.remove(key);
        }
        self.current_frame.changes.push(Update {
            account: account_id,
            key: key.to_vec(),
            operation: Operation::Remove,
        });
    }

    fn handle(&mut self, message_packet: &mut MessagePacket, allocator: &dyn Allocator) -> Result<(), ErrorCode> {
        unsafe {
            let header = message_packet.header();
            match header.message_selector {
                HAS_SELECTOR => self.has(message_packet,),
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
        self.track_access(key, Access::Read)?;
        match self.current_store.kv_store.get(key) {
            None => unsafe {
                // TODO what should we do when not found?
                packet.out1().set_slice(&[]);
            }
            Some(value) => unsafe {
                let out = allocator.alloc(Layout::from_size_align_unchecked(value.len(), 16))?;
                let out_slice = core::slice::from_raw_parts_mut(out, value.len());
                out_slice.copy_from_slice(value.as_slice());
                packet.out1().set_slice(out_slice);
            }
        }
        Ok(())
    }

    unsafe fn set(&mut self, packet: &mut MessagePacket) -> Result<(), ErrorCode> {
        self.track_access(packet.in1().get(), Access::Write)?;
        self.current_frame.changes.push(Update {
            account: self.current_frame.account,
            key: packet.in1().get().to_vec(),
            operation: Operation::Set(packet.in2().get().to_vec()),
        });
        self.current_store.kv_store.insert(packet.in1().get().to_vec(), packet.in2().get().to_vec());
        Ok(())
    }

    unsafe fn delete(&mut self, packet: &mut MessagePacket) -> Result<(), ErrorCode> {
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