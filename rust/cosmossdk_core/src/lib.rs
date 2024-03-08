#[cfg(target_arch = "wasm32")]
mod wasm;

use zeropb;
use zeropb::{ZeroCopy};

mod module;
mod handler;
mod router;
mod store;
mod service;
mod client;

pub mod cosmos {
    pub mod core {
        pub mod v1alpha1 {
            pub mod bundle {
                include!("cosmos/core/v1alpha1/bundle.rs");
            }
        }
    }
}