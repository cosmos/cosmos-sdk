#[cfg(target_arch = "wasm32")]
mod wasm;

mod r#extern;

mod store;
pub use store::{KVStore, KVStoreService};
mod context;

use zeropb;
use zeropb::{ClientConn, ZeroCopy};

#[cfg(feature = "tonic")]
pub mod tonic;

#[cfg(not(target_arch = "wasm32"))]
pub mod c;

mod module;
mod services;

