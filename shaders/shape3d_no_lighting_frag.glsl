#version 410 core

in vec2 UV;

out vec4 FragColor;

uniform sampler2D u_DiffuseMap;

layout (std140) uniform BasicMaterial {
    vec4    Ambient;
    vec4    Diffuse;
    vec4    Specular;
    vec4    Emissive;
    float   Shininess;
    float   Transparency;
} u_Material;

void main() {
    vec4 mapDiffuse = texture(u_DiffuseMap, UV).rgba;
    vec4 litDiffuse = (u_Material.Ambient + u_Material.Diffuse) * mapDiffuse;
    FragColor = vec4(litDiffuse.rgb, 1.0 - u_Material.Transparency);
}
