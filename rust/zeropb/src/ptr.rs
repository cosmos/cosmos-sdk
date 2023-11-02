use core::marker::PhantomData;
use core::ops::Deref;
use crate::util::resolve_rel_ptr;
use crate::zerocopy::ZeroCopy;

struct Ptr<T: ZeroCopy> {
    offset: i16,
    length: u16,
    _phantom: PhantomData<T>,
}

unsafe impl <T: ZeroCopy> ZeroCopy for Ptr<T> {}

// impl <T:ZeroCopy> Deref for Ptr<T> {
//     type Target = T;
//
//     fn deref(&self) -> &Self::Target {
//         unsafe {
//             let base = (self as *const Self).cast::<u8>();
//             let target = resolve_rel_ptr(base, self.offset, self.length);
//             &*target.cast()
//         }
//     }
// }