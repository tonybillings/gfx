package gfx

import (
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"math"
	"sync/atomic"
)

const (
	defaultShape2DName  = "Shape2D"
	defaultTriangleName = "Triangle"
	defaultQuadName     = "Quad"
	defaultSquareName   = "Square"
	defaultDotName      = "Dot"
	defaultCircleName   = "Circle"

	thicknessEpsilon = 0.00001
)

/******************************************************************************
 Shape2D
******************************************************************************/

type Shape2D struct {
	WindowObjectBase

	viewport *Viewport

	vao uint32
	vbo uint32

	shapeShader    Shader
	texShapeShader Shader
	blurXShader    Shader
	blurYShader    Shader
	textureShader  Shader

	shapeShaderBinding    *ShaderBinding
	texShapeShaderBinding *ShaderBinding
	boundStruct           *shaderTransform

	posAttribLoc         int32
	posTexShapeAttribLoc int32
	uvTexShapeAttribLoc  int32

	shapeColorUniformLoc    int32
	texShapeColorUniformLoc int32
	shapeTexUniformLoc      int32
	blurAmountUniformLoc    int32
	blurTex1UniformLoc      int32
	blurTex2UniformLoc      int32

	texture Texture

	blurShapeFrameBuffer uint32
	blurShapeTexture     uint32
	blurXTexture         uint32
	blurXYTexture        uint32
	blurTextureVao       uint32
	blurTextureVbo       uint32
	blurTextureVertices  []float32

	vertices    []float32
	vertexCount int32

	sides     uint
	thickness float32
	length    float32
	drawMode  uint32

	uvFlipped bool

	viewportBak [4]int32

	stateChanged atomic.Bool
}

/******************************************************************************
 Object Implementation
******************************************************************************/

func (s *Shape2D) Init() (ok bool) {
	if s.Initialized() {
		return true
	}

	s.init()
	s.RefreshLayout()

	return s.WindowObjectBase.Init()
}

func (s *Shape2D) Update(deltaTime int64) (ok bool) {
	if !s.WindowObjectBase.Update(deltaTime) {
		return false
	}

	if s.stateChanged.Load() {
		s.reinit()
	}

	return true
}

func (s *Shape2D) Close() {
	s.closeShapeVao()

	if s.blurEnabled {
		s.closeBlurVao()
	}

	s.closeBindings()

	s.WindowObjectBase.Close()
}

/******************************************************************************
 DrawableObject Implementation
******************************************************************************/

func (s *Shape2D) Draw(deltaTime int64) (ok bool) {
	if !s.DrawableObjectBase.Draw(deltaTime) {
		return false
	}

	s.boundStruct.setOrigin(s.WorldOrigin())
	s.boundStruct.setPosition(s.WorldPosition())
	s.boundStruct.setRotation(s.WorldRotation())
	s.boundStruct.setScale(s.getAdjustedScale(s.WorldScale()))

	s.beginDraw()

	if s.texture == nil {
		s.shapeShader.Activate()
		gl.ActiveTexture(gl.TEXTURE0)
		gl.BindTexture(gl.TEXTURE_2D, 0)
		gl.Uniform4fv(s.shapeColorUniformLoc, 1, &s.color[0])
		s.shapeShaderBinding.Update(deltaTime)
	} else {
		s.texShapeShader.Activate()
		gl.ActiveTexture(gl.TEXTURE0)
		gl.BindTexture(gl.TEXTURE_2D, s.texture.GlName())
		gl.Uniform1i(s.shapeTexUniformLoc, 0)
		gl.Uniform4fv(s.texShapeColorUniformLoc, 1, &s.color[0])
		s.texShapeShaderBinding.Update(deltaTime)
	}

	if s.blurEnabled {
		gl.BlendFunc(gl.ONE, gl.ONE)

		gl.BindFramebuffer(gl.FRAMEBUFFER, s.blurShapeFrameBuffer)

		gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, s.blurShapeTexture, 0)
		gl.ClearColor(0, 0, 0, 0)
		gl.Clear(gl.COLOR_BUFFER_BIT)
		gl.DrawArrays(s.drawMode, 0, s.vertexCount)

		s.blurXShader.Activate()
		gl.Uniform1f(s.blurAmountUniformLoc, s.blurIntensity)
		gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, s.blurXTexture, 0)
		gl.ClearColor(0, 0, 0, 0)
		gl.Clear(gl.COLOR_BUFFER_BIT)
		s.renderBlurTexture(s.blurShapeTexture, s.blurTex1UniformLoc)

		s.blurYShader.Activate()
		gl.Uniform1f(s.blurAmountUniformLoc, s.blurIntensity)
		gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, s.blurXYTexture, 0)
		gl.ClearColor(0, 0, 0, 0)
		gl.Clear(gl.COLOR_BUFFER_BIT)
		s.renderBlurTexture(s.blurXTexture, s.blurTex1UniformLoc)

		s.textureShader.Activate()
		gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
		s.renderBlurTexture(s.blurXYTexture, s.blurTex2UniformLoc)
	} else {
		gl.DrawArrays(s.drawMode, 0, s.vertexCount)
	}

	s.endDraw()

	return s.WindowObjectBase.drawChildren(deltaTime)
}

/******************************************************************************
 Resizer Implementation
******************************************************************************/

func (s *Shape2D) Resize(oldWidth, oldHeight, newWidth, newHeight int32) {
	if s.viewport != nil {
		s.viewport.SetWindowSize(newWidth, newHeight)
	}

	if s.blurEnabled {
		s.stateMutex.Lock()
		s.closeBlurVao()
		s.initBlurVao()
		s.stateMutex.Unlock()
	}
	s.WindowObjectBase.Resize(oldWidth, oldHeight, newWidth, newHeight)
}

/******************************************************************************
 Shape2D Functions
******************************************************************************/

func (s *Shape2D) initViewport() {
	if s.viewport == nil {
		s.viewport = NewViewport(s.window.Width(), s.window.Height())
	}
}

func (s *Shape2D) initVerticesTriangle() {
	a := [2]float32{0, 1}
	b := [2]float32{-.87, -.5}
	c := [2]float32{.87, -.5}

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
}

func (s *Shape2D) initVerticesTriangleWithUV() {
	a := [2]float32{0, 1}
	b := [2]float32{-.87, -.5}
	c := [2]float32{.87, -.5}

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
		if s.uvFlipped {
			s.vertices = []float32{
				a[0], a[1], 0.5, 1.0,
				b[0], b[1], 0.0, 0.0,
				c[0], c[1], 1.0, 0.0,
			}
		} else {
			s.vertices = []float32{
				a[0], a[1], 0.5, 0.0,
				b[0], b[1], 0.0, 1.0,
				c[0], c[1], 1.0, 1.0,
			}
		}

		s.vertexCount = 3
		s.drawMode = gl.TRIANGLES
	}
}

func (s *Shape2D) initVerticesQuadrangle() {
	xAspectRatioInv := s.WorldScale().Y() / s.WorldScale().X()

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
}

func (s *Shape2D) initVerticesQuadrangleWithUV() {
	xAspectRatioInv := s.WorldScale().Y() / s.WorldScale().X()

	topLeft := [2]float32{-1.0, 1.0}
	bottomRight := [2]float32{1.0, -1.0}

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
		if s.uvFlipped {
			s.vertices = []float32{
				-1.0, 1.0, 0.0, 0.0,
				1.0, 1.0, 1.0, 0.0,
				-1.0, -1.0, 0.0, 1.0,
				1.0, 1.0, 1.0, 0.0,
				1.0, -1.0, 1.0, 1.0,
				-1.0, -1.0, 0.0, 1.0,
			}
		} else {
			s.vertices = []float32{
				-1.0, 1.0, 0.0, 1.0,
				1.0, 1.0, 1.0, 1.0,
				-1.0, -1.0, 0.0, 0.0,
				1.0, 1.0, 1.0, 1.0,
				1.0, -1.0, 1.0, 0.0,
				-1.0, -1.0, 0.0, 0.0,
			}
		}

		s.vertexCount = 6
		s.drawMode = gl.TRIANGLES
	}
}

func (s *Shape2D) initVerticesPolygon() {
	radius := float32(1)
	maxAngle := 2 * math.Pi * float64(s.length)
	angleIncrement := maxAngle / float64(s.sides)

	if maxAngle < thicknessEpsilon {
		s.vertices = make([]float32, 0)
		s.vertexCount = 0
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
}

func (s *Shape2D) initVerticesPolygonWithUV() {
	radius := float32(1)
	maxAngle := 2 * math.Pi * float64(s.length)
	angleIncrement := maxAngle / float64(s.sides)

	if maxAngle < thicknessEpsilon {
		s.vertices = make([]float32, 0)
		s.vertexCount = 0
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
			if s.uvFlipped {
				vertices = append(vertices, x, y, (x+1)/2, 1-(y+1)/2)
			} else {
				vertices = append(vertices, x, y, (x+1)/2, (y+1)/2)
			}
		}

		x := float32(math.Cos(maxAngle))
		y := float32(math.Sin(maxAngle))
		if s.uvFlipped {
			vertices = append(vertices, x, y, (x+1)/2, 1-(y+1)/2)
		} else {
			vertices = append(vertices, x, y, (x+1)/2, (y+1)/2)
		}

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
}

func (s *Shape2D) initVertices() {
	switch s.sides {
	case 3:
		if s.texture == nil {
			s.initVerticesTriangle()
		} else {
			s.initVerticesTriangleWithUV()
		}
	case 4:
		if s.texture == nil {
			s.initVerticesQuadrangle()
		} else {
			s.initVerticesQuadrangleWithUV()
		}
	default:
		if s.texture == nil {
			s.initVerticesPolygon()
		} else {
			s.initVerticesPolygonWithUV()
		}
	}
}

func (s *Shape2D) initBlurTextureVertices() {
	s.blurTextureVertices = []float32{
		-1.0, 1.0, 0.0, 1.0,
		1.0, 1.0, 1.0, 1.0,
		-1.0, -1.0, 0.0, 0.0,
		1.0, 1.0, 1.0, 1.0,
		1.0, -1.0, 1.0, 0.0,
		-1.0, -1.0, 0.0, 0.0,
	}
}

func (s *Shape2D) initShaders() {
	s.shapeShader = Assets.Get(Shape2DShader).(Shader)
	s.texShapeShader = Assets.Get(TexturedShape2DShader).(Shader)
	s.blurXShader = Assets.Get(BlurXShader).(Shader)
	s.blurYShader = Assets.Get(BlurYShader).(Shader)
	s.textureShader = Assets.Get(TextureShader).(Shader)

	s.posAttribLoc = s.shapeShader.GetAttribLocation("a_Position")
	s.posTexShapeAttribLoc = s.texShapeShader.GetAttribLocation("a_Position")
	s.uvTexShapeAttribLoc = s.texShapeShader.GetAttribLocation("a_UV")

	s.shapeColorUniformLoc = s.shapeShader.GetUniformLocation("u_Color")
	s.texShapeColorUniformLoc = s.texShapeShader.GetUniformLocation("u_Color")
	s.shapeTexUniformLoc = s.texShapeShader.GetUniformLocation("u_DiffuseMap")
	s.blurAmountUniformLoc = s.blurXShader.GetUniformLocation("u_BlurAmount")
	s.blurTex1UniformLoc = s.blurXShader.GetUniformLocation("u_TextureMap")
	s.blurTex2UniformLoc = s.textureShader.GetUniformLocation("u_TextureMap")

	s.shapeShaderBinding = NewShaderBinding(s.shapeShader, s.boundStruct, nil)
	s.shapeShaderBinding.Init()

	s.texShapeShaderBinding = NewShaderBinding(s.texShapeShader, s.boundStruct, nil)
	s.texShapeShaderBinding.Init()
}

func (s *Shape2D) initShapeVao() {
	gl.GenVertexArrays(1, &s.vao)
	gl.GenBuffers(1, &s.vbo)

	gl.BindVertexArray(s.vao)
	gl.BindBuffer(gl.ARRAY_BUFFER, s.vbo)

	stride := int32(0)

	if s.texture == nil {
		gl.EnableVertexAttribArray(0)
		gl.VertexAttribPointerWithOffset(uint32(s.posAttribLoc), 2, gl.FLOAT, false, stride, 0)
	} else {
		stride = 16
		gl.EnableVertexAttribArray(0)
		gl.VertexAttribPointerWithOffset(uint32(s.posTexShapeAttribLoc), 2, gl.FLOAT, false, stride, 0)
		gl.EnableVertexAttribArray(1)
		gl.VertexAttribPointerWithOffset(uint32(s.uvTexShapeAttribLoc), 2, gl.FLOAT, false, stride, uintptr(2*sizeOfFloat32))
	}

	if len(s.vertices) > 0 {
		gl.BufferData(gl.ARRAY_BUFFER, len(s.vertices)*sizeOfFloat32, gl.Ptr(s.vertices), gl.STATIC_DRAW)
	}

	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	gl.BindVertexArray(0)
}

func (s *Shape2D) initBlurVao() {
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

	gl.ActiveTexture(gl.TEXTURE0)

	gl.GenTextures(1, &s.blurShapeTexture)
	gl.BindTexture(gl.TEXTURE_2D, s.blurShapeTexture)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, s.Window().Width(), s.Window().Height(), 0, gl.RGBA, gl.UNSIGNED_BYTE, nil)

	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, s.blurShapeTexture, 0)

	gl.GenTextures(1, &s.blurXTexture)
	gl.BindTexture(gl.TEXTURE_2D, s.blurXTexture)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, s.Window().Width(), s.Window().Height(), 0, gl.RGBA, gl.UNSIGNED_BYTE, nil)

	gl.GenTextures(1, &s.blurXYTexture)
	gl.BindTexture(gl.TEXTURE_2D, s.blurXYTexture)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, s.Window().Width(), s.Window().Height(), 0, gl.RGBA, gl.UNSIGNED_BYTE, nil)

	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	gl.BindVertexArray(0)
}

func (s *Shape2D) reinit() {
	s.stateMutex.Lock()

	s.initViewport()
	s.initVertices()
	s.closeShapeVao()
	s.initShapeVao()

	if s.blurEnabled {
		s.closeBlurVao()
		s.initBlurVao()
	}

	if s.texture != nil && !s.texture.Initialized() {
		s.texture.Init()
	}

	s.stateChanged.Store(false)
	s.stateMutex.Unlock()
}

func (s *Shape2D) init() {
	s.stateMutex.Lock()

	s.initViewport()
	s.initVertices()
	s.initBlurTextureVertices()
	s.initShaders()
	s.closeShapeVao()
	s.initShapeVao()

	if s.blurEnabled {
		s.closeBlurVao()
		s.initBlurVao()
	}

	if s.texture != nil && !s.texture.Initialized() {
		s.texture.Init()
	}

	s.stateChanged.Store(false)
	s.stateMutex.Unlock()
}

func (s *Shape2D) closeShapeVao() {
	gl.BindVertexArray(0)
	gl.DeleteVertexArrays(1, &s.vao)
	s.vao = 0

	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	gl.DeleteBuffers(1, &s.vbo)
	s.vbo = 0
}

func (s *Shape2D) closeBlurVao() {
	gl.BindVertexArray(0)
	gl.DeleteVertexArrays(1, &s.blurTextureVao)

	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	gl.DeleteBuffers(1, &s.blurTextureVbo)

	gl.DeleteFramebuffers(1, &s.blurShapeFrameBuffer)

	gl.DeleteTextures(1, &s.blurShapeTexture)
	gl.DeleteTextures(1, &s.blurXTexture)
	gl.DeleteTextures(1, &s.blurXYTexture)
}

func (s *Shape2D) closeBindings() {
	if s.shapeShaderBinding != nil {
		s.shapeShaderBinding.Close()
	}

	if s.texShapeShaderBinding != nil {
		s.texShapeShaderBinding.Close()
	}
}

func (s *Shape2D) beginDraw() {
	s.stateMutex.Lock()

	gl.BindVertexArray(s.vao)
	gl.BindBuffer(gl.ARRAY_BUFFER, s.vbo)

	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	gl.GetIntegerv(gl.VIEWPORT, &s.viewportBak[0])
	gl.Viewport(s.viewport.Get())
}

func (s *Shape2D) endDraw() {
	s.stateMutex.Unlock()

	gl.Viewport(s.viewportBak[0], s.viewportBak[1], s.viewportBak[2], s.viewportBak[3])

	gl.Disable(gl.BLEND)

	gl.BindTexture(gl.TEXTURE_2D, 0)
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	gl.BindVertexArray(0)

	gl.UseProgram(0)
}

func (s *Shape2D) getAdjustedScale(worldScale mgl32.Vec3) (scale mgl32.Vec3) {
	if s.maintainAspectRatio {
		width := float32(s.window.width.Load())
		height := float32(s.window.height.Load())

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
	} else {
		scale[0] = worldScale[0]
		scale[1] = worldScale[1]
	}

	return
}

func (s *Shape2D) renderBlurTexture(texture uint32, tex2DLoc int32) {
	gl.BindVertexArray(s.blurTextureVao)
	gl.BindBuffer(gl.ARRAY_BUFFER, s.blurTextureVbo)
	gl.BindTexture(gl.TEXTURE_2D, texture)
	gl.Uniform1i(tex2DLoc, 0)
	gl.DrawArrays(gl.TRIANGLES, 0, 6)
}

func (s *Shape2D) SetTexture(texture Texture) *Shape2D {
	s.stateMutex.Lock()
	s.texture = texture
	s.stateChanged.Store(true)
	s.stateMutex.Unlock()
	return s
}

func (s *Shape2D) Sides() uint {
	s.stateMutex.Lock()
	sides := s.sides
	s.stateMutex.Unlock()
	return sides
}

func (s *Shape2D) SetSides(sides uint) *Shape2D {
	if sides < 3 {
		sides = 3
	}
	s.stateMutex.Lock()
	s.sides = sides
	s.stateChanged.Store(true)
	s.stateMutex.Unlock()
	return s
}

func (s *Shape2D) Thickness() float32 {
	s.stateMutex.Lock()
	thickness := s.thickness
	s.stateMutex.Unlock()
	return thickness
}

func (s *Shape2D) SetThickness(thickness float32) *Shape2D {
	if thickness < 0 {
		thickness = 0
	} else if thickness > 1 {
		thickness = 1
	}
	s.stateMutex.Lock()
	s.thickness = thickness
	s.stateChanged.Store(true)
	s.stateMutex.Unlock()
	return s
}

func (s *Shape2D) Length() float32 {
	s.stateMutex.Lock()
	length := s.length
	s.stateMutex.Unlock()
	return length
}

func (s *Shape2D) SetLength(length float32) *Shape2D {
	if length < 0 {
		length = 0
	} else if length > 1 {
		length = 1
	}
	s.stateMutex.Lock()
	s.length = length
	s.stateChanged.Store(true)
	s.stateMutex.Unlock()
	return s
}

func (s *Shape2D) Viewport() *Viewport {
	s.stateMutex.Lock()
	vp := s.viewport
	s.stateMutex.Unlock()
	return vp
}

func (s *Shape2D) SetViewport(viewport *Viewport) *Shape2D {
	s.stateMutex.Lock()
	s.viewport = viewport
	s.stateMutex.Unlock()
	return s
}

func (s *Shape2D) FlipUV(flipped bool) *Shape2D {
	s.stateMutex.Lock()
	s.uvFlipped = flipped
	s.stateMutex.Unlock()
	return s
}

/******************************************************************************
 New Shape2D Functions
******************************************************************************/

func newShape(name string) *Shape2D {
	s := NewShape2D()
	s.SetName(name)
	return s
}

func NewShape2D() *Shape2D {
	p := &Shape2D{
		WindowObjectBase: *NewWindowObject(),
		sides:            3,
		length:           1.0,
		boundStruct: &shaderTransform{
			Transform: &ShaderTransform{},
		},
	}

	p.SetName(defaultShape2DName)
	return p
}

func NewTriangle(thickness float32) *Shape2D {
	return newShape(defaultTriangleName).SetThickness(thickness)
}

func NewQuad() *Shape2D {
	return newShape(defaultQuadName).SetSides(4)
}

func NewSquare(thickness float32) *Shape2D {
	return newShape(defaultSquareName).SetSides(4).SetThickness(thickness)
}

func NewDot() *Shape2D {
	return newShape(defaultDotName).SetSides(90)
}

func NewCircle(thickness float32) *Shape2D {
	return newShape(defaultCircleName).SetSides(90).SetThickness(thickness)
}
