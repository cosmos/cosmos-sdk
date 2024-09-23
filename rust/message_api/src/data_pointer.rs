use crate::header::MESSAGE_HEADER_SIZE;

pub struct DataPointer {
    pub native_pointer: u64,
    pub len: u32,
    pub offset_or_capacity: u32,
}

impl DataPointer {
    pub unsafe fn data(&self, message_packet: *const u8, packet_len: usize) -> &[u8] {
        if self.native_pointer == 0 {
            if self.offset_or_capacity < MESSAGE_HEADER_SIZE as u32 {
                return &[];
            }
            if (self.offset_or_capacity + self.len) as usize > packet_len {
                return &[];
            }
            unsafe {
                return core::slice::from_raw_parts(message_packet.offset(self.offset_or_capacity as isize), self.len as usize);
            }
        }
        unsafe {
            core::slice::from_raw_parts(self.native_pointer as *const u8, self.len as usize)
        }
    }
}
