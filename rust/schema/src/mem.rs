//! Memory management utilities for codec implementations.
use core::cell::RefCell;
use core::ptr::NonNull;
use bump_scope::{BumpBox, BumpScope, BumpVec};

/// A memory manager that tracks allocated memory using a bump allocator and ensures that
/// memory is deallocated and dropped properly when the manager is dropped.
pub struct MemoryManager<'b, 'a: 'b> {
    scope: &'b BumpScope<'a>,
    handles: RefCell<BumpVec<'b, 'a, NonNull<dyn DeferDrop + 'b>>>,
}

impl<'b, 'a: 'b> MemoryManager<'b, 'a> {
    /// Create a new memory manager.
    pub fn new(scope: &'b BumpScope<'a>) -> MemoryManager<'b, 'a> {
        MemoryManager {
            scope,
            handles: RefCell::new(BumpVec::new_in(scope)),
        }
    }

    /// Get the bump scope for this memory manager.
    pub fn scope(&self) -> &'b bump_scope::BumpScope<'a> {
        self.handles.borrow().bump()
    }

    /// Converts a BumpVec into a borrowed slice in such a way that the drop code
    /// for T (if any) will be executed when the MemoryManager is dropped.
    pub fn unpack_slice<T>(&self, vec: BumpVec<'b, 'a, T>) -> &'a [T] {
        unsafe {
            let b = vec.into_boxed_slice();
            let slice = b.as_non_null_slice().as_ptr() as *const [T];
            struct Dropper<'a, U> {
                b: BumpBox<'a, [U]>,
            }
            let dropper = self.scope().alloc(Dropper { b });
            self.handles.borrow_mut().push(dropper.into_raw() as NonNull<dyn DeferDrop + 'b>);
            &*slice
        }
    }
}

impl<'b, 'a: 'b> Drop for MemoryManager<'b, 'a> {
    fn drop(&mut self) {
        for handle in self.handles.borrow_mut().drain(..) {
            unsafe {
                handle.as_ptr().drop_in_place();
            }
        }
    }
}

trait DeferDrop {}
impl<T> DeferDrop for T {}