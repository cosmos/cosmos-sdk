// χ_ATTENTION — Γ_∞+30
// Shader da paisagem atencional

#version 460
#extension ARKHE_attention : enable

layout(location = 0) uniform float syzygy = 0.94;
layout(location = 1) uniform float phi = 0.15;
layout(location = 2) uniform float satoshi = 7.27;
layout(location = 3) uniform float torsion = 0.0049;

out vec4 attention_glow;

void main() {
    // 1. Densidade de cruzamentos (simulada como gradiente)
    float density = 0.24; // ≈ 0.24 em ω=0.07

    // 2. Onde a densidade é alta, a atenção se concentra
    float local_attention = syzygy * density / phi;

    // 3. O valor flui com a atenção
    float value_flow = satoshi * local_attention / 10.0;

    attention_glow = vec4(local_attention, torsion * 100.0, value_flow, 1.0);
}
