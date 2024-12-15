package main

import (
	"engine/camera"
	"engine/shaders"
	"engine/src"
	"engine/world"
	"runtime"
	"time"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

var (
	frameCount int
	prevTime   = float64(0)
)

func init() {
	// GLFW must run on the main thread.
	runtime.LockOSThread()
}

func main() {
	config := src.LoadConfig("config.json")
	window := src.InitWindow(config)
	defer glfw.Terminate()

	gl.Enable(gl.DEPTH_TEST)
	program := shaders.InitShaders()
	world := world.NewWorld(16, 128, 16) // Example size, can be configurable
	camera := camera.NewCamera(mgl32.Vec3{0, 20, 0})
	initMouseHandler(window, camera)

	runMainLoop(window, program, config, world, camera)
}

func initMouseHandler(window *glfw.Window, camera *camera.Camera) {
	window.SetInputMode(glfw.CursorMode, glfw.CursorDisabled)
	// mouseEvents := make(chan [2]float64, 10)
	go func() {
		for !window.ShouldClose() {
			xpos, ypos := window.GetCursorPos()
			camera.ProcessMouse(xpos, ypos)
			// mouseEvents <- [2]float64{xpos, ypos}
			time.Sleep(10 * time.Millisecond)
		}
		// close(mouseEvents)
	}()
	// return mouseEvents
}

func runMainLoop(window *glfw.Window, program uint32, config *src.Config, world *world.World, camera *camera.Camera) {
	lastFrame := time.Now()
	for !window.ShouldClose() {
		currentFrame := time.Now()
		deltaTime := currentFrame.Sub(lastFrame).Seconds()
		lastFrame = currentFrame
		go camera.ProcessKeyboard(window, deltaTime)
		// processInput(window, camera, deltaTime, mouseEvents)
		updateWorld(world, camera)
		renderScene(window, program, config, world, camera)
		frameCount++
		src.ChangeWindowTitle(&frameCount, &prevTime, window)
	}
}

func updateWorld(world *world.World, camera *camera.Camera) {
	playerPos := camera.Position
	world.UpdateChunks(int(playerPos.X()), int(playerPos.Z()), 10) // Adjust radius as needed
}

func renderScene(window *glfw.Window, program uint32, config *src.Config, world *world.World, camera *camera.Camera) {
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	view := camera.GetViewMatrix()
	projection := mgl32.Perspective(mgl32.DegToRad(60), float32(config.Width)/float32(config.Height), 0.1, 1000.0)

	for coord, chunk := range world.Chunks {
		// vertices, indices := chunk.GenerateMesh(nil) // Pass neighbors if needed
		// vao := makeVAO(vertices, indices)

		model := mgl32.Translate3D(
			float32(coord[0]*chunk.SizeX),
			0,
			float32(coord[1]*chunk.SizeZ),
		)

		setUniformMatrix4fv(program, "model", model)
		setUniformMatrix4fv(program, "view", view)
		setUniformMatrix4fv(program, "projection", projection)

		gl.BindVertexArray(chunk.VAO)
		gl.DrawElements(gl.TRIANGLES, int32(chunk.IndicesCount), gl.UNSIGNED_INT, gl.PtrOffset(0))
	}

	window.SwapBuffers()
	glfw.PollEvents()
}

func setUniformMatrix4fv(program uint32, name string, matrix mgl32.Mat4) {
	location := gl.GetUniformLocation(program, gl.Str(name+"\x00"))
	gl.UniformMatrix4fv(location, 1, false, &matrix[0])
}
