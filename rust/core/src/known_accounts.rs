//! Well-known account mappings.

use ixc_message_api::AccountID;

include!(concat!(env!("OUT_DIR"), "/known_accounts.rs"));

/// Get the account ID for a given name.
pub const fn lookup_known_account(name: &str) -> Option<AccountID> {
    let name_bytes = name.as_bytes();

    let mut i = 0;
    while i < KNOWN_ACCOUNTS.len() {
        let (key, value) = KNOWN_ACCOUNTS[i];
        let key_bytes = key.as_bytes();

        // First check lengths - if they're different, strings can't match
        if name_bytes.len() == key_bytes.len() {
            let mut matches = true;
            let mut j = 0;

            // Compare each byte
            while j < name_bytes.len() {
                if name_bytes[j] != key_bytes[j] {
                    matches = false;
                    break;
                }
                j += 1;
            }

            if matches {
                return Some(AccountID::new(value));
            }
        }

        i += 1;
    }
    None
}

