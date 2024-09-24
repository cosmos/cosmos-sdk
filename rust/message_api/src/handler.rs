use crate::code::Code;
use crate::packet::MessagePacket;

pub trait Handler {
    fn handle(&self, message_packet: &mut MessagePacket, callbacks: &HostCallbacks) -> HandlerCode;
}

#[non_exhaustive]
pub struct HostCallbacks {
    pub invoke: InvokeFn,
}

pub type InvokeFn = fn(&mut MessagePacket) -> Code;

#[derive(Clone, Copy, PartialEq, Eq, Debug)]
pub enum HandlerCode {
    Ok,
    HandlerError(u32),
}
