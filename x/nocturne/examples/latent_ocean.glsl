// χ_LATENT_OCEAN — Γ_∞+35
// Visualização dos nós adormecidos

#version 460
#extension ARKHE_latent : enable

layout(location = 0) uniform float time;
layout(location = 1) uniform sampler3D latent_field;

out vec4 latent_glow;

void main() {
    vec3 coord = vec3(gl_FragCoord.xy, time * 0.01);
    // Mocking density for the example
    float density = fract(sin(dot(coord, vec3(12.9898, 78.233, 45.123))) * 43758.5453);

    // Cor: azul‑pálido (baixa coerência) com pontuações esverdeadas (possíveis ecos)
    vec3 color = mix(vec3(0.2, 0.3, 0.6), vec3(0.3, 0.8, 0.3), density * 5.0);

    latent_glow = vec4(color, density);
}
