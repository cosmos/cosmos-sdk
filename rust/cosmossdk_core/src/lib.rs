#[cfg(target_arch = "wasm32")]
mod wasm;

mod r#extern;

use zeropb;
use zeropb::{ZeroCopy};

#[cfg(not(target_arch = "wasm32"))]
pub mod c;

mod module;
mod handler;
mod router;

