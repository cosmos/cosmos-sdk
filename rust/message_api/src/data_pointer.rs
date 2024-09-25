use crate::header::MESSAGE_HEADER_SIZE;
use crate::packet::MessagePacket;

#[derive(Default)]
pub union DataPointer {
    pub native_pointer: NativePointer,
    pub local_pointer: LocalPointer,
}

struct NativePointer {
    pub len: u32,
    pub capacity: u32,
    pub pointer: *const (),
}

impl Default for NativePointer {
    fn default() -> Self {
        Self {
            len: 0,
            capacity: 0,
            pointer: core::ptr::null(),
        }
    }
}

#[derive(Default)]
struct LocalPointer {
    pub len: u32,
    pub offset: u32,
    pub zero: u64,
}

impl DataPointer {
    unsafe fn data(&self, message_packet: *const u8, packet_len: usize) -> &[u8] {
        if self.local_pointer.zero == 0 {
            if self.local_pointer.offset < MESSAGE_HEADER_SIZE as u32 {
                return &[];
            }
            if (self.local_pointer.offset + self.local_pointer.len) as usize > packet_len {
                return &[];
            }
            unsafe {
                return core::slice::from_raw_parts(message_packet.offset(self.local_pointer.offset as isize), self.local_pointer.len as usize);
            }
        }
        unsafe {
            core::slice::from_raw_parts(self.native_pointer.pointer as *const u8, self.native_pointer.len as usize)
        }
    }
}

pub struct DataPointerWrapper<'a>(&'a mut DataPointer, *const u8, usize);

impl<'a> DataPointerWrapper<'a> {
    pub fn get(&self) -> &[u8] {
        unsafe {
            self.0.data(self.1, self.2)
        }
    }

    pub fn set_slice(&mut self, data: *const [u8]) {
        unsafe {
            self.0.native_pointer.pointer = data as *const ();
            let len = (*data).len() as u32;
            self.0.native_pointer.len = len;
            self.0.native_pointer.capacity = len;
        }
    }
}