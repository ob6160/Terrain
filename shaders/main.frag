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

    vec3 offset = vec3(-1.0/512.0, 0.0, 1.0/512.0);

    float s1 = texture2D(tboHeightmap, fragTexCoord + offset.xy).r;
    float s2 = texture2D(tboHeightmap, fragTexCoord + offset.zy).r;
    float s3 = texture2D(tboHeightmap, fragTexCoord + offset.yx).r;
    float s4 = texture2D(tboHeightmap, fragTexCoord + offset.yz).r;
    vec3 va = normalize(vec3(0.5, 0.0, s2 - s1));
    vec3 vb = normalize(vec3(0.0, 0.5, s3 - s4));
    vec3 n = normalize(cross(va, vb));


    vec3 lightPos = vec3(0.5, 2.0, 0.5);
    vec3 lightDir = normalize(lightPos);
    float lambertTerm = max(dot(lightDir, n), 0.0);

    color = vec4(vec3(lambertTerm), 1.0);
}