#version 410

in vec2 UV;
out vec4 FragColor;

uniform sampler2D tex2D;
uniform float blurAmount;

void main() {
    vec2 texOffset = 1.0 / textureSize(tex2D, 0);
    float yOffset = texOffset.y * blurAmount;

    int samples = int(5.0 + 4.0 * blurAmount);
    samples = clamp(samples, 1, 25);

    float weightSum = 0.0;
    vec4 color = vec4(0.0);

    for(int i = -samples / 2; i <= samples / 2; i++) {
        float weight = exp(-0.5 * (i * i) / (blurAmount * blurAmount));
        weightSum += weight;
        vec2 offset = vec2(0.0, yOffset * float(i));
        color += texture(tex2D, UV + offset) * weight;
    }

    color /= weightSum;
    FragColor = color;
}
