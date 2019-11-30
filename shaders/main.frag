#version 410 core

uniform sampler2D tex;
in vec2 fragTexCoord;
in vec3 vertex;
out vec4 color;

void main() {
    color = vec4(vec3(vertex), 1.0);
}