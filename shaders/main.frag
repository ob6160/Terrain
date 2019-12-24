#version 410 core

uniform sampler2D tex;
uniform vec3 hitpos;
in float vWaterHeight;
in vec2 fragTexCoord;
in vec3 vertex;
out vec4 color;

void main() {
    vec3 colour = vec3(0.0, 1.0, 0.0) * vertex.y;

    vec3 waterColour = vec3(0.0, 0.0, 0.6) / vWaterHeight + vec3(0.0,0.0,0.3);

    if(vWaterHeight + vertex.y > vertex.y + 0.001) {
        colour = waterColour;
    }
    color = vec4(colour, 1.0);
}