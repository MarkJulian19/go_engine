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
	world := world.NewWorld(16, 256, 16) // Example size, can be configurable
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
func StartChunkListener(w *world.World, genCh, delCh <-chan [2]int, vramCh chan [3]uint32) {
	go func() {
		for {
			select {
			case coords := <-genCh:
				// Если приходят координаты для генерации
				x, z := coords[0], coords[1]
				w.GenerateChunk(x, z)

			case coords := <-delCh:
				// Если приходят координаты для удаления
				x, z := coords[0], coords[1]
				w.RemoveChunk(x, z, vramCh)
			}
		}
	}()
}
func runMainLoop(window *glfw.Window, program uint32, config *src.Config, world *world.World, camera *camera.Camera) {
	lastFrame := time.Now()
	chunkGenCh := make(chan [2]int, 1000000)
	chunkDelCh := make(chan [2]int, 1000000)
	vramGCCh := make(chan [3]uint32, 1000000)
	StartChunkListener(world, chunkGenCh, chunkDelCh, vramGCCh)
	StartChunkListener(world, chunkGenCh, chunkDelCh, vramGCCh)
	StartChunkListener(world, chunkGenCh, chunkDelCh, vramGCCh)
	StartChunkListener(world, chunkGenCh, chunkDelCh, vramGCCh)
	StartChunkListener(world, chunkGenCh, chunkDelCh, vramGCCh)
	StartChunkListener(world, chunkGenCh, chunkDelCh, vramGCCh)

	for !window.ShouldClose() {
		currentFrame := time.Now()
		deltaTime := currentFrame.Sub(lastFrame).Seconds()
		lastFrame = currentFrame
		go camera.ProcessKeyboard(window, deltaTime)
		// processInput(window, camera, deltaTime, mouseEvents)
		// fmt.Print("updateWorld")
		updateWorld(world, camera, chunkGenCh, chunkDelCh)

		renderScene(window, program, config, world, camera)
		vramGC(vramGCCh)
		frameCount++
		src.ChangeWindowTitle(&frameCount, &prevTime, window)
	}
}

func updateWorld(world *world.World, camera *camera.Camera, chunkGenCh chan [2]int, chunkDelCh chan [2]int) {
	playerPos := camera.Position
	world.UpdateChunks(int(playerPos.X()), int(playerPos.Z()), 20, chunkGenCh, chunkDelCh) // Adjust radius as needed
}

func renderScene(window *glfw.Window, program uint32, config *src.Config, world *world.World, camera *camera.Camera) {
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	// startTime := time.Now()
	view := camera.GetViewMatrix()
	projection := mgl32.Perspective(mgl32.DegToRad(60), float32(config.Width)/float32(config.Height), 0.1, 1000.0)

	// world.Mu.Lock()
	// tmp_world := world.Chunks
	// world.Mu.Unlock()
	counter := 0
	world.Mu.Lock()

	for coord, chunk := range world.Chunks {
		// vertices, indices := chunk.GenerateMesh(nil) // Pass neighbors if needed
		// vao := makeVAO(vertices, indices)

		if counter < 80 {
			if chunk.CreateBuf {
				// Создание новых буферов
				if chunk.VAO != 0 {
					gl.DeleteVertexArrays(1, &chunk.VAO)
				}
				if chunk.VBO != 0 {
					gl.DeleteBuffers(1, &chunk.VBO)
				}
				if chunk.EBO != 0 {
					gl.DeleteBuffers(1, &chunk.EBO)
				}

				var vao, vbo, ebo uint32
				gl.GenVertexArrays(1, &vao)
				gl.BindVertexArray(vao)

				gl.GenBuffers(1, &vbo)
				gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
				gl.BufferData(gl.ARRAY_BUFFER, len(chunk.Vertices)*4, gl.Ptr(chunk.Vertices), gl.STATIC_DRAW)

				gl.GenBuffers(1, &ebo)
				gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, ebo)
				gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(chunk.Indices)*4, gl.Ptr(chunk.Indices), gl.STATIC_DRAW)

				gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 6*4, gl.PtrOffset(0))
				gl.EnableVertexAttribArray(0)
				gl.VertexAttribPointer(1, 3, gl.FLOAT, false, 6*4, gl.PtrOffset(3*4))
				gl.EnableVertexAttribArray(1)

				chunk.VAO, chunk.VBO, chunk.EBO = vao, vbo, ebo
				chunk.CreateBuf = false
			} else if chunk.UpdateBuf {
				// Обновление данных в существующих буферах
				gl.BindVertexArray(chunk.VAO)

				gl.BindBuffer(gl.ARRAY_BUFFER, chunk.VBO)
				gl.BufferSubData(gl.ARRAY_BUFFER, 0, len(chunk.Vertices)*4, gl.Ptr(chunk.Vertices))

				gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, chunk.EBO)
				gl.BufferSubData(gl.ELEMENT_ARRAY_BUFFER, 0, len(chunk.Indices)*4, gl.Ptr(chunk.Indices))
			}

			chunk.UpdateBuf = false
			counter++
		}

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
	world.Mu.Unlock()
	window.SwapBuffers()
	glfw.PollEvents()
	// elapsedTime := time.Since(startTime)
	// fmt.Printf("renderScene executed in %v\n", elapsedTime)
}

func vramGC(vramGCCh chan [3]uint32) {
	//fmt.Println("ROFL")
	for len(vramGCCh) > 0 {
		x := <-vramGCCh
		gl.DeleteVertexArrays(1, &x[0])
		gl.DeleteBuffers(1, &x[1])
		gl.DeleteBuffers(1, &x[2])
	}
}

func setUniformMatrix4fv(program uint32, name string, matrix mgl32.Mat4) {
	location := gl.GetUniformLocation(program, gl.Str(name+"\x00"))
	gl.UniformMatrix4fv(location, 1, false, &matrix[0])
}
