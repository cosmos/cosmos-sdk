#![cfg_attr(feature = "no_std", no_std)]

#![no_implicit_prelude]

#[cfg(feature="alloc")]
extern crate alloc;

extern crate core;

#[cfg(target_arch = "wasm32")]
mod wasm;

mod code;
// mod module;
// mod handler;
mod router;
mod result;
mod context;
mod module;
mod raw;
mod error;
mod sync;

pub mod store;

#[cfg(any(test, feature = "test-util"))]
pub mod testing;
// mod async;
// mod store;
// mod service;
// mod client;

pub use code::Code;
pub use router::{Router, Server, ClientConnection, Client};
pub use module::{Module};
pub use context::{Context};
pub use result::Result;

// pub mod cosmos {
//     pub mod core {
//         pub mod v1alpha1 {
//             pub mod bundle {
//                 include!("cosmos/core/v1alpha1/bundle.rs");
//             }
//         }
//     }
// }
