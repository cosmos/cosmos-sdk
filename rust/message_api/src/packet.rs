//! This module contains the definition of the `MessagePacket` struct.

/// A packet containing a message and its header.
pub struct MessagePacket {
    data: *mut u8,
    len: usize,
}

