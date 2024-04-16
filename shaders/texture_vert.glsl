#version 410 core

in vec2 a_Position;
in vec2 a_UV;

out vec2 UV;

void main()
{
    gl_Position = vec4(a_Position, 0.0, 1.0);
    UV = a_UV;
}
