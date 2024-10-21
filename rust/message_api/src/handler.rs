//! The raw handler and host backend interfaces.
use crate::code::ErrorCode;
use crate::packet::MessagePacket;

/// A handler for an account.
pub trait RawHandler {
    /// Handle a message packet.
    fn handle(&self, message_packet: &mut MessagePacket, callbacks: &dyn HostBackend, allocator: &dyn Allocator) -> Result<(), ErrorCode>;
}

pub use allocator_api2::alloc::Allocator;

/// A host backend for the handler.
pub trait HostBackend {
    /// Invoke a message packet.
    fn invoke(&self, message_packet: &mut MessagePacket, allocator: &dyn Allocator) -> Result<(), ErrorCode>;
}
