//! Memory management utilities for codec implementations.

use allocator_api2::alloc::Allocator;
use allocator_api2::boxed::Box;
use allocator_api2::vec::Vec;
use bump_scope::Bump;
use core::cell::Cell;
use core::intrinsics::transmute;
use core::ptr::{drop_in_place, NonNull};

/// A memory manager that tracks allocated memory using a bump allocator and ensures that
/// memory is deallocated and dropped properly when the manager is dropped.
pub struct MemoryManager {
    pub(crate) bump: Bump,
    drop_cells: Cell<Option<NonNull<DropCell>>>,
}

struct DropCell {
    data: NonNull<dyn DeferDrop>,
    next: Option<NonNull<DropCell>>,
}

impl MemoryManager {
    /// Create a new memory manager.
    pub fn new() -> MemoryManager {
        MemoryManager {
            bump: Bump::new(),
            drop_cells: Cell::new(None),
        }
    }

    /// Get the allocator for this memory manager.
    pub fn allocator(&self) -> &dyn Allocator {
        &self.bump
    }

    /// Converts a BumpVec into a borrowed slice in such a way that the drop code
    /// for T (if any) will be executed when the MemoryManager is dropped.
    pub(crate) fn unpack_slice<'a, T>(&'a self, vec: Vec<T, &'a dyn Allocator>) -> &'a [T] {
        unsafe {
            let ptr = vec.as_ptr();
            let len = vec.len();
            let slice = core::slice::from_raw_parts(ptr, len);
            let (dropper,_) = Box::into_non_null(Box::new_in(vec, &self.bump));
            let drop_cell = Box::new_in(DropCell {
                data: transmute(dropper as NonNull<dyn DeferDrop>),
                next: self.drop_cells.get(),
            }, &self.bump);
            let (drop_cell, _) = Box::into_non_null(drop_cell);
            self.drop_cells.set(Some(drop_cell));
            slice
        }
    }
}

impl Drop for MemoryManager {
    fn drop(&mut self) {
        let mut drop_cell = self.drop_cells.get();
        while let Some(cell) = drop_cell {
            unsafe {
                let cell = cell.as_ref();
                drop_in_place(cell.data.as_ptr());
                drop_cell = cell.next;
            }
        }
    }
}

trait DeferDrop {}
impl<T> DeferDrop for T {}