#version 410 core

uniform mat4 projection;
uniform mat4 camera;
uniform mat4 model;
uniform float angle;
uniform float height;
uniform vec3 hitpos;

layout (location = 0) in vec3 vert;
layout (location = 1) in vec3 normal;
layout (location = 2) in vec2 texcoord;

out vec2 fragTexCoord;
out vec3 vertex;

void main() {
    fragTexCoord = texcoord;
    vertex = vert;
    gl_Position = projection * camera * vec4(vec3(vert.x, height * vert.y, vert.z), 1.0);
}