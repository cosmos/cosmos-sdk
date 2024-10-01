use std::alloc::Layout;
use imbl::{HashMap, OrdMap, Vector};
use thiserror::Error;
use ixc_message_api::AccountID;
use ixc_message_api::code::ErrorCode;
use ixc_message_api::packet::MessagePacket;
use ixc_message_api::handler::{HostBackend, RawHandler};
use ixc_hypervisor::{KVStore, PopFrameError, PushFrameError, Transaction};

pub struct VersionedMultiStore {
    versions: Vector<MultiStore>,
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

impl KVStore for Store {
    fn get(&self, key: &[u8]) -> Option<Vec<u8>> {
        self.kv_store.get(key).cloned()
    }

    fn set(&mut self, key: &[u8], value: &[u8]) {
        self.kv_store.insert(key.to_vec(), value.to_vec());
    }

    fn delete(&mut self, key: &[u8]) {
        self.kv_store.remove(key);
    }
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

impl Transaction for Tx {
    type KVStore = Store;

    fn init_account_storage(&mut self, account: AccountID, storage_params: &[u8]) {
        todo!()
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

    fn rollback(self) {}

    fn manager_state(&self) -> &mut Self::KVStore {
        todo!()
    }

    fn handler(&self) -> &dyn RawHandler {
        todo!()
    }
}

enum Access {
    Read,
    Write,
}

impl Tx {
    fn pop(&mut self) -> bool {
        if let Some(mut previous_frame) = self.call_stack.pop() {
            previous_frame.store.stores.insert(self.current_frame.account, self.current_store.clone());
            previous_frame.changes.append(&mut self.current_frame.changes);
            self.current_frame = previous_frame;
            self.current_store = self.current_frame.store.stores.get(&self.current_frame.account).map(|s| s.clone()).unwrap_or_default();
            true
        } else {
            false
        }
    }

    unsafe fn kv_get(&self, packet: &mut MessagePacket, backend: &dyn HostBackend) -> Result<ErrorCode, Error> {
        // let key = packet.header().in_pointer1.get(packet);
        // self.track_access(key, Access::Read)?;
        // match self.current_store.kv_store.get(key) {
        //     None => unsafe {
        //         // TODO what should we do when not found?
        //         packet.out1().set_slice(&[]);
        //     }
        //     Some(value) => unsafe {
        //         let out = backend.alloc(Layout::from_size_align_unchecked(value.len(), 16))?;
        //         let out_slice = core::slice::from_raw_parts_mut(out, value.len());
        //         out_slice.copy_from_slice(value.as_slice());
        //         packet.out1().set_slice(out_slice);
        //     }
        // }
        // Ok(ErrorCode::Ok)
        todo!()
    }

    fn kv_set(&mut self, packet: &mut MessagePacket) -> Result<ErrorCode, Error> {
        // self.track_access(packet.in1().get(), Access::Write)?;
        // self.current_frame.changes.push(Update {
        //     account: self.current_frame.account,
        //     key: packet.in1().get().to_vec(),
        //     operation: Operation::Set(packet.in2().get().to_vec()),
        // });
        // self.current_store.kv_store.insert(packet.in1().get().to_vec(), packet.in2().get().to_vec());
        // Ok(ErrorCode::Ok)
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
    volatile: bool
    // TODO events
}