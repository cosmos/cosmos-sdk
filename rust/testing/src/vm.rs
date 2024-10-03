use ixc_message_api::code::{ErrorCode, SystemErrorCode};
use ixc_message_api::handler::{HostBackend, RawHandler, Allocator, HandlerError};
use ixc_message_api::packet::MessagePacket;
use ixc_vm_api::{HandlerDescriptor, VM};
use std::collections::HashMap;
use std::sync::{Arc, RwLock};

#[derive(Clone)]
pub struct NativeVM(Arc<RwLock<NativeVMImpl>>);

pub(crate) struct NativeVMImpl {
    handlers: HashMap<String, Box<dyn RawHandler>>,
}

impl NativeVM {
    pub fn new() -> NativeVM {
        NativeVM(Arc::new(RwLock::new(NativeVMImpl {
            handlers: HashMap::new(),
        })))
    }

    pub fn register_handler<H: RawHandler + 'static>(&self, name: &str, handler: H) {
        let mut vm = self.0.write().unwrap();
        vm.handlers.insert(name.to_string(), Box::new(handler));
    }
}

impl VM for NativeVM {
    fn describe_handler(&self, vm_handler_id: &str) -> Option<HandlerDescriptor> {
        let vm = self.0.read().unwrap();
        if vm.handlers.contains_key(vm_handler_id) {
            Some(HandlerDescriptor::default())
        } else {
            None
        }
    }

    fn run_handler(&self, vm_handler_id: &str, message_packet: &mut MessagePacket, callbacks: &dyn HostBackend, allocator: &dyn Allocator) -> Result<(), ErrorCode> {
        let vm = self.0.read().unwrap();
        if let Some(handler) = vm.handlers.get(vm_handler_id) {
            handler.handle(message_packet, callbacks, allocator)
                .map_err(|e|
                    match e {
                        HandlerError::KnownCode(code) => ErrorCode::HandlerSystemError(code),
                        HandlerError::Custom(x) => ErrorCode::CustomHandlerError(x)
                    }
                )
        } else {
            Err(ErrorCode::RuntimeSystemError(SystemErrorCode::HandlerNotFound))
        }
    }
}