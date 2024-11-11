use std::any::Any;
use ixc_message_api::code::{ErrorCode, SystemCode};
use ixc_message_api::handler::{HostBackend, RawHandler, Allocator};
use ixc_message_api::packet::MessagePacket;
use ixc_vm_api::{HandlerDescriptor, VM};
use std::collections::HashMap;
use std::sync::{Arc, RwLock};

/// Wrapper that holds both the RawHandler and Any trait objects
struct HandlerWrapper {
    raw: Box<dyn RawHandler>,
    any: Box<dyn Any>,
}

impl RawHandler for HandlerWrapper {
    fn handle(&self, message_packet: &mut MessagePacket, callbacks: &dyn HostBackend, allocator: &dyn Allocator) -> Result<(), ErrorCode> {
        self.raw.handle(message_packet, callbacks, allocator)
    }
}

#[derive(Clone)]
pub struct NativeVM(Arc<RwLock<NativeVMImpl>>);

pub(crate) struct NativeVMImpl {
    handlers: HashMap<String, HandlerWrapper>,
}

pub trait AnyHandler: RawHandler + Any + Clone {}
impl <T: RawHandler + Any + Clone> AnyHandler for T {}

impl NativeVM {
    pub fn new() -> NativeVM {
        NativeVM(Arc::new(RwLock::new(NativeVMImpl {
            handlers: HashMap::new(),
        })))
    }

    pub fn register_handler(&self, name: &str, handler: Box<dyn AnyHandler>) {
        let mut vm = self.0.write().unwrap();
        let wrapper = HandlerWrapper {
            raw: Box::new(handler.clone()),
            any: handler,
        };
        vm.handlers.insert(name.to_string(), wrapper);
    }

    pub fn lookup_handler<H: AnyHandler>(&self, name: &str) -> Option<&H> {
        let vm = self.0.read().unwrap();
        let wrapper = vm.handlers.get(name)?;
        wrapper.any.downcast_ref()
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
        } else {
            Err(ErrorCode::SystemCode(SystemCode::HandlerNotFound))
        }
    }
}
