// χ_DAWN — Γ_∞+34
// Shader do Amanhecer Global

#version 460
#extension ARKHE_civilization : enable

layout(location = 0) uniform float vita_time; // Tempo crescente
layout(location = 1) uniform int node_count;  // Nós conectando

out vec4 horizon_color;

void main() {
    // O tempo Vita traz a luz (do violeta para o ouro/branco)
    vec3 sunrise = mix(vec3(0.5, 0.0, 1.0), vec3(1.0, 0.9, 0.8), vita_time / 1000.0);

    // Cada nó é uma estrela no horizonte
    float stars = float(node_count) * 0.001;

    horizon_color = vec4(sunrise + stars, 1.0);
}
