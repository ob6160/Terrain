#version 430 core

uniform sampler2D tboHeightmap;
uniform vec3 hitpos;
uniform float lightingDir;

layout (rgba32f, binding = 0) readonly uniform highp image2D nextHeightTex;
layout (rgba32f, binding = 2) readonly uniform highp image2D nextVelocityTex;

in vec2 fragTexCoord;
in vec3 vertex;

in float terrainHeight;

out vec4 color;

void main() {
//    vec2 light = vec2(-0.5, 0.5);
//    vec3 colour = vec3(0.0, 0.3, 0.0);
//
//    float brightness;
//
//    vec2 shift1 = vec2(fragTexCoord.x + light.x * lightingDir, fragTexCoord.y + light.y * lightingDir);
//    vec2 shift2 = vec2(fragTexCoord.x - light.x * lightingDir, fragTexCoord.y - light.y * lightingDir);

    vec2 du = vec2(0.5, 0.0);
    vec2 dv = vec2(0.0, 0.5);
    vec4 state_left = texture2D(tboHeightmap, fragTexCoord + du);
    vec4 state_right = texture2D(tboHeightmap, fragTexCoord - du);
    vec4 state_top = texture2D(tboHeightmap, fragTexCoord + dv);
    vec4 state_bottom = texture2D(tboHeightmap, fragTexCoord - dv);

    vec3 sunPos = vec3(2.1, 3.0, 2.8);

    float sr_water = state_right.r + state_right.g;
    float sl_water = state_left.r + state_left.g;
    float sb_water = state_bottom.r + state_bottom.g;
    float st_water = state_top.r + state_top.g;

    float dhdu = 0.5 * (sr_water - sl_water);
    float dhdv = 0.5 * (sb_water - st_water);
    vec3 normal = vec3(dhdu, 1.0, dhdv);
    float nl = max(0.0, dot(normal, sunPos.xyz));
    vec3 diff = nl * vec3(0.2);
    float ambient = 0.1;

    vec4 current_state = texture2D(tboHeightmap, fragTexCoord);
    vec4 cur_vel = imageLoad(nextVelocityTex, ivec2(fragTexCoord));

    vec3 waterColour = vec3(0.0, 0.0, 1.0);
    vec3 colour = mix(vec3(0.0, 1.0, 0.0), waterColour, current_state.g*50.0);


    color = vec4(diff * colour, 1.0);
}