//! **WARNING: This is an API preview! Expect major bugs, glaring omissions, and breaking changes!**
//!
//! Virtual Machine API
use ixc_message_api::code::ErrorCode;
use ixc_message_api::handler::{Allocator, HostBackend};
use ixc_message_api::packet::MessagePacket;

/// A unique identifier for a handler implementation.
#[derive(Debug, Clone)]
pub struct HandlerID {
    // NOTE: encoding these as strings should be considered a temporary
    /// The unique identifier for the virtual machine that the handler is implemented in.
    pub vm: String,
    /// The unique identifier for the handler within the virtual machine.
    pub vm_handler_id: String,
}

/// A virtual machine that can run message handlers.
pub trait VM {
    /// Describe a handler within the virtual machine.
    fn describe_handler(&self, vm_handler_id: &str) -> Option<HandlerDescriptor>;
    /// Run a handler within the virtual machine.
    fn run_handler(&self, vm_handler_id: &str, message_packet: &mut MessagePacket, callbacks: &dyn HostBackend, allocator: &dyn Allocator) -> Result<(), ErrorCode>;
}

/// A descriptor for a handler.
#[non_exhaustive]
#[derive(Debug, Default, Clone)]
pub struct HandlerDescriptor {}