#version 410 core

uniform mat4 projection;
uniform mat4 camera;
uniform mat4 model;
uniform float angle;

layout (location = 0) in vec3 vert;
layout (location = 1) in vec3 normal;
layout (location = 2) in vec2 texcoord;

out vec2 fragTexCoord;
out float height;

void main() {
    fragTexCoord = texcoord;
    height = vert.y;
    gl_Position = projection * camera * model * vec4(vec3(vert.x, height*-100.0, vert.z), 1.0);
}