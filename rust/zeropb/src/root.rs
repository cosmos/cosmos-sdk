use alloc::alloc::{alloc_zeroed, Layout};
use alloc::boxed::Box;
use core::{
    borrow::{Borrow, BorrowMut},
    marker::PhantomData,
    mem::size_of,
    ops::{Deref, DerefMut},
};
use core::marker::PhantomPinned;

use crate::error::Error;
use crate::util::MAX_EXTENT;
use crate::zerocopy::ZeroCopy;

pub struct RawRoot {
    pub(crate) buf: *mut u8,
    _phantom: PhantomPinned,
}

pub struct Root<T: ZeroCopy> {
    pub(crate) buf: *mut u8,
    _phantom: PhantomData<T>,
}

impl<T: ZeroCopy> Root<T> {
    pub fn new() -> Self {
        unsafe {
            let buf = alloc_zeroed(Layout::from_size_align_unchecked(0x10000, 0x10000));
            assert!(!buf.is_null());
            assert_eq!((buf as usize) & 0xFFFF, 0);
            let extent_ptr = buf.offset(MAX_EXTENT as isize) as *mut u16;
            let size_of_t = size_of::<T>();
            assert!(size_of_t <= MAX_EXTENT);
            *extent_ptr = size_of_t as u16;
            Self {
                buf,
                _phantom: PhantomData,
            }
        }
    }

    pub fn wrap(buf: Box<[u8]>) -> Result<Self, Error> {
        if buf.len() < 0x10000 {
            return Err(Error::InvalidBuffer);
        }

        let ptr = Box::into_raw(buf) as *mut u8;
        if (ptr as usize) & 0xFFFF != 0 {
            return Err(Error::InvalidBuffer);
        }

        return Ok(Self {
            buf: ptr,
            _phantom: PhantomData,
        })
    }
}

impl<T: ZeroCopy> Borrow<T> for Root<T> {
    fn borrow(&self) -> &T {
        unsafe {
            &*self.buf.cast::<T>()
        }
    }
}

impl<T: ZeroCopy> BorrowMut<T> for Root<T> {
    fn borrow_mut(&mut self) -> &mut T {
        unsafe {
            &mut *self.buf.cast::<T>()
        }
    }
}

impl<T: ZeroCopy> Deref for Root<T> {
    type Target = T;

    fn deref(&self) -> &Self::Target {
        unsafe {
            &*self.buf.cast::<T>()
        }
    }
}

impl<T: ZeroCopy> DerefMut for Root<T> {
    fn deref_mut(&mut self) -> &mut Self::Target {
        unsafe {
            &mut *self.buf.cast::<T>()
        }
    }
}

