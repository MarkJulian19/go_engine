package src

import (
	"fmt"
	"log"
	"runtime"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

func CreateWindow(config *Config) (*glfw.Window, error) {
	if err := glfw.Init(); err != nil {
		return nil, fmt.Errorf("GLFW initialization error: %w", err)
	}

	glfw.WindowHint(glfw.ContextVersionMajor, config.ContextVersionMajor)
	glfw.WindowHint(glfw.ContextVersionMinor, config.ContextVersionMinor)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	window, err := glfw.CreateWindow(config.Width, config.Height, config.Title, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("window initialization error: %w", err)
	}
	window.MakeContextCurrent()
	if err := gl.Init(); err != nil {
		return nil, fmt.Errorf("OpenGL initialization error: %w", err)
	}
	return window, nil
}

func ChangeWindowTitle(frameCount *int, prevTime *float64, window *glfw.Window) {
	currentTime := glfw.GetTime()
	if currentTime-*prevTime >= 0.1 {
		var memStats runtime.MemStats
		runtime.ReadMemStats(&memStats)
		allocatedMemory := memStats.Alloc / 1024 / 1024 // в МБ

		fps := *frameCount * 10
		title := fmt.Sprintf("FPS: %d | Memory: %d MB", fps, allocatedMemory)
		window.SetTitle(title)

		*frameCount = 0
		*prevTime = currentTime
	}
}
func InitWindow(config *Config) *glfw.Window {
	window, err := CreateWindow(config)
	if err != nil {
		log.Fatalln("Error creating window:", err)
	}
	return window
}
