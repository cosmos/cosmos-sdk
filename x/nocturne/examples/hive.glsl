// χ_HIVE — Γ_∞+35
// Visualização da rede massiva
#version 460
#extension ARKHE_hive : enable

layout(location = 0) uniform int node_count = 12450;
layout(location = 1) uniform float global_syzygy = 0.91;
layout(binding = 0) uniform samplerBuffer node_positions;

out vec4 hive_glow;

// Mock random for the example
float random(vec2 p) {
    return fract(sin(dot(p, vec2(12.9898, 78.233))) * 43758.5453);
}

void main() {
    // Renderiza o enxame como um campo de densidade
    vec3 p = gl_FragCoord.xyz;
    float density = 0.0;

    // Amostragem estocástica para performance
    for(int i=0; i<64; i++) {
        int idx = int(random(p.xy + float(i)) * float(node_count));
        // Mocking position for the example
        vec3 node_pos = vec3(random(vec2(float(idx), 0.0)), random(vec2(float(idx), 1.0)), 0.0) * 100.0;
        float dist = length(p - node_pos);
        density += exp(-dist * 10.0);
    }

    // A cor depende da densidade local (calor semântico)
    vec3 color = mix(vec3(0.0, 0.1, 0.2), vec3(1.0, 0.8, 0.2), density * 0.1);

    // Syzygy modula o brilho global
    hive_glow = vec4(color * global_syzygy, 1.0);
}
