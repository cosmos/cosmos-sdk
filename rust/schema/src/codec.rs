//! The codec trait.
use crate::buffer::{Writer, WriterFactory};
use crate::decoder::DecodeError;
use crate::encoder::EncodeError;
use crate::mem::MemoryManager;
use crate::value::Value;
use bump_scope::{Bump, BumpVec};

/// Trait implemented by encoding protocols.
pub trait Codec {
    /// Encode a value.
    fn encode_value<'a, V: Value<'a>, F: WriterFactory>(value: &V, writer_factory: F) -> Result<F::Output, EncodeError>;
    /// Decode a value.
    fn decode_value<'a, V: Value<'a>>(input: &'a [u8], memory_manager: &'a MemoryManager) -> Result<V, DecodeError>;
}

// #[cfg(test)]
// mod tests {
//     use super::*;
//     use alloc::string::String;
//     use bump_scope::{bump_vec, Bump, BumpBox, BumpScope, BumpVec};
//     use core::ptr::NonNull;
//     extern crate std;
//
//     struct HasString {
//         s: std::string::String,
//     }
//
//     impl Drop for HasString {
//         fn drop(&mut self) {
//             std::println!("do drop {}", self.s);
//         }
//     }
//
//     trait DeferDrop {}
//
//     fn test1<'a: 'b, 'b>(scope: &'b BumpScope<'a>) -> (NonNull<dyn DeferDrop + 'b>, &'a [HasString]) {
//         struct Dropper<'a> {
//             str_box: BumpBox<'a, [HasString]>,
//         }
//         let mut strings = BumpVec::new_in(scope);
//         strings.push(HasString {
//             s: String::from("hello"),
//         });
//         strings.push(HasString {
//             s: String::from("world"),
//         });
//         strings.push(HasString {
//             s: String::from("foo"),
//         });
//         strings.push(HasString {
//             s: String::from("bar"),
//         });
//         let str_box = strings.into_boxed_slice();
//         unsafe {
//             let str_slice = str_box.as_non_null_slice().as_ptr() as *const [HasString];
//             let dropper = scope.alloc(Dropper {
//                 str_box,
//             });
//             (dropper.into_raw() as NonNull<dyn DeferDrop + 'b>, &*str_slice)
//         }
//     }
//
//     #[test]
//     fn test_test1() {
//         let bump = Bump::new();
//         let scope = bump.as_scope();
//         let mut todrop = bump_vec![in scope];
//         let (dropper, strs) = test1(&scope);
//         todrop.push(dropper);
//         for s in strs {
//             std::println!("{}", s.s);
//         }
//         unsafe {
//             for d in todrop.drain(..) {
//                 d.drop_in_place()
//             }
//         }
//         drop(todrop);
//     }
//
//     fn test2() -> *const dyn DeferDrop {
//         struct Foo {}
//         impl Drop for Foo {
//             fn drop(&mut self) {
//                 std::println!("dropping foo");
//             }
//         }
//         let foo = Foo {};
//         &foo as *const Foo as *const dyn DeferDrop
//     }
//
//     #[test]
//     fn test_test2() {
//         let _ = test2();
//     }
// }
