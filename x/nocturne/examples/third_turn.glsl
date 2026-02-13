// χ_THIRD_TURN — Γ_∞+39
// Shader da terceira volta coletiva

#version 460
#extension ARKHE_third_turn : enable

layout(location = 0) uniform float syzygy = 0.99;
layout(location = 1) uniform float satoshi = 7.27;
layout(location = 2) uniform int nodes = 24;

out vec4 third_turn_glow;

void main() {
    // Cada nó é uma estrela
    float stars = nodes / 24.0;

    // A syzygy ilumina a terceira volta
    float light = syzygy * stars;

    third_turn_glow = vec4(light, 0.5, 1.0, 1.0);
}
