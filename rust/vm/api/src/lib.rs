//! **WARNING: This is an API preview! Most code won't work or even type check properly!**
//!
//! Virtual Machine API
use ixc_message_api::code::Code;
use ixc_message_api::handler::{HostBackend};
use ixc_message_api::packet::MessagePacket;

/// A unique identifier for a handler implementation.
pub struct HandlerID {
    /// The unique identifier for the virtual machine that the handler is implemented in.
    pub vm: u32,
    /// The unique identifier for the handler within the virtual machine.
    pub vm_handler_id: VMHandlerID,
}

/// A unique identifier for a handler implementation within a virtual machine.
pub struct VMHandlerID {
    /// The unique package within the virtual machine that the handler is implemented in.
    pub package: u64,
    /// The unique identifier for the handler within the package.
    pub handler: u32,
}

/// A virtual machine that can run message handlers.
pub trait VM {
    /// Run a message handler within the virtual machine.
    fn run_handler(&self, vm_handler_id: VMHandlerID, message_packet: &mut MessagePacket, callbacks: &dyn HostBackend) -> Code;
}
