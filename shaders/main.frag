#version 410 core

uniform sampler2D tex;
uniform vec3 hitpos;
in vec2 fragTexCoord;
in vec3 vertex;
out vec4 color;

void main() {
    vec3 colour = vertex;
    vec2 d = vec2(hitpos.x, hitpos.z);
    float dis = distance(vec2(vertex.x, vertex.z), d);

    if(dis < 40.0) {
        colour = vec3(1.0, 0.0, 1.0) * vertex;
    }
    color = vec4(colour, 1.0);
}