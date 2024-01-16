#version 410

in vec2 UV;

out vec4 FragColor;

uniform sampler2D tex2D;

void main()
{
    FragColor = texture(tex2D, UV);
}
