#version 430 core

uniform mat4 projection;
uniform mat4 camera;
uniform mat4 model;
uniform float angle;
uniform float height;
uniform vec3 hitpos;

layout (rgba32f, binding = 0) readonly uniform highp image2D nextHeightTex;

layout (location = 0) in vec3 vert;
layout (location = 1) in vec3 normal;
layout (location = 2) in vec2 texcoord;
layout (location = 3) in float lookupIndex;

out flat ivec2 fragTexCoord;
out vec3 vertex;

void main() {
    vertex = vert;

    fragTexCoord = ivec2(int(normal.x), int(normal.y));

    vec4 heightTexel = imageLoad(nextHeightTex, fragTexCoord);
    float terrainHeight = heightTexel.r;

    gl_Position = projection * camera * vec4(vec3(vert.x, terrainHeight * height, vert.z), 1.0);
}