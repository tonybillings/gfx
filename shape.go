package gfx

import (
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"math"
)

const (
	defaultShapeName    = "Shape"
	defaultTriangleName = "Triangle"
	defaultQuadName     = "Quad"
	defaultSquareName   = "Square"
	defaultDotName      = "Dot"
	defaultCircleName   = "Circle"

	thicknessEpsilon = 0.00001
)

type Shape struct {
	WindowObjectBase

	vao uint32
	vbo uint32

	shapeShader         uint32
	texturedShapeShader uint32
	blurXShader         uint32
	blurYShader         uint32
	textureShader       uint32

	colorUniformLoc        int32
	originUniformLoc       int32
	positionUniformLoc     int32
	rotationUniformLoc     int32
	scaleUniformLoc        int32
	aspectRatioUniformLoc  int32
	shapeTex2DUniformLoc   int32
	blurTex2DUniformLoc    int32
	blurAmountUniformLoc   int32
	textureTex2DUniformLoc int32

	texture         uint32
	textureFilename string

	blurShapeFrameBuffer uint32
	blurShapeTexture     uint32
	blurXTexture         uint32
	blurXYTexture        uint32
	blurTextureVao       uint32
	blurTextureVbo       uint32
	blurTextureVertices  []float32

	sides     uint
	thickness float32
	length    float32
	drawMode  uint32

	label *Label
}

/******************************************************************************
 Label Initialization
******************************************************************************/

func (s *Shape) initLabel(window *Window) bool {
	return s.label.Init(window)
}

/******************************************************************************
 Vertices Initialization
******************************************************************************/

func (s *Shape) initVerticesTriangle() {
	a := [2]float32{0, 1}
	b := [2]float32{-.87, -.5}
	c := [2]float32{.87, -.5}

	s.stateMutex.Lock()

	if s.thickness > thicknessEpsilon {
		centroid := [2]float32{
			(a[0] + b[0] + c[0]) / 3.0,
			(a[1] + b[1] + c[1]) / 3.0,
		}

		interpolate := func(p [2]float32) [2]float32 {
			return [2]float32{
				p[0] + (centroid[0]-p[0])*s.thickness,
				p[1] + (centroid[1]-p[1])*s.thickness,
			}
		}

		aInner := interpolate(a)
		bInner := interpolate(b)
		cInner := interpolate(c)

		s.vertices = []float32{
			a[0], a[1],
			aInner[0], aInner[1],

			b[0], b[1],
			bInner[0], bInner[1],

			c[0], c[1],
			cInner[0], cInner[1],

			a[0], a[1],
			aInner[0], aInner[1],
		}
		s.vertexCount = 8
		s.drawMode = gl.TRIANGLE_STRIP
	} else {
		s.vertices = []float32{
			a[0], a[1],
			b[0], b[1],
			c[0], c[1],
		}
		s.vertexCount = 3
		s.drawMode = gl.TRIANGLES
	}

	s.stateMutex.Unlock()
}

func (s *Shape) initVerticesTriangleWithUV() {
	a := [2]float32{0, 1}
	b := [2]float32{-.87, -.5}
	c := [2]float32{.87, -.5}

	s.stateMutex.Lock()

	if s.thickness > thicknessEpsilon {
		centroid := [2]float32{
			(a[0] + b[0] + c[0]) / 3.0,
			(a[1] + b[1] + c[1]) / 3.0,
		}

		interpolate := func(p [2]float32) [2]float32 {
			return [2]float32{
				p[0] + (centroid[0]-p[0])*s.thickness,
				p[1] + (centroid[1]-p[1])*s.thickness,
			}
		}

		aInner := interpolate(a)
		bInner := interpolate(b)
		cInner := interpolate(c)

		s.vertices = []float32{
			a[0], a[1], 0.5, 0.0,
			aInner[0], aInner[1], 0.5, 0.0,

			b[0], b[1], 0.0, 1.0,
			bInner[0], bInner[1], 0.0, 1.0,

			c[0], c[1], 1.0, 1.0,
			cInner[0], cInner[1], 1.0, 1.0,

			a[0], a[1], 0.5, 0.0,
			aInner[0], aInner[1], 0.5, 0.0,
		}

		s.vertexCount = 8
		s.drawMode = gl.TRIANGLE_STRIP
	} else {
		s.vertices = []float32{
			a[0], a[1], 0.5, 0.0,
			b[0], b[1], 0.0, 1.0,
			c[0], c[1], 1.0, 1.0,
		}
		s.vertexCount = 3
		s.drawMode = gl.TRIANGLES
	}

	s.stateMutex.Unlock()
}

func (s *Shape) initVerticesQuadrangle() {
	xAspectRatioInv := s.WorldScale().Y() / s.WorldScale().X()

	s.stateMutex.Lock()

	if s.thickness > thicknessEpsilon {
		topLeft := [2]float32{-1.0, 1.0}
		bottomRight := [2]float32{1.0, -1.0}
		center := [2]float32{
			(topLeft[0] + bottomRight[0]) / 2.0,
			(topLeft[1] + bottomRight[1]) / 2.0,
		}

		interpolate := func(p [2]float32) [2]float32 {
			return [2]float32{
				p[0] + (center[0]-p[0])*s.thickness*xAspectRatioInv,
				p[1] + (center[1]-p[1])*s.thickness,
			}
		}

		topRight := [2]float32{bottomRight[0], topLeft[1]}
		bottomLeft := [2]float32{topLeft[0], bottomRight[1]}

		cornersOuter := [][2]float32{topLeft, topRight, bottomRight, bottomLeft, topLeft}
		cornersInner := [][2]float32{
			interpolate(topLeft),
			interpolate(topRight),
			interpolate(bottomRight),
			interpolate(bottomLeft),
			interpolate(topLeft),
		}

		vertices := make([]float32, 0, 20)
		for i := range cornersOuter {
			vertices = append(vertices, cornersOuter[i][0], cornersOuter[i][1])
			vertices = append(vertices, cornersInner[i][0], cornersInner[i][1])
		}

		s.vertices = vertices
		s.vertexCount = int32(len(vertices) / 2)
		s.drawMode = gl.TRIANGLE_STRIP
	} else {
		s.vertices = []float32{
			-1.0, 1.0,
			1.0, 1.0,
			-1.0, -1.0,
			1.0, 1.0,
			1.0, -1.0,
			-1.0, -1.0,
		}
		s.vertexCount = 6
		s.drawMode = gl.TRIANGLES
	}

	s.stateMutex.Unlock()
}

func (s *Shape) initVerticesQuadrangleWithUV() {
	xAspectRatioInv := s.WorldScale().Y() / s.WorldScale().X()

	topLeft := [2]float32{-1.0, 1.0}
	bottomRight := [2]float32{1.0, -1.0}

	s.stateMutex.Lock()

	if s.thickness > thicknessEpsilon {
		center := [2]float32{
			(topLeft[0] + bottomRight[0]) / 2.0,
			(topLeft[1] + bottomRight[1]) / 2.0,
		}

		interpolate := func(p [2]float32) [2]float32 {
			return [2]float32{
				p[0] + (center[0]-p[0])*s.thickness*xAspectRatioInv,
				p[1] + (center[1]-p[1])*s.thickness,
			}
		}

		topRight := [2]float32{bottomRight[0], topLeft[1]}
		bottomLeft := [2]float32{topLeft[0], bottomRight[1]}

		cornersOuter := [][2]float32{topLeft, topRight, bottomRight, bottomLeft, topLeft}

		cornersInner := [][2]float32{
			interpolate(topLeft),
			interpolate(topRight),
			interpolate(bottomRight),
			interpolate(bottomLeft),
			interpolate(topLeft),
		}

		uvOuter := [][2]float32{{0, 0}, {1, 0}, {1, 1}, {0, 1}, {0, 0}}
		uvInner := make([][2]float32, len(cornersInner))

		for i, corner := range cornersInner {
			uvInner[i][0] = (corner[0] + 1) / 2.0
			uvInner[i][1] = 1 - (corner[1]+1)/2.0
		}

		vertices := make([]float32, 0, 40)
		for i := range cornersOuter {
			vertices = append(vertices, cornersOuter[i][0], cornersOuter[i][1], uvOuter[i][0], uvOuter[i][1])
			vertices = append(vertices, cornersInner[i][0], cornersInner[i][1], uvInner[i][0], uvInner[i][1])
		}

		s.vertices = vertices
		s.vertexCount = int32(len(vertices) / 4)
		s.drawMode = gl.TRIANGLE_STRIP
	} else {
		s.vertices = []float32{
			-1.0, 1.0, 0.0, 0.0,
			1.0, 1.0, 1.0, 0.0,
			-1.0, -1.0, 0.0, 1.0,
			1.0, 1.0, 1.0, 0.0,
			1.0, -1.0, 1.0, 1.0,
			-1.0, -1.0, 0.0, 1.0,
		}

		s.vertexCount = 6
		s.drawMode = gl.TRIANGLES
	}

	s.stateMutex.Unlock()
}

func (s *Shape) initVerticesPolygon() {
	radius := float32(1)
	maxAngle := 2 * math.Pi * float64(s.length)
	angleIncrement := maxAngle / float64(s.sides-1)

	s.stateMutex.Lock()

	if maxAngle < thicknessEpsilon {
		s.vertices = make([]float32, 0)
		s.vertexCount = 0
		s.stateMutex.Unlock()
		return
	}

	if s.thickness > thicknessEpsilon {
		vertices := make([]float32, 0, s.sides*2*2)
		innerRadius := radius - s.thickness

		for angle := 0.0; angle <= maxAngle; angle += angleIncrement {
			outerX := float32(math.Cos(angle)) * radius
			outerY := float32(math.Sin(angle)) * radius
			innerX := float32(math.Cos(angle)) * innerRadius
			innerY := float32(math.Sin(angle)) * innerRadius

			vertices = append(vertices, outerX, outerY)
			vertices = append(vertices, innerX, innerY)
		}

		outerX := float32(math.Cos(maxAngle)) * radius
		outerY := float32(math.Sin(maxAngle)) * radius
		innerX := float32(math.Cos(maxAngle)) * innerRadius
		innerY := float32(math.Sin(maxAngle)) * innerRadius
		vertices = append(vertices, outerX, outerY)
		vertices = append(vertices, innerX, innerY)

		s.vertices = vertices
		s.vertexCount = int32(len(vertices) / 2)
		s.drawMode = gl.TRIANGLE_STRIP
	} else {
		vertices := make([]float32, 0, s.sides*2)

		for angle := 0.0; angle <= maxAngle; angle += angleIncrement {
			x := float32(math.Cos(angle))
			y := float32(math.Sin(angle))
			vertices = append(vertices, x, y)
		}

		x := float32(math.Cos(maxAngle))
		y := float32(math.Sin(maxAngle))
		vertices = append(vertices, x, y)

		triangulatedVertices := make([]float32, 0, len(vertices)*3)
		centerX, centerY := float32(0), float32(0)
		for i := 0; i < len(vertices)-2; i += 2 {
			triangulatedVertices = append(triangulatedVertices, centerX, centerY, vertices[i], vertices[i+1], vertices[i+2], vertices[i+3])
		}

		s.vertices = triangulatedVertices
		s.vertexCount = int32(len(triangulatedVertices) / 2)
		s.drawMode = gl.TRIANGLES
	}

	s.stateMutex.Unlock()
}

func (s *Shape) initVerticesPolygonWithUV() {
	radius := float32(1)
	maxAngle := 2 * math.Pi * float64(s.length)
	angleIncrement := maxAngle / float64(s.sides-1)

	s.stateMutex.Lock()

	if maxAngle < thicknessEpsilon {
		s.vertices = make([]float32, 0)
		s.vertexCount = 0
		s.stateMutex.Unlock()
		return
	}

	if s.thickness > thicknessEpsilon {
		innerRadius := radius - s.thickness
		vertices := make([]float32, 0, s.sides*4*2)

		for angle := 0.0; angle <= maxAngle; angle += angleIncrement {
			outerX := float32(math.Cos(angle)) * radius
			outerY := float32(math.Sin(angle)) * radius
			innerX := float32(math.Cos(angle)) * innerRadius
			innerY := float32(math.Sin(angle)) * innerRadius

			vertices = append(vertices, outerX, outerY, (outerX+1)/2, 1-(outerY+1)/2)
			vertices = append(vertices, innerX, innerY, (innerX+1)/2, 1-(innerY+1)/2)
		}

		outerX := float32(math.Cos(maxAngle)) * radius
		outerY := float32(math.Sin(maxAngle)) * radius
		innerX := float32(math.Cos(maxAngle)) * innerRadius
		innerY := float32(math.Sin(maxAngle)) * innerRadius
		vertices = append(vertices, outerX, outerY, (outerX+1)/2, 1-(outerY+1)/2)
		vertices = append(vertices, innerX, innerY, (innerX+1)/2, 1-(innerY+1)/2)

		s.vertices = vertices
		s.vertexCount = int32(len(vertices) / 4)
		s.drawMode = gl.TRIANGLE_STRIP
	} else {
		vertices := make([]float32, 0, s.sides*4)

		for angle := 0.0; angle <= maxAngle; angle += angleIncrement {
			x := float32(math.Cos(angle))
			y := float32(math.Sin(angle))
			vertices = append(vertices, x, y, (x+1)/2, 1-(y+1)/2)
		}

		x := float32(math.Cos(maxAngle))
		y := float32(math.Sin(maxAngle))
		vertices = append(vertices, x, y, (x+1)/2, 1-(y+1)/2)

		triangulatedVertices := make([]float32, 0, (s.sides-1)*6*2)
		centerX, centerY, centerU, centerV := float32(0), float32(0), float32(0.5), float32(0.5)
		for i := 0; i < len(vertices)/4-1; i++ {
			triangulatedVertices = append(triangulatedVertices,
				centerX, centerY, centerU, centerV,
				vertices[i*4], vertices[i*4+1], vertices[i*4+2], vertices[i*4+3],
				vertices[(i+1)*4], vertices[(i+1)*4+1], vertices[(i+1)*4+2], vertices[(i+1)*4+3],
			)
		}

		s.vertices = triangulatedVertices
		s.vertexCount = int32(len(triangulatedVertices) / 4)
		s.drawMode = gl.TRIANGLES
	}

	s.stateMutex.Unlock()
}

func (s *Shape) initVertices() {
	switch s.sides {
	case 3:
		if s.textureFilename == "" {
			s.initVerticesTriangle()
		} else {
			s.initVerticesTriangleWithUV()
		}
	case 4:
		if s.textureFilename == "" {
			s.initVerticesQuadrangle()
		} else {
			s.initVerticesQuadrangleWithUV()
		}
	default:
		if s.textureFilename == "" {
			s.initVerticesPolygon()
		} else {
			s.initVerticesPolygonWithUV()
		}
	}
}

func (s *Shape) initBlurTextureVertices() {
	s.stateMutex.Lock()

	s.blurTextureVertices = []float32{
		-1.0, 1.0, 0.0, 1.0,
		1.0, 1.0, 1.0, 1.0,
		-1.0, -1.0, 0.0, 0.0,
		1.0, 1.0, 1.0, 1.0,
		1.0, -1.0, 1.0, 0.0,
		-1.0, -1.0, 0.0, 0.0,
	}

	s.stateMutex.Unlock()
}

/******************************************************************************
 Texture Initialization
******************************************************************************/

func (s *Shape) initTexture() {
	s.stateMutex.Lock()
	if s.vao == 0 {
		s.stateMutex.Unlock()
		return
	}

	gl.DeleteTextures(1, &s.texture)
	s.texture = 0

	if s.textureFilename != "" {
		s.texture = loadImage(s.textureFilename)
	}

	s.stateMutex.Unlock()
}

/******************************************************************************
 Shader Initialization
******************************************************************************/

func (s *Shape) initShaders() {
	s.stateMutex.Lock()

	s.shapeShader = GetShaderProgram(ShapeShaderProgram)
	s.texturedShapeShader = GetShaderProgram(ShapeTexturedShaderProgram)
	s.blurXShader = GetShaderProgram(BlurXShaderProgram)
	s.blurYShader = GetShaderProgram(BlurYShaderProgram)
	s.textureShader = GetShaderProgram(TextureShaderProgram)

	s.colorUniformLoc = gl.GetUniformLocation(s.shapeShader, gl.Str("color\x00"))
	s.originUniformLoc = gl.GetUniformLocation(s.shapeShader, gl.Str("origin\x00"))
	s.positionUniformLoc = gl.GetUniformLocation(s.shapeShader, gl.Str("position\x00"))
	s.rotationUniformLoc = gl.GetUniformLocation(s.shapeShader, gl.Str("rotation\x00"))
	s.scaleUniformLoc = gl.GetUniformLocation(s.shapeShader, gl.Str("scale\x00"))
	s.shapeTex2DUniformLoc = gl.GetUniformLocation(s.texturedShapeShader, gl.Str("tex2D\x00"))
	s.aspectRatioUniformLoc = gl.GetUniformLocation(s.shapeShader, gl.Str("aspectRatio\x00"))
	s.textureTex2DUniformLoc = gl.GetUniformLocation(s.textureShader, gl.Str("tex2D\x00"))
	s.blurAmountUniformLoc = gl.GetUniformLocation(s.blurXShader, gl.Str("blurAmount\x00"))

	s.stateMutex.Unlock()
}

/******************************************************************************
 VAO Initialization
******************************************************************************/

func (s *Shape) initVao(updateOnly bool) {
	s.stateMutex.Lock()

	if !updateOnly {
		gl.GenVertexArrays(1, &s.vao)
		gl.GenBuffers(1, &s.vbo)
	}

	if s.vao == 0 {
		s.stateMutex.Unlock()
		return
	}

	gl.BindVertexArray(s.vao)
	gl.BindBuffer(gl.ARRAY_BUFFER, s.vbo)

	if s.textureFilename == "" {
		stride := int32(2 * sizeOfFloat32)
		gl.EnableVertexAttribArray(0)
		gl.VertexAttribPointerWithOffset(0, 2, gl.FLOAT, false, stride, 0)
		gl.DisableVertexAttribArray(1)
	} else {
		stride := int32(4 * sizeOfFloat32)
		gl.EnableVertexAttribArray(0)
		gl.VertexAttribPointerWithOffset(0, 2, gl.FLOAT, false, stride, 0)
		gl.EnableVertexAttribArray(1)
		gl.VertexAttribPointerWithOffset(1, 2, gl.FLOAT, false, stride, uintptr(2*sizeOfFloat32))
	}

	if len(s.vertices) > 0 {
		gl.BufferData(gl.ARRAY_BUFFER, len(s.vertices)*sizeOfFloat32, gl.Ptr(s.vertices), gl.STATIC_DRAW)
	}

	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	gl.BindVertexArray(0)

	s.stateMutex.Unlock()
}

func (s *Shape) initBlurVao() {
	s.stateMutex.Lock()

	gl.GenVertexArrays(1, &s.blurTextureVao)
	gl.GenBuffers(1, &s.blurTextureVbo)

	gl.BindVertexArray(s.blurTextureVao)
	gl.BindBuffer(gl.ARRAY_BUFFER, s.blurTextureVbo)

	gl.GenFramebuffers(1, &s.blurShapeFrameBuffer)
	gl.BindFramebuffer(gl.FRAMEBUFFER, s.blurShapeFrameBuffer)

	stride := int32(4 * sizeOfFloat32)
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointerWithOffset(0, 2, gl.FLOAT, false, stride, 0)
	gl.EnableVertexAttribArray(1)
	gl.VertexAttribPointerWithOffset(1, 2, gl.FLOAT, false, stride, uintptr(2*sizeOfFloat32))

	gl.BufferData(gl.ARRAY_BUFFER, len(s.blurTextureVertices)*sizeOfFloat32, gl.Ptr(s.blurTextureVertices), gl.STATIC_DRAW)

	gl.GenTextures(1, &s.blurShapeTexture)
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, s.blurShapeTexture)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, s.Window().Width(), s.Window().Height(), 0, gl.RGBA, gl.UNSIGNED_BYTE, nil)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, s.blurShapeTexture, 0)

	gl.GenTextures(1, &s.blurXTexture)
	gl.BindTexture(gl.TEXTURE_2D, s.blurXTexture)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, s.Window().Width(), s.Window().Height(), 0, gl.RGBA, gl.UNSIGNED_BYTE, nil)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)

	gl.GenTextures(1, &s.blurXYTexture)
	gl.BindTexture(gl.TEXTURE_2D, s.blurXYTexture)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, s.Window().Width(), s.Window().Height(), 0, gl.RGBA, gl.UNSIGNED_BYTE, nil)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)

	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	gl.BindVertexArray(0)

	s.stateMutex.Unlock()
}

func (s *Shape) uninitBlurVao() {
	s.stateMutex.Lock()

	gl.DeleteVertexArrays(1, &s.blurTextureVao)
	gl.DeleteBuffers(1, &s.blurTextureVbo)

	gl.DeleteFramebuffers(1, &s.blurShapeFrameBuffer)

	gl.DeleteTextures(1, &s.blurShapeTexture)
	gl.DeleteTextures(1, &s.blurXTexture)
	gl.DeleteTextures(1, &s.blurXYTexture)

	s.stateMutex.Unlock()
}

/******************************************************************************
 OpenGL Initialization
******************************************************************************/

func (s *Shape) initGl() {
	s.initShaders()
	s.initVao(false)
	s.initBlurVao()
	s.initTexture()
}

func (s *Shape) uninitGl() {
	if !s.Initialized() {
		return
	}
	s.SetInitialized(false)
	s.stateMutex.Lock()

	gl.BindVertexArray(s.vao)
	gl.DisableVertexAttribArray(0)
	gl.DisableVertexAttribArray(1)
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	gl.BindVertexArray(0)
	gl.DeleteBuffers(1, &s.vbo)
	gl.DeleteVertexArrays(1, &s.vao)
	gl.DeleteTextures(1, &s.texture)

	s.uninitBlurVao()

	s.stateMutex.Unlock()
}

/******************************************************************************
 WindowObject Implementation
******************************************************************************/

func (s *Shape) setShaderScaleUniform(worldScale mgl32.Vec3) {
	if s.maintainAspectRatio {
		width := float32(s.window.width.Load())
		height := float32(s.window.height.Load())
		scale := [2]float32{}
		switch {
		case width > height:
			scale[0] = worldScale[0] * (height / width)
			scale[1] = worldScale[1]
		case height > width:
			scale[0] = worldScale[0]
			scale[1] = worldScale[1] * (width / height)
		default:
			scale[0] = worldScale[0]
			scale[1] = worldScale[1]
		}
		gl.Uniform2fv(s.scaleUniformLoc, 1, &scale[0])
	} else {
		gl.Uniform2fv(s.scaleUniformLoc, 1, &worldScale[0])
	}
}

func (s *Shape) renderBlurTexture(texture uint32, tex2DLoc int32) {
	gl.BindVertexArray(s.blurTextureVao)
	gl.BindBuffer(gl.ARRAY_BUFFER, s.blurTextureVbo)
	gl.BindTexture(gl.TEXTURE_2D, texture)
	gl.Uniform1i(tex2DLoc, 0)
	gl.DrawArrays(gl.TRIANGLES, 0, 6)
}

func (s *Shape) Init(window *Window) (ok bool) {
	if !s.WindowObjectBase.Init(window) {
		return false
	}

	s.initVertices()
	s.initBlurTextureVertices()
	s.initGl()
	s.initialized.Store(true)

	return s.initLabel(window)
}

func (s *Shape) Update(deltaTime int64) (ok bool) {
	if !s.WindowObjectBase.Update(deltaTime) {
		return false
	}

	return s.label.Update(deltaTime)
}

func (s *Shape) Draw(deltaTime int64) (ok bool) {
	if !s.visible.Load() || s.closed.Load() {
		return false
	}

	if s.closing.Load() {
		s.uninitGl()
		s.closed.Store(true)
		s.closing.Store(false)
		return false
	}

	window := s.Window()
	windowWidth := window.Width()
	windowHeight := window.Height()
	worldPos := s.WorldPosition()
	worldScale := s.WorldScale()
	worldRot := s.WorldRotation()

	s.stateMutex.Lock()

	gl.BindVertexArray(s.vao)
	gl.BindBuffer(gl.ARRAY_BUFFER, s.vbo)

	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	gl.Viewport(0, 0, windowWidth, windowHeight)

	if s.textureFilename == "" {
		gl.UseProgram(s.shapeShader)
		gl.BindTexture(gl.TEXTURE_2D, 0)
	} else {
		gl.UseProgram(s.texturedShapeShader)
		gl.BindTexture(gl.TEXTURE_2D, s.texture)
		gl.Uniform1i(s.shapeTex2DUniformLoc, 0)
	}

	gl.Uniform4fv(s.colorUniformLoc, 1, &s.color[0])
	gl.Uniform3fv(s.originUniformLoc, 1, &s.origin[0])
	gl.Uniform3fv(s.positionUniformLoc, 1, &worldPos[0])
	gl.Uniform1f(s.rotationUniformLoc, worldRot[2])
	s.setShaderScaleUniform(worldScale)

	if s.blurEnabled {
		gl.BlendFunc(gl.ONE, gl.ONE)

		gl.BindFramebuffer(gl.FRAMEBUFFER, s.blurShapeFrameBuffer)

		gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, s.blurShapeTexture, 0)
		gl.ClearColor(0, 0, 0, 0)
		gl.Clear(gl.COLOR_BUFFER_BIT)
		gl.DrawArrays(s.drawMode, 0, s.vertexCount)

		gl.UseProgram(s.blurXShader)
		gl.Uniform1f(s.blurAmountUniformLoc, s.blurIntensity)
		gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, s.blurXTexture, 0)
		gl.ClearColor(0, 0, 0, 0)
		gl.Clear(gl.COLOR_BUFFER_BIT)
		s.renderBlurTexture(s.blurShapeTexture, s.blurTex2DUniformLoc)

		gl.UseProgram(s.blurYShader)
		gl.Uniform1f(s.blurAmountUniformLoc, s.blurIntensity)
		gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, s.blurXYTexture, 0)
		gl.ClearColor(0, 0, 0, 0)
		gl.Clear(gl.COLOR_BUFFER_BIT)
		s.renderBlurTexture(s.blurXTexture, s.blurTex2DUniformLoc)

		gl.UseProgram(s.textureShader)
		gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
		s.renderBlurTexture(s.blurXYTexture, s.textureTex2DUniformLoc)
	} else {
		gl.DrawArrays(s.drawMode, 0, s.vertexCount)
	}

	s.stateMutex.Unlock()

	gl.BindTexture(gl.TEXTURE_2D, 0)
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	gl.BindVertexArray(0)
	gl.UseProgram(0)
	gl.Disable(gl.BLEND)

	return s.WindowObjectBase.drawChildren(deltaTime) && s.label.Draw(deltaTime)
}

func (s *Shape) Close() {
	s.label.Close()
	s.WindowObjectBase.Close()
}

func (s *Shape) Resize(_, _, _, _ int32) {
	s.uninitBlurVao()
	s.initBlurVao()
}

/******************************************************************************
 Shape Functions
******************************************************************************/

func (s *Shape) Texture() string {
	s.stateMutex.Lock()
	tex := s.textureFilename
	s.stateMutex.Unlock()
	return tex
}

func (s *Shape) SetTexture(pathToPng string) *Shape {
	s.stateMutex.Lock()
	s.textureFilename = pathToPng
	s.stateChanged.Store(true)
	s.stateMutex.Unlock()
	s.initVertices()
	s.initVao(true)
	s.initTexture()
	return s
}

func (s *Shape) Sides() uint {
	s.stateMutex.Lock()
	sides := s.sides
	s.stateMutex.Unlock()
	return sides
}

func (s *Shape) SetSides(sides uint) *Shape {
	if sides < 3 {
		sides = 3
	}
	s.stateMutex.Lock()
	s.sides = sides
	s.stateChanged.Store(true)
	s.stateMutex.Unlock()
	s.initVertices()
	s.initVao(true)
	return s
}

func (s *Shape) Thickness() float32 {
	s.stateMutex.Lock()
	thickness := s.thickness
	s.stateMutex.Unlock()
	return thickness
}

func (s *Shape) SetThickness(thickness float32) *Shape {
	if thickness < 0 {
		thickness = 0
	} else if thickness > 1 {
		thickness = 1
	}
	s.stateMutex.Lock()
	s.thickness = thickness
	s.stateChanged.Store(true)
	s.stateMutex.Unlock()
	s.initVertices()
	s.initVao(true)
	return s
}

func (s *Shape) Length() float32 {
	s.stateMutex.Lock()
	length := s.length
	s.stateMutex.Unlock()
	return length
}

func (s *Shape) SetLength(length float32) *Shape {
	if length < 0 {
		length = 0
	} else if length > 1 {
		length = 1
	}
	s.stateMutex.Lock()
	s.length = length
	s.stateChanged.Store(true)
	s.stateMutex.Unlock()
	s.initVertices()
	s.initVao(true)
	return s
}

func (s *Shape) Label() *Label {
	return s.label
}

/******************************************************************************
 New Shape Functions
******************************************************************************/

func newShape(name string) *Shape {
	s := NewShape()
	s.SetName(name)
	return s
}

func NewShape() *Shape {
	lbl := NewLabel()

	p := &Shape{
		WindowObjectBase: *NewObject(nil),
		sides:            3,
		length:           1.0,
		label:            lbl,
	}

	p.SetName(defaultShapeName)
	lbl.SetParent(p)
	return p
}

func NewTriangle(thickness float32) *Shape {
	return newShape(defaultTriangleName).SetThickness(thickness)
}

func NewQuad() *Shape {
	return newShape(defaultQuadName).SetSides(4)
}

func NewSquare(thickness float32) *Shape {
	return newShape(defaultSquareName).SetSides(4).SetThickness(thickness)
}

func NewDot() *Shape {
	return newShape(defaultDotName).SetSides(90)
}

func NewCircle(thickness float32) *Shape {
	return newShape(defaultCircleName).SetSides(90).SetThickness(thickness)
}
