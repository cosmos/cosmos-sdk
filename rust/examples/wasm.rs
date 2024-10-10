use std::ptr::NonNull;
use allocator_api2::alloc::Allocator;
use ixc_message_api::code::{ErrorCode, SystemCode};
use ixc_message_api::handler::HostBackend;
use ixc_message_api::packet::MessagePacket;
use ixc_vm_api::{HandlerDescriptor, VM};
use wasmtime::*;
use ixc_message_api::header::MessageHeader;

pub struct SimpleWasmtimeVM {
    path: std::path::Path
}

impl VM for SimpleWasmtimeVM {
    fn describe_handler(&self, vm_handler_id: &str) -> Option<HandlerDescriptor> {
        if let Some(_) = self.parse_handler(vm_handler_id) {
            Some(HandlerDescriptor {})
        } else {
            None
        }
    }

    fn run_handler(&self, vm_handler_id: &str, message_packet: &mut MessagePacket, callbacks: &dyn HostBackend, allocator: &dyn Allocator) -> Result<(), ErrorCode> {
        let (path, handler_name) = match self.parse_handler(vm_handler_id) {
            Some((path, handler_name)) => (path, handler_name),
            None => return Err(ErrorCode::SystemCode(SystemCode::HandlerNotFound)),
        };
        let engine = Engine::default();
        let module = Module::from_file(&engine, path).unwrap();
        let mut linker = Linker::new(&engine);
        linker.func_wrap("ixc", "invoke", |packet: *const u8, len: usize| -> u32 unsafe {
            // TODO map guest inputs to host memory
            let mut message_packet = MessagePacket::new(NonNull::new(packet as *mut MessageHeader).unwrap(), len);
            // maybe we should use system allocator
            let res = callbacks.invoke(&mut message_packet, allocator);
            // TODO map host outputs to guest memory
            match res {
                Ok(()) => 0,
                Err(code) => code.into() as u32,
            }
        });
        todo!()
    }
}

impl SimpleWasmtimeVM {
    fn parse_handler(&self, vm_handler_id: &str) -> Option<(std::path::PathBuf, String)> {
        let parts = vm_handler_id.splitn(2, ':').collect::<Vec<_>>();
        if parts.len() != 2 {
            return None;
        }
        let (filename, handler_name) = (parts[0], parts[1]);
        let path = self.path.join(filename);
        // check if file exists
        if !path.exists() {
            return None;
        }
        Some((path, handler_name.to_string()))
    }
}

fn main() {}