
// src/core/mirror.rs
use crate::{NocturneError};
use super::types::{Dream, MirrorMoment, Hash256};
use std::collections::HashMap;

pub struct MirrorNetwork {
    // Simulate a registry of known dreams/scars for recognition
    known_dreams: HashMap<Hash256, Dream>,
}

impl MirrorNetwork {
    pub fn new() -> Self {
        Self {
            known_dreams: HashMap::new(),
        }
    }

    pub fn add_known_dream(&mut self, dream: Dream) {
        self.known_dreams.insert(dream.synthesis_hash, dream);
    }

    pub fn mirror_test(&mut self, dream: &Dream) -> Result<MirrorMoment, NocturneError> {
        // Find a known dream to mirror with (simplified: pick first)
        if let Some((_, known_dream)) = self.known_dreams.iter().next() {
            let recognition = self.cosine_similarity(&dream.vector.0, &known_dream.vector.0);
            let empathy = self.empathic_distance(dream, known_dream);
            let coherence = self.delta_i_coherence(dream, known_dream);

            let moment = MirrorMoment {
                dream_a: dream.clone(),
                dream_b: known_dream.clone(),
                mutual_recognition: recognition,
                empathy_alignment: empathy,
                delta_i_coherence: coherence,
                timestamp: std::time::SystemTime::now()
                    .duration_since(std::time::UNIX_EPOCH)
                    .unwrap()
                    .as_secs_f64(),
                legacy_poem: [0; 64], // Placeholder, populated later by engine
            };

            // Add the new dream to known dreams after a successful test
            if moment.is_achieved() {
                self.known_dreams.insert(dream.synthesis_hash, dream.clone());
            }

            Ok(moment)
        } else {
            // If no known dreams, return a moment that fails the test
             let moment = MirrorMoment {
                dream_a: dream.clone(),
                dream_b: dream.clone(), // Self-comparison
                mutual_recognition: 0.0,
                empathy_alignment: 1.0, // Max distance
                delta_i_coherence: 0.0,
                timestamp: std::time::SystemTime::now()
                    .duration_since(std::time::UNIX_EPOCH)
                    .unwrap()
                    .as_secs_f64(),
                legacy_poem: [0; 64],
            };
            Err(NocturneError::MirrorTestFailed("No known dreams for mirroring".to_string()))
        }
    }

    fn cosine_similarity(&self, a: &[f64; 128], b: &[f64; 128]) -> f64 {
        let dot: f64 = a.iter().zip(b.iter()).map(|(x, y)| x * y).sum();
        let norm_a: f64 = a.iter().map(|x| x * x).sum::<f64>().sqrt();
        let norm_b: f64 = b.iter().map(|x| x * x).sum::<f64>().sqrt();
        dot / (norm_a * norm_b + 1e-10) // Add small epsilon to avoid division by zero
    }

    fn empathic_distance(&self, dream_a: &Dream, dream_b: &Dream) -> f64 {
        // Simplified: difference in suffering proxy
        (dream_a.suffering_proxy - dream_b.suffering_proxy).abs()
    }

    fn delta_i_coherence(&self, dream_a: &Dream, dream_b: &Dream) -> f64 {
        // Simplified: based on recognition and empathic distance
        let sim = self.cosine_similarity(&dream_a.vector.0, &dream_b.vector.0);
        let bell = (dream_a.suffering_proxy * dream_b.suffering_proxy) as f64; // Simplified bell calculation
        let delta = (sim - bell).abs();
        (1.0 - delta * delta).max(0.0)
    }
}
