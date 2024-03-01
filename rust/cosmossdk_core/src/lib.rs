#[cfg(target_arch = "wasm32")]
mod wasm;

mod r#extern;

mod context;

use zeropb;
use zeropb::{ZeroCopy};

#[cfg(feature = "tonic")]
pub mod tonic;

#[cfg(not(target_arch = "wasm32"))]
pub mod c;

mod module;
mod services;

