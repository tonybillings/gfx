package obj

import (
	"fmt"
	"github.com/go-gl/mathgl/mgl32"
	"strconv"
	"strings"
)

/******************************************************************************
 Parsing Functions
******************************************************************************/

func parseString(fields []string) (string, error) {
	if len(fields) < 1 {
		return "", fmt.Errorf("value not provided")
	}
	return fields[0], nil
}

func parseVec2(components []string) ([2]float32, error) {
	if len(components) < 2 {
		return [2]float32{}, fmt.Errorf("not enough components for [2]float32")
	}

	var u, v float64
	var err error

	u, err = strconv.ParseFloat(components[0], 64)
	if err != nil {
		return [2]float32{}, err
	}

	v, err = strconv.ParseFloat(components[1], 64)
	if err != nil {
		return [2]float32{}, err
	}

	return [2]float32{float32(u), float32(v)}, nil
}

func parseVec3(components []string) ([3]float32, error) {
	if len(components) < 3 {
		return [3]float32{}, fmt.Errorf("not enough components for [3]float32")
	}

	var x, y, z float64
	var err error

	x, err = strconv.ParseFloat(components[0], 64)
	if err != nil {
		return [3]float32{}, err
	}

	y, err = strconv.ParseFloat(components[1], 64)
	if err != nil {
		return [3]float32{}, err
	}

	z, err = strconv.ParseFloat(components[2], 64)
	if err != nil {
		return [3]float32{}, err
	}

	return [3]float32{float32(x), float32(y), float32(z)}, nil
}

func parseVec4(components []string) ([4]float32, error) {
	if vec3, err := parseVec3(components); err != nil {
		return [4]float32{}, err
	} else {
		return [4]float32{vec3[0], vec3[1], vec3[2]}, nil
	}
}

func parseFloat(components []string) (float32, error) {
	if len(components) < 1 {
		return 0, fmt.Errorf("not enough components for float")
	}

	value, err := strconv.ParseFloat(components[0], 64)
	return float32(value), err
}

func parseFace(components []string) (*Face, error) {
	var face Face
	for _, part := range components {
		var index int
		var err error

		indices := strings.Split(part, "/")
		index, err = strconv.Atoi(indices[0])
		if err != nil {
			return &Face{}, fmt.Errorf("invalid vertex index: %v", err)
		}
		face.vertices = append(face.vertices, index-1)

		if len(indices) > 1 && indices[1] != "" {
			index, err = strconv.Atoi(indices[1])
			if err != nil {
				return &Face{}, fmt.Errorf("invalid texture coordinate index: %v", err)
			}
			face.uvs = append(face.uvs, index-1)
		}

		if len(indices) == 3 && indices[2] != "" {
			index, err = strconv.Atoi(indices[2])
			if err != nil {
				return &Face{}, fmt.Errorf("invalid normal index: %v", err)
			}
			face.normals = append(face.normals, index-1)
		}
	}

	return &face, nil
}

func setFaceTangents(model *Model, face *Face) {
	v0Idx := face.vertices[0] * 3
	v1Idx := face.vertices[1] * 3
	v2Idx := face.vertices[2] * 3

	v0 := mgl32.Vec3{model.vertices[v0Idx], model.vertices[v0Idx+1], model.vertices[v0Idx+2]}
	v1 := mgl32.Vec3{model.vertices[v1Idx], model.vertices[v1Idx+1], model.vertices[v1Idx+2]}
	v2 := mgl32.Vec3{model.vertices[v2Idx], model.vertices[v2Idx+1], model.vertices[v2Idx+2]}

	deltaPos1 := v1.Sub(v0)
	deltaPos2 := v2.Sub(v0)

	uv0Idx := face.uvs[0] * 2
	uv1Idx := face.uvs[1] * 2
	uv2Idx := face.uvs[2] * 2

	uv0 := mgl32.Vec2{model.uvs[uv0Idx], model.uvs[uv0Idx+1]}
	uv1 := mgl32.Vec2{model.uvs[uv1Idx], model.uvs[uv1Idx+1]}
	uv2 := mgl32.Vec2{model.uvs[uv2Idx], model.uvs[uv2Idx+1]}

	deltaUV1 := uv1.Sub(uv0)
	deltaUV2 := uv2.Sub(uv0)

	d := deltaUV1.X()*deltaUV2.Y() - deltaUV1.Y()*deltaUV2.X()
	if d == 0 {
		return
	}

	r := 1.0 / d
	tan := ((deltaPos1.Mul(deltaUV2.Y()).Sub(deltaPos2.Mul(deltaUV1.Y()))).Mul(r)).Normalize()
	bitan := ((deltaPos2.Mul(deltaUV1.X()).Sub(deltaPos1.Mul(deltaUV2.X()))).Mul(r)).Normalize()

	model.tangents = append(model.tangents, []float32{tan[0], tan[1], tan[2]}...)
	model.bitangents = append(model.bitangents, []float32{bitan[0], bitan[1], bitan[2]}...)

	tanIdx := (len(model.tangents) / 3) - 1
	bitanIdx := (len(model.bitangents) / 3) - 1

	face.tangents = []int{tanIdx, tanIdx, tanIdx, tanIdx}
	face.bitangents = []int{bitanIdx, bitanIdx, bitanIdx, bitanIdx}
}
