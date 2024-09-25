//! This module contains the definition of the `MessagePacket` struct.

use crate::data_pointer::DataPointerWrapper;
use crate::header::MessageHeader;

/// A packet containing a message and its header.
pub struct MessagePacket {
    data: *mut MessageHeader,
    len: usize,
}

impl MessagePacket {
    /// Creates a new message packet.
    pub unsafe fn new(data: *mut MessageHeader, len: usize) -> Self {
        Self { data, len }
    }

    /// Returns the message header.
    pub fn header(&self) -> &MessageHeader {
        unsafe { &*(self.data as *const MessageHeader) }
    }

    /// Returns a mutable reference to the message header.
    pub unsafe fn header_mut(&self) -> &mut MessageHeader {
        &mut *self.data
    }

    /// Returns input data pointer 1.
    pub fn in1(&self) -> DataPointerWrapper {
        unsafe { DataPointerWrapper(&mut self.header_mut().in_pointer1, self.data as *const u8, self.len) }
    }

    /// Returns input data pointer 2.
    pub fn in2(&self) -> DataPointerWrapper {
        unsafe { DataPointerWrapper(&mut self.header_mut().in_pointer2, self.data as *const u8, self.len) }
    }

    /// Returns output data pointer 1.
    pub fn out1(&self) -> DataPointerWrapper {
        unsafe { DataPointerWrapper(&mut self.header_mut().out_pointer1, self.data as *const u8, self.len) }
    }

    /// Returns output data pointer 2.
    pub fn out2(&self) -> DataPointerWrapper {
        unsafe { DataPointerWrapper(&mut self.header_mut().out_pointer2, self.data as *const u8, self.len) }
    }
}
