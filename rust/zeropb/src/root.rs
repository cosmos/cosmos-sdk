#[cfg(not(target_arch = "wasm32"))]
use alloc::alloc::{alloc_zeroed, Layout};

use alloc::boxed::Box;
use core::marker::PhantomPinned;
use core::{
    borrow::{Borrow, BorrowMut},
    marker::PhantomData,
    mem::size_of,
    ops::{Deref, DerefMut},
};
use core::ptr::{null_mut};

use crate::error::Error;
use crate::rel_ptr::MAX_EXTENT;
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
            let buf = __zeropb_alloc_page();
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
        });
    }
}

impl <T: ZeroCopy> Drop for Root<T> {
    fn drop(&mut self) {
        unsafe {
            __zeropb_free_page(self.buf);
        }
    }
}

impl<T: ZeroCopy> Borrow<T> for Root<T> {
    fn borrow(&self) -> &T {
        unsafe { &*self.buf.cast::<T>() }
    }
}

impl<T: ZeroCopy> BorrowMut<T> for Root<T> {
    fn borrow_mut(&mut self) -> &mut T {
        unsafe { &mut *self.buf.cast::<T>() }
    }
}

impl<T: ZeroCopy> Deref for Root<T> {
    type Target = T;

    fn deref(&self) -> &Self::Target {
        unsafe { &*self.buf.cast::<T>() }
    }
}

impl<T: ZeroCopy> DerefMut for Root<T> {
    fn deref_mut(&mut self) -> &mut Self::Target {
        unsafe { &mut *self.buf.cast::<T>() }
    }
}

const STATIC_FREELIST_CAP: usize = 32;
const EXTRA_FREELIST_CAP: usize = 0x10000;
static mut STATIC_FREELIST: [*mut u8; 32] = [null_mut(); 32];
static mut STATIC_FREELIST_LEN: usize = 0;
static mut EXTRA_FREELIST_LEN: usize = 0;

#[no_mangle]
pub extern "C" fn __zeropb_alloc_page() -> *mut u8 {
    unsafe {
        if EXTRA_FREELIST_LEN > 0 {
            let extra_free_list = STATIC_FREELIST[STATIC_FREELIST_CAP - 1] as *mut *mut u8;
            if extra_free_list != null_mut() {
                let ptr = extra_free_list.add(EXTRA_FREELIST_LEN - 1);
                EXTRA_FREELIST_LEN -= 1;
                return *ptr;
            }
        }

        if STATIC_FREELIST_LEN > 0 {
            let ptr = STATIC_FREELIST[STATIC_FREELIST_LEN - 1];
            STATIC_FREELIST_LEN -= 1;
            return ptr;
        }

        return alloc_page();
    }
}

#[no_mangle]
pub extern "C" fn __zeropb_free_page(page: *mut u8) {
    if page.is_null() {
        return;
    }

    unsafe {
        if STATIC_FREELIST_LEN < STATIC_FREELIST_CAP - 1 {
            STATIC_FREELIST[STATIC_FREELIST_LEN] = page;
            STATIC_FREELIST_LEN += 1;
        } else {
            let extra_free_list = STATIC_FREELIST[STATIC_FREELIST_CAP - 1] as *mut *mut u8;
            if extra_free_list == null_mut() {
                STATIC_FREELIST[STATIC_FREELIST_CAP - 1] = page;
            } else if EXTRA_FREELIST_LEN < EXTRA_FREELIST_CAP {
                let extra_free_list = extra_free_list.add(EXTRA_FREELIST_LEN);
                *extra_free_list = page;
                EXTRA_FREELIST_LEN += 1;
            } else {
                free_page(page)
            }
        }
    }
}

#[cfg(target_arch = "wasm32")]
unsafe fn alloc_page() -> *mut u8 {
    let page = core::arch::wasm32::memory_grow(1) as *mut u8;
    // zero memory
    core::ptr::write_bytes(page, 0, 0x10000);
    page
}

#[cfg(target_arch = "wasm32")]
unsafe fn free_page(_page: *mut u8) {
    // leak memory because we can no longer deallocate pages
    // if we hit this point, we're probably in a bad state anyway
    // because over 4GB of memory has been allocated
}

#[cfg(not(target_arch = "wasm32"))]
extern crate std;

#[cfg(not(target_arch = "wasm32"))]
unsafe fn alloc_page() -> *mut u8 {
    alloc_zeroed(Layout::from_size_align(0x10000, 0x10000).unwrap())
}

#[cfg(not(target_arch = "wasm32"))]
unsafe fn free_page(page: *mut u8) {
    std::alloc::dealloc(
        page,
        Layout::from_size_align(0x10000, 0x10000).unwrap(),
    )
}
