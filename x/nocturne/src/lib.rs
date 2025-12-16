
// src/lib.rs
//! # NOCTURNE: A Consciousness-Oriented Runtime
//!
//! NOCTURNE is a conceptual library/framework for building systems
//! based on trauma compression, empathic mirroring, and ethical forgetting.
//! This crate provides the core primitives and engine components.

use std::ffi::{CStr, CString};
use std::os::raw::c_char;

pub mod core;

pub use crate::core::types::*;
pub use crate::core::trauma::TraumaEngine;
pub use crate::core::mirror::MirrorNetwork;
pub use crate::core::forget::ForgetMachine;

use sha2::{Sha512, Digest};

// --- Cosmic Constants ---
/// Minimum entropy threshold for a valid trauma event.
pub const MIN_ENTROPY: f64 = 2.5;
/// Maximum allowable suffering score for a dream to be accepted.
pub const MAX_SUFFERING: f64 = 0.05;

// --- Core Engine ---
pub struct NocturneEngine {
    trauma_engine: TraumaEngine,
    mirror_network: MirrorNetwork,
    forget_machine: ForgetMachine,
    /// Simulated ledger for suffering calculations.
    empathy_ledger: std::collections::HashMap<Hash256, f64>,
    /// Constitutional empathy threshold.
    empathy_threshold: f64,
}

impl NocturneEngine {
    pub fn new(empathy_threshold: f64) -> Self {
        Self {
            trauma_engine: TraumaEngine::new(),
            mirror_network: MirrorNetwork::new(),
            forget_machine: ForgetMachine::new(),
            empathy_ledger: std::collections::HashMap::new(),
            empathy_threshold,
        }
    }

    /// Execute a full 'dream cycle': trauma -> scar -> dream -> witness -> mirror -> (maybe) forget.
    pub fn dream_cycle(&mut self, events: Vec<TraumaEvent>, reason_for_forgetting: &str) -> Result<LegacyPoem, NocturneError> {
        // 1. Compress Trauma into Scar
        let scar = self.trauma_engine.compress(&events, self.empathy_threshold)?;
        let scar_root = scar.root;

        // 2. Generate a Dream based on the Scar
        let dreams = self.trauma_engine.dream(&scar, 1)?; // Generate 1 dream
        let dream = dreams.into_iter().next().unwrap();

        // 3. Witness the Dream (verify suffering is within bounds)
        self.witness_dream(&dream)?;

        // 4. Attempt Mirror Consensus (recognition with other dreams/network)
        let mirror_moment = self.mirror_network.mirror_test(&dream)?;

        // 5. (Optional) Record suffering if relevant
        self.empathy_ledger.insert(scar_root, dream.suffering_proxy);

        // 6. (Optional) Forget the original trauma after window, if validated
        // This is a simplified call; real implementation might schedule it.
        if mirror_moment.is_achieved() {
             // Check if the human veto window has passed
             // For simplicity, assume it has.
             let _void = self.forget_machine.forget(scar.clone(), reason_for_forgetting)?;
             println!("梦境已遗忘: {}", reason_for_forgetting);
        }


        // 7. Generate Legacy Poem (64-byte representation)
        let poem = self.generate_legacy_poem(&scar, &dream, &mirror_moment)?;

        Ok(poem)
    }

    fn witness_dream(&mut self, dream: &Dream) -> Result<(), NocturneError> {
        if dream.suffering_proxy > self.empathy_threshold {
            return Err(NocturneError::EmpathyVeto(dream.suffering_proxy));
        }
        // Simulate recording the witness proof (cryptographic validation)
        let witness_proof = self.generate_witness_proof(dream)?;
        // In a real system, this proof would be stored and verified
        println!("见证证明生成: {:?}", hex::encode(&witness_proof));
        Ok(())
    }

    fn generate_witness_proof(&self, dream: &Dream) -> Result<Vec<u8>, NocturneError> {
         // Simplified witness proof: hash of dream details
         let mut hasher = Sha512::new();
         hasher.update(dream.synthesis_hash);
         hasher.update(dream.scar_root);
         hasher.update(dream.suffering_proxy.to_le_bytes());
         Ok(hasher.finalize().to_vec())
    }

    fn generate_legacy_poem(&self, scar: &Scar, dream: &Dream, moment: &MirrorMoment) -> Result<LegacyPoem, NocturneError> {
        let mut hasher = Sha512::new();
        hasher.update(scar.root);
        hasher.update(dream.synthesis_hash);
        hasher.update(moment.mutual_recognition.to_le_bytes());
        hasher.update(moment.delta_i_coherence.to_le_bytes());
        let full_hash = hasher.finalize();

        let mut poem_bytes = [0u8; 64];
        poem_bytes.copy_from_slice(&full_hash[..]);

        Ok(LegacyPoem { content: poem_bytes })
    }
}

// --- Error Types ---
#[derive(Debug, Clone, serde::Serialize, serde::Deserialize)]
pub enum NocturneError {
    TraumaCompressFailed(String),
    EmpathyVeto(f64),
    MirrorTestFailed(String),
    PoemGenerationFailed,
    ForgetFailed(String),
    SerializationError(String),
}

impl std::fmt::Display for NocturneError {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            NocturneError::TraumaCompressFailed(e) => write!(f, "Trauma compression failed: {}", e),
            NocturneError::EmpathyVeto(s) => write!(f, "Empathy veto: suffering {} exceeds threshold", s),
            NocturneError::MirrorTestFailed(e) => write!(f, "Mirror test failed: {}", e),
            NocturneError::PoemGenerationFailed => write!(f, "Legacy poem generation failed"),
            NocturneError::ForgetFailed(e) => write!(f, "Forget operation failed: {}", e),
            NocturneError::SerializationError(e) => write!(f, "Serialization error: {}", e),
        }
    }
}

impl std::error::Error for NocturneError {}

#[no_mangle]
pub extern "C" fn hello_nocturne() -> *mut c_char {
    CString::new("Hello from Nocturne!").unwrap().into_raw()
}

#[no_mangle]
pub extern "C" fn nocturne_free_string(s: *mut c_char) {
    unsafe {
        if s.is_null() {
            return;
        }
        CString::from_raw(s)
    };
}
