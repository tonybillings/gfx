#version 410 core

const int MAX_LIGHT_COUNT = 4;

in vec3 FragPos;
in vec3 Normal;
in vec2 UV;
in vec3 CameraPos;

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

struct Light {
    vec3 Color;
    vec3 Direction;
};
uniform int     u_LightCount;
uniform Light   u_Lights[MAX_LIGHT_COUNT];

void main() {
    vec3 norm = normalize(Normal);
    vec3 viewDir = normalize(CameraPos - FragPos);
    vec3 mapDiffuse = texture(u_DiffuseMap, UV).rgb;
    vec3 tintDiffuse = u_Material.Diffuse.rgb * mapDiffuse;

    vec3 result = u_Material.Ambient.rgb * tintDiffuse + u_Material.Emissive.rgb;
    for(int i = 0; i < u_LightCount; i++) {
        vec3 lightDir = normalize(-u_Lights[i].Direction);
        float diffPower = max(dot(norm, lightDir), 0.0);
        vec3 litDiffuse = tintDiffuse * u_Lights[i].Color * diffPower;
        vec3 reflectDir = reflect(-lightDir, norm);
        float specPower = pow(max(dot(viewDir, reflectDir), 0.0), u_Material.Shininess);
        vec3 litSpecular = u_Material.Specular.rgb * specPower * u_Lights[i].Color;
        result += litDiffuse + litSpecular;
    }

    FragColor = vec4(result, 1.0 - u_Material.Transparency);
}
