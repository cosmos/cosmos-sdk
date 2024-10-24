//! Memory management utilities for codec implementations.

use core::alloc::Layout;
use allocator_api2::alloc::{AllocError, Allocator};
use allocator_api2::boxed::Box;
use allocator_api2::vec::Vec;
use core::cell::Cell;
use core::mem::transmute;
use core::ptr::{drop_in_place, NonNull};
use ixc_message_api::header::{MessageHeader, MESSAGE_HEADER_SIZE};
use ixc_message_api::packet::MessagePacket;

/// A memory manager that tracks allocated memory using a bump allocator and ensures that
/// memory is deallocated and dropped properly when the manager is dropped.
pub struct MemoryManager {
    #[cfg(feature = "bumpalo")]
    bump: bumpalo::Bump,
    #[cfg(not(feature = "bumpalo"))]
    bump: crate::bump::Bump,
    drop_cells: Cell<Option<NonNull<DropCell>>>,
}

struct DropCell {
    dropper: NonNull<dyn DeferDrop>,
    next: Option<NonNull<DropCell>>,
}

impl MemoryManager {
    /// Create a new memory manager.
    pub fn new() -> MemoryManager {
        MemoryManager {
            bump: Default::default(),
            drop_cells: Cell::new(None),
        }
    }

    /// Converts a BumpVec into a borrowed slice in such a way that the drop code
    /// for T (if any) will be executed when the MemoryManager is dropped.
    pub(crate) fn unpack_slice<'a, T>(&'a self, vec: Vec<T, &'a dyn Allocator>) -> &'a [T] {
        unsafe {
            let ptr = vec.as_ptr();
            let len = vec.len();
            let slice = core::slice::from_raw_parts(ptr, len);
            let (dropper, _) = Box::into_non_null(Box::new_in(vec, &self.bump));
            let drop_cell = Box::new_in(DropCell {
                /// Rust doesn't know what the lifetime of this data is, but we do because
                /// we allocated it and own the allocator,
                /// so we transmute it to have the appropriate lifetime
                dropper: transmute(dropper as NonNull<dyn DeferDrop>),
                next: self.drop_cells.get(),
            }, &self.bump);
            let (drop_cell, _) = Box::into_non_null(drop_cell);
            self.drop_cells.set(Some(drop_cell));
            slice
        }
    }
}

unsafe impl Allocator for MemoryManager {
    fn allocate(&self, layout: Layout) -> Result<NonNull<[u8]>, AllocError> {
        (&self.bump).allocate(layout)
    }

    fn allocate_zeroed(&self, layout: Layout) -> Result<NonNull<[u8]>, AllocError> {
        (&self.bump).allocate_zeroed(layout)
    }

    unsafe fn deallocate(&self, ptr: NonNull<u8>, layout: Layout) {
        (&self.bump).deallocate(ptr, layout)
    }

    unsafe fn grow(&self, ptr: NonNull<u8>, old_layout: Layout, new_layout: Layout) -> Result<NonNull<[u8]>, AllocError> {
        (&self.bump).grow(ptr, old_layout, new_layout)
    }

    unsafe fn grow_zeroed(&self, ptr: NonNull<u8>, old_layout: Layout, new_layout: Layout) -> Result<NonNull<[u8]>, AllocError> {
        (&self.bump).grow_zeroed(ptr, old_layout, new_layout)
    }

    unsafe fn shrink(&self, ptr: NonNull<u8>, old_layout: Layout, new_layout: Layout) -> Result<NonNull<[u8]>, AllocError> {
        (&self.bump).shrink(ptr, old_layout, new_layout)
    }
}

impl Drop for MemoryManager {
    fn drop(&mut self) {
        let mut drop_cell = self.drop_cells.get();
        while let Some(cell) = drop_cell {
            unsafe {
                let cell = cell.as_ref();
                drop_in_place(cell.dropper.as_ptr());
                drop_cell = cell.next;
            }
        }
    }
}

trait DeferDrop {}
impl<T> DeferDrop for T {}