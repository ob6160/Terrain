#version 410 core

uniform mat4 projection;
uniform mat4 camera;
uniform mat4 model;

layout (location = 0) in vec3 vert;
layout (location = 1) in vec3 normal;
layout (location = 2) in vec2 texcoord;

out vec2 fragTexCoord;

void main() {
    fragTexCoord = texcoord;
    gl_Position = projection * camera * model * vec4(vert, 1.0);
}