#![cfg(target_arch = "wasm32")]

use crate::root::Root;
use alloc::alloc::{alloc_zeroed, dealloc, Layout};
use lol_alloc::{FreeListAllocator, LockedAllocator};

#[global_allocator]
static ALLOCATOR: LockedAllocator<FreeListAllocator> =
    LockedAllocator::new(FreeListAllocator::new());

#[panic_handler]
fn panic(_info: &core::panic::PanicInfo) -> ! {
    loop {}
}

#[no_mangle]
pub extern "C" fn __zeropb_alloc_page() -> *mut u8 {
    // let root = Root::<u8>::new();
    // root.buf
    todo!()
}

#[no_mangle]
pub extern "C" fn __zeropb_free_page(ptr: *const u8) {
    unsafe {
        dealloc(
            ptr as *mut u8,
            Layout::from_size_align_unchecked(0x10000, 0x10000),
        );
    }
}
