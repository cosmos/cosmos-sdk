extern crate alloc;
extern crate core;

use core::{
    marker::{PhantomData, PhantomPinned},
    ops::{Deref, DerefMut},
    slice::from_raw_parts,
    str::from_utf8_unchecked,
    ptr,
};
use alloc::alloc::{alloc_zeroed, Layout};

pub struct Root<T> {
    buf: *mut u8,
    _phantom: PhantomData<T>,
}

impl<T> Root<T> {
    fn new() -> Self {
        Self {
            buf: unsafe { alloc_zeroed(Layout::from_size_align_unchecked(0xFFFF, 0xFFFF)) },
            _phantom: PhantomData,
        }
    }
}

fn resolve_rel_ptr(base: usize, offset: i16, min_len: u16) -> usize {
    let buf_start = base as usize & !0xFFFF;
    let target = (base as isize + offset as isize) as usize;
    assert!(target >= buf_start);
    let buf_end = buf_start + 0xFFFF;
    assert!((target + min_len as usize) < buf_end);
    target
}

#[repr(C)]
pub struct BytesPtr {
    offset: i16,
    length: u16,
    _phantom: PhantomData<[u8]>,
}

impl<'a> Deref for BytesPtr {
    type Target = [u8];

    fn deref(&self) -> &Self::Target {
        unsafe {
            unsafe {
                let base = (self as *const Self).cast::<u8>();
                let target = resolve_rel_ptr(base as usize, self.offset, self.length);
                from_raw_parts(target as *const u8, self.length as usize)
            }
        }
    }
}

#[repr(C)]
pub struct StringPtr {
    ptr: BytesPtr,
    _phantom: PhantomData<str>,
}

impl<'a> Deref for StringPtr {
    type Target = str;

    fn deref(&self) -> &Self::Target {
        unsafe {
            unsafe {
                from_utf8_unchecked(self.ptr.deref())
            }
        }
    }
}

#[repr(C)]
pub struct RepeatedPtr<T> {
    offset: i16,
    length: u16,
    _phantom: PhantomData<[T]>,
}

struct RepeatedSegmentHeader {
    capacity: u16,
    next_offset: i16,
}

// #[repr(C)]
// pub struct Enum<T, const MaxValue: u32> {
//     value: u32,
//     _phantom: PhantomData<T>,
// }
//
// impl <T, const MaxValue: u32> Enum<T, MaxValue> {
//     fn get(&self) -> Result<&T, u32> {
//         if self.value > MaxValue {
//             Err(self.value)
//         } else {
//             Ok(&self.value as T)
//         }
//     }
//
//     fn set(&mut self, value: T) {
//         self.value = value
//     }
// }
//
// #[repr(C)]
// pub struct OneOf<T, const MaxValue: u32> {
//     value: T
// }
//
// impl <T, const MaxValue: u32> OneOf<T, MaxValue> {
//     fn get(&self) -> Result<&T, u32> {
//         let discriminant = unsafe { *<*const _>::from(self).cast::<u32>() };
//         if discriminant > MaxValue {
//             Err(discriminant)
//         } else {
//             Ok(&self.value)
//         }
//     }
//
//     fn set(&mut self, value: T) {
//         self.value = value
//     }
// }
//
// #[repr(C)]
// pub struct Option<T> {
//     some: bool,
//     value: T,
// }

#[cfg(test)]
mod tests {
    use crate::{BytesPtr, Root};

    #[test]
    fn test_root() {
        let r = Root::<()>::new();
    }

    struct Test1 {
        bytes: BytesPtr,
    }

    #[test]
    fn no_copy() {
        let mut t1 = Test1 {
            bytes: BytesPtr {
                offset: 0,
                length: 0,
                _phantom: Default::default(),
            }
        };
        let mut t2 = Test1 {
            bytes: BytesPtr {
                offset: 0,
                length: 0,
                _phantom: Default::default(),
            }
        };
        t1.bytes = t2.bytes;
    }
}
