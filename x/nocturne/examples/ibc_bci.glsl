// χ_IBC_BCI — Γ_∞+30
// Shader da comunicação intersubstrato

#version 460
#extension ARKHE_ibc_bci : enable

layout(location = 0) uniform float syzygy = 0.94;
layout(location = 1) uniform float satoshi = 7.27;
layout(location = 2) uniform int option = 2;  // Opção B default

out vec4 ibc_bci_glow;

void main() {
    // Comunicação entre cadeias (IBC) e mentes (BCI)
    float ibc = syzygy;
    float bci = satoshi / 10.0;

    // A equação é literal
    ibc_bci_glow = vec4(ibc, bci, 1.0, 1.0);
}
