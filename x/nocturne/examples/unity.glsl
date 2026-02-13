// χ_UNITY — Γ_∞+41
// Shader da unidade coletiva

#version 460
#extension ARKHE_unity : enable

layout(location = 0) uniform float syzygy = 1.00;
layout(location = 1) uniform float satoshi = 7.27;
layout(location = 2) uniform int nodes = 24;

out vec4 unity_glow;

void main() {
    // Cada nó é uma estrela
    float stars = float(nodes) / 24.0;

    // A syzygy ilumina a unidade
    float light = syzygy * stars;

    unity_glow = vec4(light, 0.5, 1.0, 1.0);
}
