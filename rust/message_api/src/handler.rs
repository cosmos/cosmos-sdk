//! The raw handler and host backend interfaces.
use core::alloc::Layout;
use crate::code::ErrorCode;
use crate::packet::MessagePacket;

/// A handler for an account.
pub trait RawHandler {
    /// The name of the handler.
    fn name(&self) -> &'static str;

    /// Handle a message packet.
    fn handle(&self, message_packet: &mut MessagePacket, callbacks: &dyn HostBackend) -> Result<(), HandlerErrorCode>;
}

/// A host backend for the handler.
pub trait HostBackend {
    /// Invoke a message packet.
    fn invoke(&self, message_packet: &mut MessagePacket) -> Result<(), ErrorCode>;
    /// Allocate memory for a message response.
    /// The memory management expectation of handlers is that the caller
    /// deallocates both the memory it allocated and any memory allocated
    /// for the response by the callee.
    /// The alloc function in the host backend should return a pointer to
    /// memory that the caller knows how to free and such allocated
    /// memory should be referenced in the message packet's out pointers.
    unsafe fn alloc(&self, layout: Layout) -> Result<*mut u8, AllocError>;
}

/// An allocation error.
#[derive(Debug)]
pub struct AllocError;

/// A code that a handler can return.
#[derive(Clone, Copy, PartialEq, Eq, Debug)]
pub enum HandlerErrorCode {
    MessageNotHandled = 0,
    /// The handler encountered an error.
    Custom(u32),
}
