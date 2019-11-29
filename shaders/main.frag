#version 410 core

uniform sampler2D tex;
in vec2 fragTexCoord;
in float height;
out vec4 color;

void main() {
    color = vec4(vec3(height), 1.0);
}