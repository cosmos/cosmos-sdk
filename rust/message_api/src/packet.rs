//! This module contains the definition of the `MessagePacket` struct.

use crate::data_pointer::DataPointerWrapper;
use crate::header::MessageHeader;

/// A packet containing a message and its header.
pub struct MessagePacket {
    data: *mut MessageHeader,
    len: usize,
}

impl MessagePacket {
    pub unsafe fn new(data: *mut MessageHeader, len: usize) -> Self {
        Self { data, len }
    }

    pub unsafe fn header(&self) -> &MessageHeader {
        &*(self.data as *const MessageHeader)
    }

    pub unsafe fn header_mut(&self) -> &mut MessageHeader {
        &mut *self.data
    }

    pub fn in1<'a>(&self) -> DataPointerWrapper<'a> {
        unsafe { DataPointerWrapper(&mut self.header_mut().in_pointer1, self.data as *const u8, self.len) }
    }

    pub fn in2<'a>(&self) -> DataPointerWrapper<'a> {
        unsafe { DataPointerWrapper(&mut self.header_mut().in_pointer2, self.data as *const u8, self.len) }
    }

    pub fn out1<'a>(&self) -> DataPointerWrapper<'a> {
        unsafe { DataPointerWrapper(&mut self.header_mut().out_pointer1, self.data as *const u8, self.len) }
    }

    pub fn out2<'a>(&self) -> DataPointerWrapper<'a> {
        unsafe { DataPointerWrapper(&mut self.header_mut().out_pointer2, self.data as *const u8, self.len) }
    }
}
