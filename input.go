package gfx

import (
	"github.com/go-gl/glfw/v3.3/glfw"
)

type KeyEventHandler struct {
	Key      glfw.Key
	Action   glfw.Action
	Callback func(*Window, glfw.Key, glfw.Action)
}
