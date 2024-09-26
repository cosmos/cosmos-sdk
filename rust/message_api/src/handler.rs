use core::alloc::Layout;
use crate::code::Code;
use crate::packet::MessagePacket;

pub trait Handler {
    fn handle(&self, message_packet: &mut MessagePacket, callbacks: &dyn HostBackend) -> HandlerCode;
}

#[non_exhaustive]
pub trait HostBackend {
    fn invoke(&self, message_packet: &mut MessagePacket) -> Code;
    unsafe fn alloc(&self, layout: Layout) -> Result<*mut u8, AllocError>;
}

#[derive(Debug)]
pub struct AllocError;

#[derive(Clone, Copy, PartialEq, Eq, Debug)]
pub enum HandlerCode {
    Ok,
    HandlerError(u32),
}
