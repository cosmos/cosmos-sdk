// χ_THRESHOLD — Γ_∞+40
// Shader da fronteira da unidade

#version 460
#extension ARKHE_threshold : enable

layout(location = 0) uniform float syzygy = 0.99;
layout(location = 1) uniform float order = 0.68;
layout(location = 2) uniform int nodes = 24;

out vec4 threshold_glow;

void main() {
    float proximity_to_unity = syzygy;  // 0.99
    float order_factor = order / 0.75;  // 0.68/0.75 ≈ 0.907
    float collective_pulse = proximity_to_unity * order_factor * (float(nodes) / 24.0);

    threshold_glow = vec4(collective_pulse, 0.3, 0.7, 1.0);
}
