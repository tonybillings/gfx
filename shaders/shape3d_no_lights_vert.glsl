#version 410 core

in vec3 a_Position;
in vec2 a_UV;

out vec2 UV;

uniform mat4 u_WorldMat;

layout (std140) uniform BasicCamera {
    vec4 Position;
    vec4 Target;
    vec4 Up;
    mat4 ViewProjMat;
} u_Camera;

void main() {
    UV = a_UV;
    vec3 worldPos = vec3(u_WorldMat * vec4(a_Position, 1.0));
    gl_Position = u_Camera.ViewProjMat * vec4(worldPos, 1.0);
}
