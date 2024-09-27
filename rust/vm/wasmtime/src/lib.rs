use ixc_message_api::code::Code;
use ixc_message_api::handler::HostBackend;
use ixc_message_api::packet::MessagePacket;
use ixc_vm_api::{VMHandlerID, VM};

pub struct WasmtimeVM {
}

impl VM for WasmtimeVM {
    fn run_handler(&self, vm_handler_id: VMHandlerID, message_packet: &mut MessagePacket, callbacks: &dyn HostBackend) -> Code {
        todo!()
    }
}