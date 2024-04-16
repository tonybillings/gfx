#version 410 core

in vec2 UV;

out vec4 FragColor;

uniform vec4 u_Color;
uniform sampler2D u_DiffuseMap;

void main()
{
    vec4 mapDiffuse = texture(u_DiffuseMap, UV);
    FragColor = mapDiffuse * u_Color;
}
