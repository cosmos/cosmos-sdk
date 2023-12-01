#[cfg(target_arch = "wasm32")]
mod wasm;

#[cfg(not(target_arch = "wasm32"))]
mod ffi;

mod store;
pub use store::{KVStore, KVStoreService};
mod context;
mod tonic;

use zeropb;
use zeropb::{ClientConn, ZeroCopy};

