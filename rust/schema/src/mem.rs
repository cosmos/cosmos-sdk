//! Memory management utilities for codec implementations.
use core::cell::RefCell;
use core::ptr::NonNull;
use bump_scope::{Bump, BumpBox, BumpVec};

/// A memory manager that tracks allocated memory using a bump allocator and ensures that
/// memory is deallocated and dropped properly when the manager is dropped.
pub struct MemoryManager<'a> {
    bump: Bump,
    handles: RefCell<Option<BumpVec<'a, 'a, NonNull<dyn DeferDrop + 'a>>>>,
}

impl<'a> MemoryManager<'a> {
    /// Create a new memory manager.
    pub fn new() -> MemoryManager<'a> {
        let bump = Bump::new();
        MemoryManager {
            bump,
            handles: RefCell::new(None),
        }
    }

    pub(crate) fn new_vec<'b, T>(&'a self) -> BumpVec<'b, 'a, T> {
        BumpVec::new_in(&self.bump)
    }

    /// Converts a BumpVec into a borrowed slice in such a way that the drop code
    /// for T (if any) will be executed when the MemoryManager is dropped.
    pub(crate) fn unpack_slice<T>(&self, vec: BumpVec<'_, 'a, T>) -> &'a [T] {
        unsafe {
            let b = vec.into_boxed_slice();
            let slice = b.as_non_null_slice().as_ptr() as *const [T];
            struct Dropper<'a, U> {
                b: BumpBox<'a, [U]>,
            }
            let dropper = self.bump.alloc(Dropper { b });
            let mut handles = self.handles.borrow_mut();
            if handles.is_none() {
                *handles = Some(BumpVec::new_in(self.bump.as_scope()));
            }
            let mut handles = handles.as_mut().unwrap();
            handles.push(dropper.into_raw() as NonNull<dyn DeferDrop + 'a>);
            &*slice
        }
    }

}

impl<'a> Drop for MemoryManager<'a> {
    fn drop(&mut self) {
        if let Some(handles) = self.handles.get_mut() {
            for handle in handles.drain(..) {
                unsafe {
                    handle.as_ptr().drop_in_place();
                }
            }
        }
    }
}

trait DeferDrop {}
impl<T> DeferDrop for T {}