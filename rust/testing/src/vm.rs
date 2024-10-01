use ixc_message_api::code::{ErrorCode, SystemErrorCode};
use ixc_message_api::handler::{HostBackend, RawHandler, Allocator};
use ixc_message_api::packet::MessagePacket;
use ixc_vm_api::{HandlerDescriptor, VM};
use std::collections::HashMap;

struct NativeVM {
    handlers: HashMap<String, Box<dyn RawHandler>>,
}

impl NativeVM {
    fn register_handler<H: RawHandler + 'static>(&mut self, name: &str, handler: H) {
        self.handlers.insert(name.to_string(), Box::new(handler));
    }
}

impl VM for NativeVM {
    fn describe_handler(&self, vm_handler_id: &str) -> Option<HandlerDescriptor> {
        if self.handlers.contains_key(vm_handler_id) {
            Some(HandlerDescriptor::default())
        } else {
            None
        }
    }

    fn run_handler(&self, vm_handler_id: &str, message_packet: &mut MessagePacket, callbacks: &dyn HostBackend, allocator: &dyn Allocator) -> Result<(), ErrorCode> {
        if let Some(handler) = self.handlers.get(vm_handler_id) {
            handler.handle(message_packet, callbacks, allocator)
                .map_err(|e| todo!())
        } else {
            Err(ErrorCode::RuntimeSystemError(SystemErrorCode::HandlerNotFound))
        }
    }
}