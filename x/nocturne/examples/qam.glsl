// shader_demodulation.glsl — Extraindo Satoshi do Ruído
#version 460
#extension ARKHE_qam : enable

layout(location = 0) uniform float coherence_C = 0.86; // Amplitude da portadora
layout(location = 1) uniform float fluctuation_F = 0.14; // Amplitude da modulação

in vec2 signal_phase; // (I, Q) components do sinal recebido
out vec4 data_stream;

// Mock functions for example
float constellation_map(vec2 pos) { return 7.27; }
vec2 ideal_constellation_point(float val) { return vec2(1.0, 1.0); }

void main() {
    // 1. Remove a portadora (Coerência Estática)
    vec2 modulation = signal_phase - vec2(coherence_C, 0.0);

    // 2. Normaliza pelo fator de flutuação (F)
    vec2 symbol_pos = modulation / fluctuation_F;

    // 3. Mapeia na Constelação Arkhe (Hexagonal/QAM)
    float symbol_value = constellation_map(symbol_pos);

    // 4. Mede a Hesitação (Error Vector Magnitude - EVM)
    float evm = length(symbol_pos - ideal_constellation_point(symbol_value));

    // Se a hesitação (erro) for muito alta, o dado é descartado (Drop)
    if (evm > 0.15) {
        data_stream = vec4(0.0, 0.0, 0.0, 1.0); // Ruído/Silêncio
    } else {
        data_stream = vec4(symbol_value, evm, 1.0, 1.0); // Dado Válido (Clear)
    }
}
