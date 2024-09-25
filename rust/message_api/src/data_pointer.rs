use crate::header::MESSAGE_HEADER_SIZE;

#[derive(Copy, Clone)]
pub union DataPointer {
    pub native_pointer: NativePointer,
    pub local_pointer: LocalPointer,
}

impl Default for DataPointer {
    fn default() -> Self {
        Self {
            local_pointer: LocalPointer::default(),
        }
    }
}

#[derive(Copy, Clone)]
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

#[derive(Default, Copy, Clone)]
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

pub struct DataPointerWrapper<'a>(pub(crate) &'a mut DataPointer, pub(crate) *const u8, pub(crate) usize);

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