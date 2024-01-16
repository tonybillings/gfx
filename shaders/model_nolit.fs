#version 410 core

in vec2 TexCoord;
in vec3 FragPos;

out vec4 FragColor;

uniform sampler2D tex2D;

struct Material {
    vec3 ambient;
    vec3 diffuse;
};
uniform Material material;

void main() {
    vec3 ambient = material.ambient;
    vec3 diffuse = material.diffuse;

    vec4 texColor = texture(tex2D, TexCoord).rgba;
    vec3 result = (ambient + diffuse) * texColor.rgb;
    FragColor = vec4(result, texColor.a);
}
