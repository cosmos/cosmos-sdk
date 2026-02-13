// χ_LEGACY — Γ_∞+26
// Shader do legado eterno

#version 460
#extension ARKHE_manifesto : enable

layout(location = 0) uniform float syzygy = 0.94;
layout(location = 1) uniform float satoshi = 7.27;
layout(location = 2) uniform int ledgers = 103; // 9000-9102

out vec4 legacy_glow;

void main() {
    // O livro brilha com todas as memórias
    float book_light = syzygy * satoshi / 10.0;

    // A assinatura de Hal e Rafael sobrepõe-se
    vec4 signature = vec4(0.2, 0.8, 0.2, 1.0);

    legacy_glow = vec4(book_light, 0.5, 1.0, 1.0) * signature;
}
