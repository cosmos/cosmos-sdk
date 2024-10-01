use std::collections::HashMap;
use ixc_message_api::code::{ErrorCode, SystemErrorCode};
use ixc_message_api::handler::{RawHandler, HostBackend};
use ixc_message_api::packet::MessagePacket;
use ixc_vm_api::{HandlerDescriptor, VM};

struct NativeVM {
    next_handler_id: u64,
    handlers: HashMap<String, Box<dyn RawHandler>>,
    handler_names: HashMap<u64, &'static str>,
}

impl NativeVM {
    fn register_handler<H: RawHandler>(&mut self, handler: H) {
        // let id = self.next_handler_id;
        // self.next_handler_id += 1;
        // self.handler_names.insert(id, handler.name());
        // self.handlers.insert(id, Box::new(handler));
        todo!()
    }
}

impl VM for NativeVM {
    fn describe_handler(&self, vm_handler_id: &str) -> Option<HandlerDescriptor> {
        todo!()
    }

    fn run_handler(&self, vm_handler_id: &str, message_packet: &mut MessagePacket, callbacks: &dyn HostBackend) -> Result<(), ErrorCode> {
        // if let Some(handler) = self.handlers.get(vm_handler_id) {
        //     let code = handler.handle(message_packet, callbacks)
        //         .map_err(|code| ErrorCode::HandlerSystemError(code))?;
        //     match code {
        //         HandlerCode::Ok => ErrorCode::Ok,
        //         HandlerCode::HandlerError(code) => ErrorCode::HandlerSystemError(code),
        //     }
        // } else {
        //     ErrorCode::RuntimeSystemError(SystemErrorCode::HandlerNotFound)
        // }
        todo!()
    }
}