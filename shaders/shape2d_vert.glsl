#version 410 core

in vec2 a_Position;
in vec2 a_UV;

out vec2 UV;

layout (std140) uniform Transform {
    vec4 Origin;
    vec4 Position;
    vec4 Rotation;
    vec4 Scale;
} u_Transform;

void main()
{
    mat2 rot = mat2(cos(u_Transform.Rotation.z), -sin(u_Transform.Rotation.z), sin(u_Transform.Rotation.z), cos(u_Transform.Rotation.z));

    vec2 adjustedPos = (a_Position - u_Transform.Origin.xy);
    vec2 rotatedPos = rot * adjustedPos;
    vec2 scaledPos = rotatedPos * u_Transform.Scale.xy;
    vec2 finalPos = scaledPos + u_Transform.Position.xy;

    UV = a_UV;
    gl_Position = vec4(finalPos, u_Transform.Position.z, 1.0);
}
