#version 410 core

in vec3 a_Position;
in vec3 a_Normal;
in vec2 a_UV;

out vec3 FragPos;
out vec3 Normal;
out vec2 UV;
out vec3 CameraPos;

uniform mat4 u_WorldMat;

layout (std140) uniform BasicCamera {
    vec4 Position;
    vec4 Target;
    vec4 Up;
    mat4 ViewProjMat;
} u_Camera;

void main() {
    FragPos = vec3(u_WorldMat * vec4(a_Position, 1.0));
    Normal = mat3(transpose(inverse(u_WorldMat))) * a_Normal;
    UV = a_UV;
    CameraPos = u_Camera.Position.xyz;
    gl_Position = u_Camera.ViewProjMat * vec4(FragPos, 1.0);
}
