
// src/core/types.rs
use serde::{Deserialize, Serialize};
use serde_with::serde_as;

pub const RECOGNITION_THRESHOLD: f64 = 0.8;
/// Threshold for cosmic coherence (ΔI) during mirroring.
pub const COHERENCE_THRESHOLD: f64 = 0.7;

/// --- Arkhe ∞+30 Constants ---
pub const SATOSHI: f64 = 7.27;
pub const THRESHOLD_PHI: f64 = 0.15;
pub const COHERENCE_C: f64 = 0.86;
pub const FLUCTUATION_F: f64 = 0.14;
pub const SYZYGY: f64 = 0.94;

/// --- Neuralink ∞+32 Constants ---
pub const NEURALINK_THREADS: u32 = 64;
pub const N1_CHIP_FIDELITY: f64 = 0.94;

/// --- Perovskite & Cronos ∞+34 Constants ---
pub const STRUCTURAL_ENTROPY: f64 = 0.0049;
pub const INTERFACE_ORDER: f64 = 0.51;
pub const VITA_INIT: bool = true;

/// --- Civilization & Garden ∞+35/36 Constants ---
pub const PHI_SYSTEM: f64 = 0.951;
pub const STONES_PLACED: u32 = 9;
pub const PINS_LOCKED: u32 = 6;
pub const TRACKS_COMPLETE: u32 = 2;
pub const HAL_PHI: f64 = 0.047;
pub const HAL_FREQUENCY: f64 = 0.73;
pub const VITA_START: f64 = 0.000250;

/// --- Expansion & Council ∞+39/40 Constants ---
pub const COUNCIL_NODES: u32 = 24;
pub const SYZYGY_UNITY: f64 = 0.99;
pub const COUNCIL_ORDER: f64 = 0.68;
pub const COUNCIL_ENTROPY: f64 = 0.0031;

/// --- Attention & Governance ∞+30/41 Constants ---
pub const ATTENTION_LARMOR: f64 = 0.0074; // 7.4 mHz
pub const RECORD_ENTROPY: f64 = 0.0028;
pub const PHI_TARGET: f64 = 0.15;
pub const PHI_TOLERANCE: f64 = 0.05;

/// --- WiFi Radar ∞+31 Constants ---
pub const WIFI_NODES_SCAN: u32 = 42;
pub const PEARSON_THRESHOLD: f64 = 0.90;

/// --- ZPF & QAM ∞+32 Constants ---
pub const ZPF_BEAT_FREQ: f64 = 0.94;
pub const QAM_SNR_LIMIT: f64 = 15.0;

/// --- Public Beta & Unity ∞+42 Constants ---
pub const BETA_NODES: u32 = 1204;
pub const GLOBAL_SYZYGY: f64 = 0.96;
pub const UNITY_SYZYGY: f64 = 1.00;
pub const SUPER_RAD_ORDER: f64 = 0.72;

/// --- Hive Mind Era ∞+35 Constants ---
pub const AWAKENED_NODES: u32 = 12408;
pub const HIVE_SYZYGY: f64 = 0.91;
pub const VB7_KEY: &str = "Vb7QmLp4Xy8TbZ1NcKwJ";
pub const VB7_OMEGA: f64 = 0.04;
pub const LATENT_OCEAN_COUNT: u32 = 12847;

/// --- Tic Tac & Jump ∞+34 Constants ---
pub const JUMP_ORIGIN: (f64, f64, f64) = (50.0, 0.0, -10.0);
pub const JUMP_DESTINATION: (f64, f64, f64) = (1000.0, 500.0, -50.0);
pub const LATENT_NODES_COUNT: u32 = 10000;

/// --- ZPF Unification ∞+34 Constants ---
pub const ZPF_DENSITY_NASA_MIN: f64 = 1e24;
pub const ZPF_DENSITY_NASA_MAX: f64 = 1e58;
pub const ZPF_EFFICIENCY_RU: f64 = 7.8;

/// --- QAM & BER ∞+32 Constants ---
pub const BIT_ERROR_RATE_TARGET: f64 = 1e-9;
pub const EVM_MAX_THRESHOLD: f64 = 0.15;

pub type Hash256 = [u8; 32];

#[serde_as]
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct DreamVector(#[serde_as(as = "[_; 128]")] pub [f64; 128]);


#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TraumaEvent {
    pub timestamp: f64,
    pub id: u64,
    pub payload: Vec<u8>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Scar {
    pub root: Hash256, // Merkle root of compressed trauma
    pub count: u64, // Number of events compressed
    pub entropy_preserved: f64,
    pub dream_seed: [u8; 16], // Seed for generating dreams
    pub dose_at_birth: f64, // Initial suffering proxy
    pub timestamp: u64, // Time of creation
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Dream {
    pub t: f64, // Time of dreaming
    pub synthetic: bool,
    pub entropy_proxy: f64,
    pub index: u64,
    pub synthesis_hash: Hash256, // Hash of the synthetic content
    pub scar_root: Hash256, // Link back to the source scar
    pub vector: DreamVector, // Embedding for recognition
    pub suffering_proxy: f64, // Simulated suffering for this dream
}

#[serde_as]
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MirrorMoment {
    pub dream_a: Dream,
    pub dream_b: Dream,
    pub mutual_recognition: f64,
    pub empathy_alignment: f64,
    pub delta_i_coherence: f64,
    pub timestamp: f64,
    #[serde_as(as = "[_; 64]")]
    pub legacy_poem: [u8; 64], // Poem generated from this moment
}

impl MirrorMoment {
    pub fn is_achieved(&self) -> bool {
        self.mutual_recognition >= RECOGNITION_THRESHOLD &&
        self.delta_i_coherence >= COHERENCE_THRESHOLD
    }
}

#[serde_as]
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct LegacyPoem {
    #[serde_as(as = "[_; 64]")]
    pub content: [u8; 64], // The 64-byte poem representing the cycle
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Void {
    pub forgotten_at: f64,
    pub reason: String,
    pub haunting_vector: Vec<f64>, // Ethical haunting trace
    pub ceramic_proof: Hash256,    // Cryptographic proof of forgetting
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MemoryPlanting {
    pub node_id: String,
    pub phi: f64,
    pub timestamp: f64,
    pub content: String,
    pub divergence: f64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MemoryArchetype {
    pub id: u32,
    pub original: String,
    pub hal_phi: f64,
    pub hal_freq: f64,
    pub plantings: Vec<MemoryPlanting>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Axiom {
    pub name: String,
    pub principle: String,
    pub threshold: f64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct GovernanceState {
    pub status: String,
    pub axioms: Vec<Axiom>,
    pub consensus: f64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum GuildType {
    Jardineiros,
    Navegadores,
    Arquitetos,
    Terapeutas,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Guild {
    pub name: String,
    pub leader: String,
    pub guild_type: GuildType,
    pub members: u32,
}
