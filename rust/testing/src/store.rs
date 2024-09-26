use std::alloc::Layout;
use imbl::{HashMap, OrdMap, Vector};
use thiserror::Error;
use ixc_message_api::AccountID;
use ixc_message_api::code::Code;
use ixc_message_api::packet::MessagePacket;
use ixc_message_api::handler::{AllocError, HostBackend};

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

enum Access {
    Read,
    Write,
}

impl Tx {
    fn push(&mut self, account: AccountID) {
        self.current_frame.store.stores.insert(self.current_frame.account, self.current_store.clone());
        self.current_store = self.current_frame.store.stores.get(&account).unwrap_or_default().clone();
        let next_frame = Frame {
            store: self.current_frame.store.clone(),
            account,
            changes: vec![],
        };
        self.call_stack.push(self.current_frame.clone());
        self.current_frame = next_frame;
    }

    fn pop(&mut self) -> bool {
        if let Some(mut previous_frame) = self.call_stack.pop() {
            previous_frame.store.stores.insert(self.current_frame.account, self.current_store.clone());
            previous_frame.changes.append(&mut self.current_frame.changes);
            self.current_frame = previous_frame;
            self.current_store = self.current_frame.store.stores.get(&self.current_frame.account).unwrap_or_default().clone();
            true
        } else {
            false
        }
    }

    fn kv_get(&self, packet: &mut MessagePacket, backend: &dyn HostBackend) -> Result<Code, Error> {
        self.track_access(packet.in1().get(), Access::Read)?;
        match self.current_store.kv_store.get(&packet.in1().get()) {
            None => unsafe {
                // TODO what should we do when not found?
                packet.out1().set_slice(&[]);
            }
            Some(value) => unsafe {
                let out = backend.alloc(Layout::from_size_align_unchecked(value.len(), 16))?;
                let out_slice = core::slice::from_raw_parts_mut(out, value.len());
                out_slice.copy_from_slice(value.as_slice());
                packet.out1().set_slice(out_slice);
            }
        }
        Ok(Code::Ok)
    }

    fn kv_set(&mut self, packet: &mut MessagePacket) -> Result<Code, Error> {
        self.track_access(packet.in1().get(), Access::Write)?;
        self.current_frame.changes.push(Update {
            account: self.current_frame.account,
            key: packet.in1().get().to_vec(),
            operation: Operation::Set(packet.in2().get().to_vec()),
        });
        self.current_store.kv_store.insert(packet.in1().get().to_vec(), packet.in2().get().to_vec());
        Ok(Code::Ok)
    }

    fn track_access(&self, key: &[u8], access: Access) -> Result<(), AccessError> {
        // TODO track reads and writes for parallel execution
        Ok(())
    }
}

#[derive(Debug, Error)]
enum Error {
    AllocError(#[from] AllocError),
    AccessError(#[from] AccessError),
}

#[derive(Debug, Error)]
struct AccessError;

#[derive(Clone)]
pub struct Frame {
    store: MultiStore,
    account: AccountID,
    changes: ChangeSet,
    // TODO events
}