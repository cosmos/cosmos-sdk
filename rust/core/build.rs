//! Build script.
use quote::quote;
use std::collections::HashMap;
use std::io::Write;

fn main() {
    let known_accounts = std::env::var("IXC_KNOWN_ACCOUNTS").unwrap_or_default();
    println!("cargo:rerun-if-env-changed=IXC_KNOWN_ACCOUNTS");

    let known_accounts: HashMap<String, String> = toml::from_str(&known_accounts).unwrap_or_default();

    let mut entries: Vec<_> = known_accounts.iter()
        .map(|(k, v)| (k, u128::from_str_radix(v, 16).unwrap()))
        .collect();
    entries.sort_by(|(k1, _), (k2, _)| k1.cmp(k2));

    let entries = entries.iter().map(|(k, v)| {
        quote! {
            (#k, #v)
        }
    });

    let out_dir = std::env::var("OUT_DIR").unwrap();
    let mut file = std::fs::File::create(format!("{}/known_accounts.rs", out_dir)).unwrap();

    let output = quote! {
        /// Well-known account mappings.
        pub const KNOWN_ACCOUNTS: &[(&str, u128)] = &[
            #(#entries),*
        ];
    };

    write!(file, "{}", output).unwrap();
}