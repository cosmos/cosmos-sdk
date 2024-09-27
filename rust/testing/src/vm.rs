use std::collections::HashMap;
use ixc_message_api::code::{Code, SystemErrorCode};
use ixc_message_api::handler::{Handler, HandlerCode, HostBackend};
use ixc_message_api::packet::MessagePacket;
use ixc_vm_api::{VMHandlerID, VM};

struct NativeVM {
    next_handler_id: u64,
    handlers: HashMap<u64, Box<dyn Handler>>,
    handler_names: HashMap<u64, &'static str>,
}

impl NativeVM {
    fn register_handler<H: Handler>(&mut self, handler: H) {
        let id = self.next_handler_id;
        self.next_handler_id += 1;
        self.handler_names.insert(id, handler.name());
        self.handlers.insert(id, Box::new(handler));
    }
}

impl VM for NativeVM {
    fn run_handler(&self, vm_handler_id: VMHandlerID, message_packet: &mut MessagePacket, callbacks: &dyn HostBackend) -> Code {
        if let Some(handler) = self.handlers.get(&vm_handler_id.package) {
            let code = handler.handle(message_packet, callbacks);
            match code {
                HandlerCode::Ok => Code::Ok,
                HandlerCode::HandlerError(code) => Code::HandlerError(code),
            }
        } else {
            Code::SystemError(SystemErrorCode::HandlerNotFound)
        }
    }
}