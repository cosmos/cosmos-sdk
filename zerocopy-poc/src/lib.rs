mod cosmos_proto;
mod module_abi;
mod lib2;

use std::marker::PhantomData;
use std::ops::{Deref, DerefMut};

struct Ptr<T: ?Sized> {
    offset: u16,
    _phantom: PhantomData<T>,
}

struct Array<T: ?Sized> {
    offset: u16,
    _phantom: PhantomData<T>,
}

// impl<T> Deref for Ptr<T> {
//     type Target = Option<T>;
//
//     fn deref(&self) -> &Self::Target {
//         match self.offset {
//             0 => &None,
//             _ => unimplemented!(),
//         }
//     }
// }
//
// impl<T> DerefMut for Ptr<T> {
//     fn deref_mut(&self) -> &mut Self::Target {
//         match self.offset {
//             0 => &mut None,
//             _ => unimplemented!(),
//         }
//     }
// }
//
// // impl<T> Deref for RelArray<T> {
// //     type Target = Option<[T]>;
// //
// //     fn deref(&self) -> &Self::Target {
// //         match self.offset {
// //             0 => &None,
// //             _ => unimplemented!(),
// //         }
// //     }
// // }
//
// struct Buffer {
//     data: *mut u8,
//     size: usize,
//     capacity: usize,
// }
//
// struct MsgSend {
//     // from_address: RelPtr<String>,
//     // to_address: RelPtr<String>,
//     // amount: RelPtr<Vec<Coin>>
// }
//
// struct Coin {
//     denom: Ptr<String>,
//     amount: Ptr<String>,
// }
//
// fn main() {
//     println!("Hello, world!");
// }
