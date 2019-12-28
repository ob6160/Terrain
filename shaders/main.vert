#version 410 core

uniform mat4 projection;
uniform mat4 camera;
uniform mat4 model;
uniform float angle;
uniform float height;
uniform vec3 hitpos;

uniform samplerBuffer tboWaterHeight;
uniform samplerBuffer tboHeightmap;

layout (location = 0) in vec3 vert;
layout (location = 1) in vec3 normal;
layout (location = 2) in vec2 texcoord;
layout (location = 3) in float lookupIndex;

out vec2 fragTexCoord;
out vec3 vertex;
out float waterHeight;
out float terrainHeight;

void main() {
    fragTexCoord = texcoord;
    vertex = vert;

    terrainHeight = texelFetch(tboHeightmap, int(normal.x)).r;

    // Water calculations
    waterHeight = texelFetch(tboWaterHeight, int(normal.x)).r;
    float waterHeightMod = 0.0;
    if(waterHeight > 0.0) {
        waterHeightMod = waterHeight * height;
    }
    gl_Position = projection * camera * vec4(vec3(vert.x, terrainHeight * height, vert.z), 1.0);
//    gl_Position = projection * camera * vec4(vec3(vert.x, height * vert.y, vert.z), 1.0);
}