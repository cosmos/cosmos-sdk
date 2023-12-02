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
mod c;

