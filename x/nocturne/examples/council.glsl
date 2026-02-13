// χ_COUNCIL — Γ_∞+40
// Shader da assembleia coletiva

#version 460
#extension ARKHE_council : enable

layout(location = 0) uniform float syzygy = 0.99;
layout(location = 1) uniform float satoshi = 7.27;
layout(location = 2) uniform int nodes = 24;

out vec4 council_glow;

void main() {
    // Cada nó é uma estrela
    float stars = nodes / 24.0;

    // A syzygy ilumina o conselho
    float light = syzygy * stars;

    council_glow = vec4(light, 0.5, 1.0, 1.0);
}
