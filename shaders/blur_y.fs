#version 410

in vec2 UV;

out vec4 FragColor;

uniform sampler2D tex2D;
uniform float blurAmount;

const int SAMPLES = 9;

const float weights[SAMPLES] = float[](
    0.0093, 0.028002, 0.065984, 0.121703, 0.175037,
    0.121703, 0.065984, 0.028002, 0.0093
);

const float weightSum = 0.6097;

void main() {
    vec2 texOffset = 1.0 / textureSize(tex2D, 0);
    float yOffset = texOffset.y * blurAmount;

    vec4 color = texture(tex2D, UV) * weights[SAMPLES / 2];

    for(int i = 1; i <= SAMPLES / 2; i++) {
        float weightPos = weights[(SAMPLES / 2) + i];
        float weightNeg = weights[(SAMPLES / 2) - i];

        vec2 offsetPos = vec2(0.0, yOffset * float(i));
        vec2 offsetNeg = -offsetPos;

        color += texture(tex2D, UV + offsetPos) * weightPos;
        color += texture(tex2D, UV + offsetNeg) * weightNeg;
    }

    color *= 1.0 / weightSum;
    FragColor = color;
}
