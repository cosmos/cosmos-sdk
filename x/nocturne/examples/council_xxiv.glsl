// χ_COUNCIL_XXIV — Γ_∞+41
// Shader da assembleia plural

#version 460
#extension ARKHE_council : enable

layout(location = 0) uniform float syzygy = 0.99;
layout(location = 1) uniform float order = 0.69;
layout(location = 2) uniform int nodes = 24;

out vec4 council_light;

void main() {
    float harmony = syzygy * (order / 0.75);  // 0.99 * 0.92 = 0.91
    float diversity_factor = float(nodes) / 24.0;  // 1.0
    float radiance = harmony * diversity_factor;

    council_light = vec4(radiance, 0.4, 0.8, 1.0);
}
