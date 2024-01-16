#version 410

in vec2 TexCoords;
out vec4 color;
uniform sampler2D textTexture;

void main() {
    vec4 sampled = texture(textTexture, TexCoords);
    color = vec4(sampled);
}
