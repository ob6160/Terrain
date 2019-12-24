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
layout (location = 3) in float inWaterHeight;

out vec2 fragTexCoord;
out vec3 vertex;
out float vWaterHeight;

void main() {
    fragTexCoord = texcoord;
    vertex = vert;
    vWaterHeight = inWaterHeight * 50.0;

    float waterHeightMod = 0.0;
    if(inWaterHeight > 0.0) {
        waterHeightMod = inWaterHeight * height;
    }

    gl_Position = projection * camera * vec4(vec3(vert.x, height * vert.y + waterHeightMod, vert.z), 1.0);
}