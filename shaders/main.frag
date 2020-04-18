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
    vec3 lightColour = vec3(1.0);
    vec3 ambient = 0.1 * lightColour;

    vec3 height_colours[3];
    height_colours[0] = vec3(0.1, 0.8, 0.1);
    height_colours[1] =  vec3(0.4, 0.8, 0.4);
    height_colours[2] = vec3(1.0,1.0,1.0);

    vec3 terrainColour = vec3(0.0, 1.0, 0.0);
    vec3 waterColour = vec3(0.0, 0.0, 1.0);
    vec4 hmSample = texture2D(tboHeightmap, fragTexCoord);
    float height = hmSample.r * 2.0;
    if(height < 1.0) {
        terrainColour = mix(height_colours[1], height_colours[0], 2.0 - height);
    } else {
        terrainColour = mix(height_colours[2], height_colours[1], 2.0 - height);
    }
    terrainColour = mix(terrainColour, waterColour, hmSample.g * 10.0);

    vec3 offset = vec3(-1.0/512.0, 0.0, 1.0/512.0);

    float s1 = texture2D(tboHeightmap, fragTexCoord + offset.xy).r;
    float s2 = texture2D(tboHeightmap, fragTexCoord + offset.zy).r;
    float s3 = texture2D(tboHeightmap, fragTexCoord + offset.yx).r;
    float s4 = texture2D(tboHeightmap, fragTexCoord + offset.yz).r;
    vec3 va = normalize(vec3(lightingDir, 0.0, s2 - s1));
    vec3 vb = normalize(vec3(0.0, lightingDir, s3 - s4));
    vec3 n = normalize(cross(va, vb));


    vec3 lightPos = vec3(0.5, 2.0, 0.5);
    vec3 lightDir = normalize(lightPos);
    float diff = max(dot(lightDir, n), 0.0);

    vec3 diffuse = lightColour * diff;

    vec3 result = (ambient + diffuse) * terrainColour;

    color = vec4(result, 1.0);
}