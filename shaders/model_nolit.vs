#version 410 core

layout(location = 0) in vec3 inPosition;
layout(location = 1) in vec3 inNormal;
layout(location = 2) in vec2 inTexCoord;

uniform mat4 world;
uniform mat4 viewProj;

out vec3 FragPos;
out vec2 TexCoord;

void main() {
    TexCoord = inTexCoord;
    FragPos = vec3(world * vec4(inPosition, 1.0));
    gl_Position = viewProj * vec4(FragPos, 1.0);
}
