//! This module contains the definition of the `MessagePacket` struct.

use crate::header::MessageHeader;

/// A packet containing a message and its header.
pub struct MessagePacket {
    data: *mut u8,
    len: usize,
}

impl MessagePacket {
    pub unsafe fn new(data: *mut u8, len: usize) -> Self {
        Self { data, len }
    }

    pub unsafe fn header(&self) -> &MessageHeader {
        &*(self.data as *const MessageHeader)
    }

    pub unsafe fn header_mut(&self) -> &mut MessageHeader {
        &mut *(self.data as *mut MessageHeader)
    }

    pub unsafe fn in_data1(&self) -> &[u8] {
        self.header().in_pointer1.data(self.data, self.len)
    }

    pub unsafe fn in_data2(&self) -> &[u8] {
        self.header().in_pointer2.data(self.data, self.len)
    }
}
