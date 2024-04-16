package obj

import (
	"fmt"
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
