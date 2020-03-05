#version 430 core

uniform sampler2D tex;
uniform vec3 hitpos;

layout (rgba32f, binding = 0) readonly uniform highp image2D nextHeightTex;
layout (rgba32f, binding = 2) readonly uniform highp image2D nextVelocityTex;

in flat ivec2 fragTexCoord;
in vec3 vertex;

in float terrainHeight;

out vec4 color;

void main() {
    ivec2 light = ivec2(1.0, 1.0);
    vec3 colour = vec3(0.0, 0.4, 0.0);

    float brightness;
    ivec2 shift1 = ivec2(int(fragTexCoord.x + 0.1), int(fragTexCoord.y + 0.1));
    ivec2 shift2 = ivec2(int(fragTexCoord.x - 0.1), int(fragTexCoord.y - 0.1));

    vec4 vecTexel = imageLoad(nextVelocityTex, shift1);
    vec4 texel1 = imageLoad(nextHeightTex, shift1);
    vec4 texel2 = imageLoad(nextHeightTex, shift2);

    brightness = (
        (texel1.r + texel1.g) * 200.0 -
        (texel2.r + texel2.g) * 200.0
    ) / 2.0 + 0.7;

//    colour = mix(colour, vec3(0.0, 0.0, 1.0), clamp(texel1.g*500.0, 0.0, 1.0));
//    vec3 waterMod = vec3(0.0, 1.0-texel1.g*10.0, texel1.g*10.0);
    if(texel1.g > .0001) {
        colour = vec3(0.0, 0.0, clamp(texel1.g * 100.0, 0.5, 1.0));
//        colour.b -= clamp(texel1.b * 100000000000.0, 0.4, 0.6);
        colour.r += vecTexel.r;
    }



    color = vec4(colour * brightness, 1.0);
}