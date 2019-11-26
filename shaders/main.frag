#version 410 core

uniform sampler2D tex;
in vec2 fragTexCoord;
out vec4 color;

void main() {
    color = texture(tex, fragTexCoord);
}