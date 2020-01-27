#version 430 core

uniform sampler2D tex;
uniform vec3 hitpos;

uniform samplerBuffer tboWaterHeight;

in vec2 fragTexCoord;
in vec3 vertex;

in float terrainHeight;
in float waterHeight;

out vec4 color;

void main() {
    vec3 colour = vec3(0.0, 0.5 * terrainHeight, 0.0);
    float iWater = 1.0 / clamp(waterHeight, 1.0, 0.01);
    vec3 waterColour = vec3(0.0, 0.0, 0.8*iWater);

    vec3 mixed = mix(colour, waterColour, clamp(waterHeight*20.0, 0.0, 1.0));
    color = vec4(mixed, 1.0);
}