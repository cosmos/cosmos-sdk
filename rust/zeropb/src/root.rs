#[cfg(not(target_arch = "wasm32"))]
extern crate alloc;

#[cfg(not(target_arch = "wasm32"))]
use alloc::alloc::{alloc, dealloc, Layout};

use core::marker::PhantomPinned;
use core::{
    borrow::{Borrow, BorrowMut},
    marker::PhantomData,
    mem::size_of,
    ops::{Deref, DerefMut},
};
use core::ptr::{null_mut};

use crate::error::Error;
use crate::rel_ptr::{MAX_EXTENT, resolve_start_extent};
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

    pub unsafe fn unsafe_wrap(ptr: *mut u8) -> Self {
        assert!(ptr.is_null() || (ptr as usize) & 0xFFFF == 0);

        return Self {
            buf: ptr,
            _phantom: PhantomData,
        };
    }

    pub unsafe fn unsafe_unwrap(&self) -> *mut u8 {
        self.buf
    }

    pub fn as_slice(&self) -> &[u8] {
        unsafe {
            let buf = self.buf;
            let (_, extent_ptr) = resolve_start_extent(buf);
            let extent = *extent_ptr as usize;
            core::slice::from_raw_parts(buf, extent)
        }
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

const STATIC_FREELIST_CAP: usize = 128;
const EXTRA_FREELIST_CAP: usize = 0x10000;
static mut STATIC_FREELIST: [*mut u8; 128] = [null_mut(); 128];
static mut STATIC_FREELIST_LEN: usize = 0;
static mut EXTRA_FREELIST_LEN: usize = 0;

#[no_mangle]
pub extern "C" fn __zeropb_alloc_page() -> *mut u8 {
    unsafe {
        let page = unsafe { do_alloc_page() };
        // zero memory
        let extent_ptr = page.offset((MAX_EXTENT - 2) as isize) as *mut u16;
        let extent = *extent_ptr as usize;
        core::ptr::write_bytes(page, 0, extent);
        *extent_ptr = 0;
        page
    }
}

unsafe fn do_alloc_page() -> *mut u8 {
    if STATIC_FREELIST_LEN > 0 {
        let ptr = STATIC_FREELIST[STATIC_FREELIST_LEN - 1];
        STATIC_FREELIST_LEN -= 1;
        return ptr;
    }

    return alloc_page();
}

#[no_mangle]
pub extern "C" fn __zeropb_free_page(page: *mut u8) {
    if page.is_null() {
        return;
    }

    unsafe {
        if STATIC_FREELIST_LEN < STATIC_FREELIST_CAP {
            STATIC_FREELIST[STATIC_FREELIST_LEN] = page;
            STATIC_FREELIST_LEN += 1;
        } else {
            free_page(page)
        }
    }
}

#[cfg(target_arch = "wasm32")]
unsafe fn alloc_page() -> *mut u8 {
    let page = (core::arch::wasm32::memory_grow(0, 1) * 0x10000usize) as *mut u8;
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
    alloc(Layout::from_size_align(0x10000, 0x10000).unwrap())
}

#[cfg(not(target_arch = "wasm32"))]
unsafe fn free_page(page: *mut u8) {
    dealloc(
        page,
        Layout::from_size_align(0x10000, 0x10000).unwrap(),
    )
}

#[cfg(target_arch = "wasm32")]
#[panic_handler]
fn panic(_info: &core::panic::PanicInfo) -> ! {
    loop {}
}
