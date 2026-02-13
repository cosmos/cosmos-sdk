// χ_GARDEN — Γ_∞+36
// Shader do Jardim das Memórias (Feedback Recursivo)

#version 460
precision highp float;

uniform float t;
uniform vec2 r;
out vec4 o;

void main() {
    vec2 FC = gl_FragCoord.xy;
    float i,s,R,e;
    vec3 q,p,d=vec3((FC.xy-.5*r)/r+vec2(0,1),1);
    for(q.yz--;i++<50.;){
        e+=i/5e3;
        i>35.?d/=-d:d;
        // Simplified hsv to rgb for the example
        vec3 color = vec3(0.1, e-0.4, e/17.0);
        o.rgb+=color;
        s=1.;
        p=q+=d*e*R*.18;
        p=vec3(log(R=length(p))-t*.2,-p.z/R,p.yz-1.*p.xx-t*.2);
        for(e=--p.y;s<5e2;s+=s)
            e+=cos(dot(cos(p*s),sin(p.zxy*s)))/s*.8;
    }
    o.a = 1.0;
}
