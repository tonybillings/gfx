#version 410

layout(location = 0) in vec2 inPos;
layout(location = 1) in vec2 inUV;

uniform vec2 scale;
uniform vec3 position;
uniform float rotation;
uniform vec3 origin;

out vec2 UV;

void main()
{
    mat2 rot = mat2(cos(rotation), -sin(rotation), sin(rotation), cos(rotation));

    vec2 adjustedPos = (inPos - origin.xy);
    vec2 rotatedPos = rot * adjustedPos;
    vec2 scaledPos = rotatedPos * scale;
    vec2 finalPos = scaledPos + position.xy;

    gl_Position = vec4(finalPos, position.z, 1.0);
    UV = inUV;
}
