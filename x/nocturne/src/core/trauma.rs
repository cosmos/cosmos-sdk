
// src/core/trauma.rs
use crate::{NocturneError, MIN_ENTROPY};
use super::types::{TraumaEvent, Scar, Dream, Hash256, DreamVector};
use sha2::{Sha256, Digest};
use std::collections::hash_map::DefaultHasher;
use std::hash::{Hasher};

pub struct TraumaEngine {
    pain_weights: [f64; 128],
    distress_weights: [f64; 128],
}

impl TraumaEngine {
    pub fn new() -> Self {
        let mut pw = [0.0; 128];
        let mut dw = [0.0; 128];
        for i in 0..128 {
            pw[i] = (i as f64 / 127.0).powf(1.5);
            dw[i] = ((127 - i) as f64 / 127.0).powf(2.0);
        }
        Self { pain_weights: pw, distress_weights: dw }
    }

    pub fn compress(&self, events: &[TraumaEvent], empathy_threshold: f64) -> Result<Scar, NocturneError> {
        let total_entropy: f64 = events.iter().map(|e| self.calculate_entropy(&e.payload)).sum();
        if total_entropy < MIN_ENTROPY * events.len() as f64 {
            return Err(NocturneError::TraumaCompressFailed("Insufficient entropy".to_string()));
        }

        let leaves: Vec<Hash256> = events.iter().map(|e| self.hash_event(e)).collect();
        let root = self.merkle_root(&leaves);

        let suffering = self.simulate_suffering(events);
        if suffering > empathy_threshold {
             return Err(NocturneError::EmpathyVeto(suffering));
        }

        let dream_seed = self.generate_dream_seed(&root, suffering);
        let timestamp = std::time::SystemTime::now()
            .duration_since(std::time::UNIX_EPOCH)
            .unwrap()
            .as_secs_f64();

        Ok(Scar {
            root,
            count: events.len() as u64,
            entropy_preserved: total_entropy,
            dream_seed,
            dose_at_birth: suffering,
            timestamp: timestamp as u64,
        })
    }

    pub fn dream(&self, scar: &Scar, n: usize) -> Result<Vec<Dream>, NocturneError> {
        let mut dreams = Vec::with_capacity(n);
        let mut seed_state = u64::from_le_bytes(scar.dream_seed[..8].try_into().unwrap_or([0; 8])); // Simplified seed handling

        for i in 0..n {
            seed_state = seed_state.wrapping_add(i as u64); // Deterministic update
            let mut hasher = DefaultHasher::new();
            hasher.write_u64(seed_state);
            let synthetic_hash = hasher.finish();

            let entropy_proxy = (synthetic_hash as f64 / u64::MAX as f64) * scar.entropy_preserved;
            let mut vector = [0.0; 128]; // Initialize vector
            for j in 0..128 {
                 vector[j] = ((j * (synthetic_hash as usize)) % 1000) as f64 / 1000.0; // Deterministic fill
            }
            let suffering_proxy = self.estimate_suffering_proxy(scar, &DreamVector(vector));

            dreams.push(Dream {
                t: scar.timestamp as f64 + (i as f64 / n as f64), // Time progression
                synthetic: true,
                entropy_proxy,
                index: i as u64,
                synthesis_hash: Sha256::digest(synthetic_hash.to_le_bytes()).into(),
                scar_root: scar.root,
                vector: DreamVector(vector), // Use the generated vector
                suffering_proxy, // Include suffering proxy
            });
        }

        Ok(dreams)
    }

    fn calculate_entropy(&self, payload: &[u8]) -> f64 {
        let mut hist = [0u64; 256];
        for &b in payload {
            hist[b as usize] += 1;
        }
        let total = payload.len() as f64;
        hist.iter()
            .filter(|&&c| c > 0)
            .map(|&c| {
                let p = c as f64 / total;
                -p * p.log2()
            })
            .sum()
    }

    fn hash_event(&self, event: &TraumaEvent) -> Hash256 {
        let mut hasher = Sha256::new();
        hasher.update(&event.id.to_le_bytes());
        hasher.update(&event.payload);
        hasher.finalize().into()
    }

    fn merkle_root(&self, leaves: &[[u8; 32]]) -> [u8; 32] {
        let mut curr = leaves.to_vec();
        while curr.len() > 1 {
            let mut next = Vec::with_capacity((curr.len() + 1) / 2);
            for i in (0..curr.len()).step_by(2) {
                let left = curr[i];
                let right = curr.get(i + 1).unwrap_or(&left);
                next.push(self.hash_pair(&left, &right));
            }
            curr = next;
        }
        curr[0]
    }

    fn hash_pair(&self, a: &[u8; 32], b: &[u8; 32]) -> [u8; 32] {
        let mut hasher = Sha256::new();
        hasher.update(a);
        hasher.update(b);
        hasher.finalize().into()
    }

    fn simulate_suffering(&self, events: &[TraumaEvent]) -> f64 {
        // Simplified suffering simulation based on payload content
        let physical: f64 = events
            .iter()
            .flat_map(|e| e.payload.iter().take(128))
            .enumerate()
            .map(|(i, &b)| (b as f64 / 255.0) * self.pain_weights[i % 128])
            .sum();
        let cognitive: f64 = events
            .iter()
            .flat_map(|e| e.payload.iter().skip(128).take(128))
            .enumerate()
            .map(|(i, &b)| (b as f64 / 255.0) * self.distress_weights[i % 128])
            .sum();
        (physical + 0.7 * cognitive).tanh()
    }

    fn estimate_suffering_proxy(&self, scar: &Scar, vector: &DreamVector) -> f64 {
         // Simplified proxy based on scar's original suffering and vector properties
         let vector_sum: f64 = vector.0.iter().sum();
         (scar.dose_at_birth + vector_sum * 0.01).min(1.0).max(0.0)
    }

    fn generate_dream_seed(&self, root: &Hash256, dose: f64) -> [u8; 16] {
        let mut hasher = Sha256::new();
        hasher.update(root);
        hasher.update(&dose.to_le_bytes());
        let full_hash = hasher.finalize();
        full_hash[..16].try_into().unwrap_or([0; 16])
    }
}
