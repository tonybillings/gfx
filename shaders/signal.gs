#version 410 core

layout(lines_adjacency) in;
layout(triangle_strip, max_vertices = 4) out;

uniform float step;
uniform float thickness;

void main() {
    vec4 tangent1 = gl_in[1].gl_Position - gl_in[0].gl_Position;
    vec4 tangent2 = gl_in[2].gl_Position - gl_in[1].gl_Position;

    vec4 dir = 2.0 * tangent2 + tangent1;
    dir = normalize(dir);

    vec4 normal = vec4(-dir.y, dir.x, 0.0, 0.0);

    float thicknessNorm = step * thickness / 2.0;

    gl_Position = gl_in[1].gl_Position + normal * thicknessNorm;
    EmitVertex();

    gl_Position = gl_in[1].gl_Position - normal * thicknessNorm;
    EmitVertex();

    gl_Position = gl_in[2].gl_Position + normal * thicknessNorm;
    EmitVertex();

    gl_Position = gl_in[2].gl_Position - normal * thicknessNorm;
    EmitVertex();

    EndPrimitive();
}
