#version 410 core

in vec3 a_Position;
in vec3 a_Normal;
in vec2 a_UV;
in vec3 a_Tangent;
in vec3 a_Bitangent;

out vec3 FragPos;
out mat3 TBN;
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
    vec3 T = normalize(mat3(u_WorldMat) * a_Tangent);
    vec3 B = normalize(mat3(u_WorldMat) * a_Bitangent);
    vec3 N = normalize(mat3(u_WorldMat) * a_Normal);
    TBN = mat3(T, B, N);
    UV = a_UV;
    CameraPos = u_Camera.Position.xyz;
    gl_Position = u_Camera.ViewProjMat * vec4(FragPos, 1.0);
}
