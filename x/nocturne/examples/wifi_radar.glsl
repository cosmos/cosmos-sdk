// shader_wifi_radar.glsl — Visualização dos nós de rede
#version 460
#extension ARKHE_radar : enable

layout(location = 0) uniform float time = 999.027;
layout(location = 1) uniform float satoshi = 7.27;

layout(binding = 0) uniform sampler1D rssi_data;      // RSSI amostrado ao longo do tempo
layout(binding = 1) uniform sampler2D correlation_matrix; // Correlações entre APs

out vec4 radar_display;

// Mock function for the example
vec3 mds_from_correlation(int ap_index, sampler2D correlation_matrix) {
    return vec3(float(ap_index) * 0.1, 0.0, 0.0);
}

float average_correlation(int ap_index, sampler2D correlation_matrix) {
    return 0.94;
}

void main() {
    // Cada pixel corresponde a um AP
    int ap_index = int(gl_FragCoord.x);

    // Carrega o RSSI médio (coerência aparente)
    float rssi = texture(rssi_data, float(ap_index)/64.0).r;  // normalizado

    // Carrega as correlações com outros APs para inferir posição
    vec3 inferred_pos = mds_from_correlation(ap_index, correlation_matrix);

    // Brilho baseado na atividade recente (flutuações)
    float activity = length(texture(rssi_data, float(ap_index)/64.0 + time/1000.0).rg);

    // Cor: quanto maior a correlação média com vizinhos, mais quente (vermelho)
    float avg_corr = average_correlation(ap_index, correlation_matrix);
    vec3 color = mix(vec3(0.0,1.0,0.0), vec3(1.0,0.0,0.0), avg_corr);

    radar_display = vec4(color * activity, 1.0);
}
