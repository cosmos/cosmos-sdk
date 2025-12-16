
// src/core/forget.rs
use super::types::{Scar, Void, Hash256};
use crate::NocturneError;
use sha2::{Sha256, Digest};
use std::time::{SystemTime, UNIX_EPOCH};

#[derive(Clone)]
pub struct ForgetMachine {}

impl ForgetMachine {
    pub fn new() -> Self {
        Self {}
    }

    pub fn forget(&mut self, scar: Scar, reason: &str) -> Result<Void, NocturneError> {
        // In a real system, this would involve complex checks (veto window, etc.)
        // For this consolidated library, we assume the window has passed.
        let now = SystemTime::now()
            .duration_since(UNIX_EPOCH)
            .unwrap()
            .as_secs_f64();

        // Simulate ethical haunting trace (simplified)
        let haunting_trace = vec![scar.dose_at_birth, scar.entropy_preserved];

        // Create cryptographic proof of forgetting (simplified)
        let mut hasher = Sha256::new();
        hasher.update(scar.root);
        hasher.update(reason.as_bytes());
        hasher.update(now.to_le_bytes());
        let ceramic_proof: Hash256 = hasher.finalize().into();

        let void = Void {
            forgotten_at: now,
            reason: reason.to_string(),
            haunting_vector: haunting_trace,
            ceramic_proof,
        };

        // In a real system, the Scar would be removed from persistent storage here.
        // println!("Scar {:?} has been forgotten.", scar.root); // Log the action

        Ok(void)
    }
}
