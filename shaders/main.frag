#version 410 core

uniform sampler2D tex;
uniform vec3 hitpos;

uniform samplerBuffer tboWaterHeight;

in vec2 fragTexCoord;
in vec3 vertex;
in float waterHeight;

out vec4 color;

void main() {
    vec3 colour = vec3(0.1, 0.8, 0.2);
    float iWater = 1.0 / (waterHeight*10.0);
    vec3 waterColour = vec3(0.0, 0.0, 1.0*iWater);


    vec3 mixed = mix(colour, waterColour, waterHeight*10.0);
    color = vec4(vec3(0.0, 0.3, 0.0) + mixed, 1.0);
}