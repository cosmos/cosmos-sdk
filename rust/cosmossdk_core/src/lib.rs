#[cfg(target_arch = "wasm32")]
mod wasm;

mod code;
// mod module;
// mod handler;
mod router;
mod result;
mod context;
mod module;
// mod store;
// mod service;
// mod client;

pub use code::Code;
pub use router::{Server};
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