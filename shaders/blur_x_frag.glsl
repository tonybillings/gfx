#version 410 core

in vec2 UV;

out vec4 FragColor;

uniform sampler2D u_TextureMap;
uniform float u_BlurAmount;

void main() {
    vec2 texOffset = 1.0 / textureSize(u_TextureMap, 0);
    float xOffset = texOffset.x * u_BlurAmount;

    int samples = int(5.0 + 4.0 * u_BlurAmount);
    samples = clamp(samples, 1, 25);

    float weightSum = 0.0;
    vec4 color = vec4(0.0);

    for(int i = -samples / 2; i <= samples / 2; i++) {
        float weight = exp(-0.5 * (i * i) / (u_BlurAmount * u_BlurAmount));
        weightSum += weight;
        vec2 offset = vec2(xOffset * float(i), 0.0);
        color += texture(u_TextureMap, UV + offset) * weight;
    }

    color /= weightSum;
    FragColor = color;
}
