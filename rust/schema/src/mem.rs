use core::cell::RefCell;
use core::ptr::NonNull;
use bump_scope::{BumpBox, BumpScope, BumpVec};

pub struct MemoryManager<'b, 'a: 'b> {
    scope: &'b BumpScope<'a>,
    handles: RefCell<BumpVec<'b, 'a, NonNull<dyn DeferDrop + 'b>>>,
}

impl<'b, 'a: 'b> MemoryManager<'b, 'a> {
    pub fn new(scope: &'b BumpScope<'a>) -> MemoryManager<'b, 'a> {
        MemoryManager {
            scope,
            handles: RefCell::new(BumpVec::new_in(scope)),
        }
    }

    pub fn scope(&self) -> &'b bump_scope::BumpScope<'a> {
        self.handles.borrow().bump()
    }

    pub fn unpack_slice<T>(&self, vec: BumpVec<'b, 'a, T>) -> &'b [T] {
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