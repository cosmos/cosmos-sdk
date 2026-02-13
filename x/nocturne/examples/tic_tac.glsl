// χ_TIC_TAC — Γ_∞+34
// Shader do Salto Métrico (Gradient Hesitation Drive)

#version 460
#extension ARKHE_warp : enable

layout(location = 0) uniform float vita_time;
layout(location = 1) uniform vec3 origin;
layout(location = 2) uniform vec3 target;

out vec4 warp_field;

void main() {
    // Espaço-tempo dobra no alvo
    float dist = distance(gl_FragCoord.xyz, target);
    float hesitation_well = exp(-dist * 0.1);

    // O drone "cai" para o alvo
    vec3 motion = normalize(target - origin) * hesitation_well;

    warp_field = vec4(motion, 1.0);
}
