//! A pointer to input or output data in a message packet.
use crate::header::MESSAGE_HEADER_SIZE;

/// A pointer to input or output data in a message packet.
#[derive(Copy, Clone)]
pub union DataPointer {
    /// A pointer to the data outside the message packet itself.
    pub native_pointer: NativePointer,
    /// A pointer to the data inside the message packet.
    pub local_pointer: LocalPointer,
}

impl Default for DataPointer {
    fn default() -> Self {
        Self {
            local_pointer: LocalPointer::default(),
        }
    }
}

/// A pointer to data outside the message packet.
#[derive(Copy, Clone)]
pub struct NativePointer {
    /// The length of the data.
    pub len: u32,
    /// The capacity of the data.
    pub capacity: u32,
    /// The pointer to the data.
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
/// A pointer to data inside the message packet.
pub struct LocalPointer {
    /// The length of the data.
    pub len: u32,
    /// The offset of the data from the start of the message packet.
    pub offset: u32,
    /// Should be set to zero to denote that the data is inside the message packet
    /// and not outside.
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

/// A wrapper around a `DataPointer` that provides a safe interface to the data.
pub struct DataPointerWrapper<'a>(pub(crate) &'a mut DataPointer, pub(crate) *const u8, pub(crate) usize);

impl<'a> DataPointerWrapper<'a> {
    /// Returns the data as a slice of bytes.
    pub fn get(&self) -> &[u8] {
        unsafe {
            self.0.data(self.1, self.2)
        }
    }

    /// Sets a slice of bytes as the data that lives outside the message packet.
    pub unsafe fn set_slice(&mut self, data: *const [u8]) {
        unsafe {
            self.0.native_pointer.pointer = data as *const ();
            let len = (*data).len() as u32;
            self.0.native_pointer.len = len;
            self.0.native_pointer.capacity = len;
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_data_pointer_default_size() {
        assert_eq!(core::mem::size_of::<DataPointer>(), 16);
    }

    #[test]
    fn test_data_pointer() {
        let mut data_pointer = DataPointer::default();
        let data = [1, 2, 3, 4, 5];
        let data_ptr = data.as_ptr();
        unsafe {
            data_pointer.native_pointer.pointer = data_ptr as *const ();
            data_pointer.native_pointer.len = data.len() as u32;
            data_pointer.native_pointer.capacity = data.len() as u32;
        }
        let data_pointer_wrapper = DataPointerWrapper(&mut data_pointer, data_ptr as *const u8, data.len());
        assert_eq!(data_pointer_wrapper.get(), data);
    }
}