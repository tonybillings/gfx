#version 410

in vec2 UV;

out vec4 FragColor;

uniform vec4 color;
uniform sampler2D tex2D;

void main()
{
    vec4 sampled = texture(tex2D, UV);
    FragColor = sampled * color;
}
