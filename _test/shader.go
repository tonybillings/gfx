package _test

var ColorVertShader = `#version 410 core
in vec3 a_Position;
in vec3 a_Color;
out vec4 Color;
void main() {
	Color = vec4(a_Color, 1.0);
  	gl_Position = vec4(a_Position, 1.0);
}
`

var ColorFragShader = `#version 410 core
in vec4 Color;
out vec4 FragColor;
void main() {
	FragColor = Color;
}
`

var TextureVertShader = `#version 410 core
in vec3 a_Position;
in vec2 a_UV;
out vec2 UV;
void main() {
	UV = a_UV;
  	gl_Position = vec4(a_Position, 1.0);
}
`

var TextureFragShader = `#version 410 core
in vec2 UV;
out vec4 FragColor;
uniform sampler2D u_DiffuseMap; 
void main() {
	FragColor = texture(u_DiffuseMap, UV);
}
`
