use std::alloc::Layout;
use imbl::HashMap;
use ixc::*;
use ixc_message_api::code::ErrorCode;
use ixc_message_api::handler::{AllocError, RawHandler, HostBackend};
use ixc_message_api::packet::MessagePacket;
use crate::store::VersionedMultiStore;

pub struct AccountManager {
    accounts: Map<u128, u64>,
}

pub struct Hypervisor {
    handlers: HashMap<u64, Box<dyn RawHandler>>,
    account_manager: AccountManager,
    state: VersionedMultiStore,
}

impl Hypervisor {
    fn invoke(&mut self, message_packet: &mut MessagePacket) -> ErrorCode {
        unsafe {
            let target = message_packet.header().account;
            // with account manager read ID of current handler from current state
            // get handler
            // invoke handler with message packet linked to current state
            todo!()
        }
    }
}

struct ExecContext {

}

impl HostBackend for ExecContext {
    fn invoke(&self, message_packet: &mut MessagePacket) -> ErrorCode {
        todo!()
    }

    unsafe fn alloc(&self, layout: Layout) -> Result<*mut u8, AllocError> {
        todo!()
    }
}