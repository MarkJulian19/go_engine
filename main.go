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
	Cunt_ch    int
)

// Параметры для теней
var (
	depthMapFBO  uint32
	depthMap     uint32
	shadowWidth  = int32(16384) // Очень высокое разрешение (16k!) — возможно, слишком большое.
	shadowHeight = int32(16384)
)

// Параметры для отражения (planar reflection)
var (
	reflectionFBO uint32
	reflectionTex uint32
	reflectWidth  = int32(1024)
	reflectHeight = int32(1024)
	waterLevel    = float32(0.0) // Плоскость, относительно которой зеркалим камеру
)

func init() {
	// GLFW должен работать в главной (main) горутине
	runtime.LockOSThread()
}

func main() {
	config := src.LoadConfig("config.json")
	window := src.InitWindow(config)
	defer glfw.Terminate()

	gl.Enable(gl.DEPTH_TEST)
	// (Опционально) Включаем сглаживание, если нужно:
	// gl.Enable(gl.MULTISAMPLE)

	// Инициализируем шейдеры
	renderProgram := shaders.InitShaders()    // Основной шейдер с тенями и отражениями
	depthProgram := shaders.InitDepthShader() // Шейдер для рендера карты глубины (теней)

	// Создаём FBO и текстуру для карты теней
	createDepthMap()

	// Создаём FBO и текстуру для отражений
	createReflectionFBO()

	// Настраиваем мир и камеру
	worldObj := world.NewWorld(16, 32, 16) // Пример размера чанка
	cameraObj := camera.NewCamera(mgl32.Vec3{0, 60, 0})

	initMouseHandler(window, cameraObj)

	runMainLoop(window, renderProgram, depthProgram, config, worldObj, cameraObj)
}

func initMouseHandler(window *glfw.Window, camera *camera.Camera) {
	window.SetInputMode(glfw.CursorMode, glfw.CursorDisabled)
	go func() {
		for !window.ShouldClose() {
			xpos, ypos := window.GetCursorPos()
			camera.ProcessMouse(xpos, ypos)
			time.Sleep(10 * time.Millisecond)
		}
	}()
}

// ---------------------------------------------------
//     Создание вспомогательных FBO/текстур
// ---------------------------------------------------

// Текстура глубины (shadowMap) + FBO для теней
func createDepthMap() {
	gl.GenFramebuffers(1, &depthMapFBO)

	gl.GenTextures(1, &depthMap)
	gl.BindTexture(gl.TEXTURE_2D, depthMap)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.DEPTH_COMPONENT,
		shadowWidth, shadowHeight, 0,
		gl.DEPTH_COMPONENT, gl.FLOAT, nil)

	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_BORDER)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_BORDER)

	// Цвет границы (при выходе за UV=[0,1]) — обычно белый (означает "нет тени")
	borderColor := []float32{1.0, 1.0, 1.0, 1.0}
	gl.TexParameterfv(gl.TEXTURE_2D, gl.TEXTURE_BORDER_COLOR, &borderColor[0])

	gl.BindFramebuffer(gl.FRAMEBUFFER, depthMapFBO)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.DEPTH_ATTACHMENT, gl.TEXTURE_2D, depthMap, 0)
	gl.DrawBuffer(gl.NONE)
	gl.ReadBuffer(gl.NONE)
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
}

// Текстура и FBO для «зеркального» отражения
func createReflectionFBO() {
	gl.GenFramebuffers(1, &reflectionFBO)
	gl.BindFramebuffer(gl.FRAMEBUFFER, reflectionFBO)

	// Текстура цвета
	gl.GenTextures(1, &reflectionTex)
	gl.BindTexture(gl.TEXTURE_2D, reflectionTex)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA,
		reflectWidth, reflectHeight, 0,
		gl.RGBA, gl.UNSIGNED_BYTE, nil)

	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)

	// Прикрепляем цветовой буфер
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, reflectionTex, 0)

	// Текстура глубины или рендербуфер глубины — на выбор
	var rbo uint32
	gl.GenRenderbuffers(1, &rbo)
	gl.BindRenderbuffer(gl.RENDERBUFFER, rbo)
	gl.RenderbufferStorage(gl.RENDERBUFFER, gl.DEPTH_COMPONENT24, reflectWidth, reflectHeight)
	gl.FramebufferRenderbuffer(gl.FRAMEBUFFER, gl.DEPTH_ATTACHMENT, gl.RENDERBUFFER, rbo)

	// Проверка статуса
	if gl.CheckFramebufferStatus(gl.FRAMEBUFFER) != gl.FRAMEBUFFER_COMPLETE {
		panic("Reflection Framebuffer is not complete!")
	}
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
}
func getLightProjection(playerPos mgl32.Vec3) mgl32.Mat4 {
	// Допустим, хотим ±200 вокруг игрока
	left := playerPos.X() - 200
	right := playerPos.X() + 200
	bottom := playerPos.Z() - 200
	top := playerPos.Z() + 200

	// Или Y можно взять в зависимости от высоты мира
	return mgl32.Ortho(left, right, bottom, top, 0.1, 500.0)
}
func getDynamicLightPos(playerPos mgl32.Vec3) mgl32.Vec3 {
	return mgl32.Vec3{
		playerPos.X() + 100,
		200,
		playerPos.Z() + 100,
	}
}

// ---------------------------------------------------
//
//	Основной цикл
//
// ---------------------------------------------------
func runMainLoop(
	window *glfw.Window,
	renderProgram, depthProgram uint32,
	config *src.Config,
	worldObj *world.World,
	cameraObj *camera.Camera,
) {
	lastFrame := time.Now()

	// Каналы для асинхронной генерации/удаления чанков
	chunkGenCh := make(chan [2]int, 100)
	chunkDelCh := make(chan [2]int, 1000000)
	vramGCCh := make(chan [3]uint32, 1000000)
	for i := 0; i < 16; i++ {
		src.ChunkСreatorWorker(worldObj, chunkGenCh, chunkDelCh, vramGCCh)
		src.ChunkDeleterWorker(worldObj, chunkGenCh, chunkDelCh, vramGCCh)
	}

	// Позиция "солнца" и направление
	// lightPos := mgl32.Vec3{100, 200, 100}
	// lightTarget := mgl32.Vec3{0, 0, 0}
	lightUp := mgl32.Vec3{0, 1, 0}

	// Ортопроекция для теней (чтобы «охватить» сцену)
	// lightProjection := mgl32.Ortho(-200, 200, -200, 200, 0.1, 500.0)

	for !window.ShouldClose() {
		currentFrame := time.Now()
		deltaTime := currentFrame.Sub(lastFrame).Seconds()
		lastFrame = currentFrame

		// Обработка клавиш + физики
		go cameraObj.ProcessKeyboard(window, deltaTime, worldObj)
		cameraObj.UpdatePhysics(deltaTime, worldObj)

		// Обновляем мир (генерация / удаление чанков)
		updateWorld(worldObj, cameraObj, chunkGenCh, chunkDelCh)

		// Освобождаем буферы из VRAM, если надо
		vramGC(vramGCCh)

		dynamicLightPos := getDynamicLightPos(cameraObj.Position)

		// 2) Ортопроекция вокруг игрока
		lightProjection := getLightProjection(cameraObj.Position)

		// 3) Направляем луч из динамического света к игроку
		lightView := mgl32.LookAtV(dynamicLightPos, cameraObj.Position, lightUp)

		// 4) Итоговая матрица
		lightSpaceMatrix := lightProjection.Mul4(lightView)

		// 5) Дальше рендерим depth map, reflection и т.д.:
		renderDepthMap(depthProgram, worldObj, lightSpaceMatrix)
		renderReflection(renderProgram, config, worldObj, cameraObj, lightSpaceMatrix, dynamicLightPos)
		renderScene(window, renderProgram, config, worldObj, cameraObj, lightSpaceMatrix, dynamicLightPos)

		frameCount++
		src.ChangeWindowTitle(&frameCount, &prevTime, window, Cunt_ch)
	}
}

func updateWorld(
	worldObj *world.World,
	cameraObj *camera.Camera,
	chunkGenCh, chunkDelCh chan [2]int,
) {
	playerPos := cameraObj.Position
	// Радиус 32 для прогрузки
	worldObj.UpdateChunks(int(playerPos.X()), int(playerPos.Z()), 32, chunkGenCh, chunkDelCh)
}

// ---------------------------------------------------
//
//	Первый проход: карта глубины (теней)
//
// ---------------------------------------------------
func renderDepthMap(depthProgram uint32, worldObj *world.World, lightSpaceMatrix mgl32.Mat4) {
	gl.Viewport(0, 0, shadowWidth, shadowHeight)
	gl.BindFramebuffer(gl.FRAMEBUFFER, depthMapFBO)
	gl.Clear(gl.DEPTH_BUFFER_BIT)

	gl.UseProgram(depthProgram)
	setUniformMatrix4fv(depthProgram, "lightSpaceMatrix", lightSpaceMatrix)

	worldObj.Mu.Lock()
	for coord, chunk := range worldObj.Chunks {
		// Создаём буферы, если надо
		if chunk.VAO == 0 && len(chunk.Vertices) > 0 && len(chunk.Indices) > 0 {
			safeDeleteBuffers(&chunk.VAO, &chunk.VBO, &chunk.EBO)
			createChunkBuffers(chunk)
			Cunt_ch++
		} else if chunk.UpdateBuf {
			updateChunkBuffers(chunk)
		}

		if chunk.VAO == 0 {
			continue
		}

		// Здесь НЕ делаем isChunkVisible(...) по КАМЕРНОМУ фрустуму!
		// при желании можно сделать culling со стороны света, но НЕ от камеры
		model := mgl32.Translate3D(
			float32(coord[0]*chunk.SizeX),
			0,
			float32(coord[1]*chunk.SizeZ),
		)
		setUniformMatrix4fv(depthProgram, "model", model)

		gl.BindVertexArray(chunk.VAO)
		gl.DrawElements(gl.TRIANGLES, int32(chunk.IndicesCount), gl.UNSIGNED_INT, nil)
	}
	worldObj.Mu.Unlock()

	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
}

// ---------------------------------------------------
//
//	Второй проход: рендер «зеркального» отражения
//
// ---------------------------------------------------
func renderReflection(
	program uint32,
	config *src.Config,
	worldObj *world.World,
	cameraObj *camera.Camera,
	lightSpaceMatrix mgl32.Mat4,
	lightPos mgl32.Vec3,
) {
	// 1) Сохраняем старое положение камеры
	originalPos := cameraObj.Position
	originalPitch := cameraObj.Pitch

	// 2) «Отражаем» камеру по оси Y относительно waterLevel
	distance := (cameraObj.Position.Y() - waterLevel) * 2.0
	cameraObj.Position = mgl32.Vec3{
		cameraObj.Position.X(),
		cameraObj.Position.Y() - distance,
		cameraObj.Position.Z(),
	}
	cameraObj.Pitch = -cameraObj.Pitch

	// 3) Рендер в reflectionFBO
	gl.Viewport(0, 0, reflectWidth, reflectHeight)
	gl.BindFramebuffer(gl.FRAMEBUFFER, reflectionFBO)
	gl.ClearColor(0.0, 0.5, 0.8, 1.0) // фоновый цвет для отражения (небо)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	gl.UseProgram(program)

	// Матрицы вида/проекции с «отражённой» камерой
	view := cameraObj.GetViewMatrix()
	projection := mgl32.Perspective(mgl32.DegToRad(60),
		float32(reflectWidth)/float32(reflectHeight),
		0.1, 3000.0)

	setUniformMatrix4fv(program, "view", view)
	setUniformMatrix4fv(program, "projection", projection)
	setUniformMatrix4fv(program, "lightSpaceMatrix", lightSpaceMatrix)

	// Направление света
	lightDir := lightPos.Sub(mgl32.Vec3{0, 0, 0}).Normalize()
	lightDirLoc := gl.GetUniformLocation(program, gl.Str("lightDir\x00"))
	gl.Uniform3f(lightDirLoc, lightDir.X(), lightDir.Y(), lightDir.Z())

	// Остальные uniform’ы (тени, туман и т.д.)
	setupCommonUniforms(program, cameraObj.Position)

	// Привязываем shadowMap
	gl.ActiveTexture(gl.TEXTURE1)
	gl.BindTexture(gl.TEXTURE_2D, depthMap)
	shadowMapLoc := gl.GetUniformLocation(program, gl.Str("shadowMap\x00"))
	gl.Uniform1i(shadowMapLoc, 1)

	// Отрисовываем чанки (как обычно)
	frustumPlanes := calculateFrustumPlanes(view, projection)

	worldObj.Mu.Lock()
	for coord, chunk := range worldObj.Chunks {
		// Если буферы не созданы / обновлены, создадим:
		if chunk.VAO == 0 && len(chunk.Vertices) > 0 && len(chunk.Indices) > 0 {
			safeDeleteBuffers(&chunk.VAO, &chunk.VBO, &chunk.EBO)
			createChunkBuffers(chunk)
			Cunt_ch++
		} else if chunk.UpdateBuf {
			updateChunkBuffers(chunk)
		}

		// Фрустум-отсечение
		chunkBounds := chunk.GetBoundingBox(coord)
		if !isChunkVisible(frustumPlanes, chunkBounds) {
			continue
		}

		if chunk.VAO == 0 {
			continue
		}

		model := mgl32.Translate3D(
			float32(coord[0]*chunk.SizeX),
			0,
			float32(coord[1]*chunk.SizeZ),
		)
		setUniformMatrix4fv(program, "model", model)

		gl.BindVertexArray(chunk.VAO)
		gl.DrawElements(gl.TRIANGLES, int32(chunk.IndicesCount), gl.UNSIGNED_INT, gl.PtrOffset(0))
	}
	worldObj.Mu.Unlock()

	// 4) Возвращаем камеру на место
	cameraObj.Position = originalPos
	cameraObj.Pitch = originalPitch

	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
}

// ---------------------------------------------------
//
//	Третий проход: основной рендер сцены
//
// ---------------------------------------------------
func renderScene(
	window *glfw.Window,
	program uint32,
	config *src.Config,
	worldObj *world.World,
	cameraObj *camera.Camera,
	lightSpaceMatrix mgl32.Mat4,
	lightPos mgl32.Vec3,
) {
	// Настраиваем вьюпорт под размер окна
	gl.Viewport(0, 0, int32(config.Width), int32(config.Height))
	gl.ClearColor(0.7, 0.8, 1.0, 1.0) // небо
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	gl.UseProgram(program)

	// Матрицы вида и проекции (с нормальной, не-зеркальной камерой)
	view := cameraObj.GetViewMatrix()
	projection := mgl32.Perspective(mgl32.DegToRad(60),
		float32(config.Width)/float32(config.Height),
		0.1, 3000.0)

	setUniformMatrix4fv(program, "view", view)
	setUniformMatrix4fv(program, "projection", projection)
	setUniformMatrix4fv(program, "lightSpaceMatrix", lightSpaceMatrix)

	// Направление света — из lightPos в сцену
	lightDir := lightPos.Sub(mgl32.Vec3{0, 0, 0}).Normalize()
	lightDirLoc := gl.GetUniformLocation(program, gl.Str("lightDir\x00"))
	gl.Uniform3f(lightDirLoc, lightDir.X(), lightDir.Y(), lightDir.Z())

	// Общие uniform’ы (lightColor, ambient, fog, и т.д.)
	setupCommonUniforms(program, cameraObj.Position)

	// Привязываем shadowMap
	gl.ActiveTexture(gl.TEXTURE1)
	gl.BindTexture(gl.TEXTURE_2D, depthMap)
	shadowMapLoc := gl.GetUniformLocation(program, gl.Str("shadowMap\x00"))
	gl.Uniform1i(shadowMapLoc, 1)

	// Привязываем reflectionTex (на TEXTURE2)
	gl.ActiveTexture(gl.TEXTURE2)
	gl.BindTexture(gl.TEXTURE_2D, reflectionTex)
	reflectionMapLoc := gl.GetUniformLocation(program, gl.Str("reflectionMap\x00"))
	gl.Uniform1i(reflectionMapLoc, 2)

	frustumPlanes := calculateFrustumPlanes(view, projection)

	// Рендерим чанки
	worldObj.Mu.Lock()
	for coord, chunk := range worldObj.Chunks {
		// Точно так же: если буфер не создан, создаём
		if chunk.VAO == 0 && len(chunk.Vertices) > 0 && len(chunk.Indices) > 0 {
			safeDeleteBuffers(&chunk.VAO, &chunk.VBO, &chunk.EBO)
			createChunkBuffers(chunk)
			Cunt_ch++
		} else if chunk.UpdateBuf {
			updateChunkBuffers(chunk)
		}

		if !isChunkVisible(frustumPlanes, chunk.GetBoundingBox(coord)) {
			continue
		}

		if chunk.VAO == 0 {
			continue
		}

		model := mgl32.Translate3D(
			float32(coord[0]*chunk.SizeX),
			0,
			float32(coord[1]*chunk.SizeZ),
		)
		setUniformMatrix4fv(program, "model", model)

		gl.BindVertexArray(chunk.VAO)
		gl.DrawElements(gl.TRIANGLES, int32(chunk.IndicesCount), gl.UNSIGNED_INT, gl.PtrOffset(0))
	}
	worldObj.Mu.Unlock()

	window.SwapBuffers()
	glfw.PollEvents()
}

// ---------------------------------------------------
//     Вспомогательные функции
// ---------------------------------------------------

func setupCommonUniforms(program uint32, cameraPos mgl32.Vec3) {
	// Цвет/интенсивность света, ambient и т.д.
	lightColorLoc := gl.GetUniformLocation(program, gl.Str("lightColor\x00"))
	gl.Uniform3f(lightColorLoc, 1.0, 1.0, 1.0)

	ambientColorLoc := gl.GetUniformLocation(program, gl.Str("ambientColor\x00"))
	gl.Uniform3f(ambientColorLoc, 0.2, 0.2, 0.2)

	viewPosLoc := gl.GetUniformLocation(program, gl.Str("viewPos\x00"))
	gl.Uniform3f(viewPosLoc, cameraPos.X(), cameraPos.Y(), cameraPos.Z())

	// Параметры блика
	shininessLoc := gl.GetUniformLocation(program, gl.Str("shininess\x00"))
	gl.Uniform1f(shininessLoc, 32.0)

	specStrLoc := gl.GetUniformLocation(program, gl.Str("specularStrength\x00"))
	gl.Uniform1f(specStrLoc, 0.5)

	// Параметры тумана
	fogStartLoc := gl.GetUniformLocation(program, gl.Str("fogStart\x00"))
	fogEndLoc := gl.GetUniformLocation(program, gl.Str("fogEnd\x00"))
	fogColorLoc := gl.GetUniformLocation(program, gl.Str("fogColor\x00"))

	gl.Uniform1f(fogStartLoc, 50.0)
	gl.Uniform1f(fogEndLoc, 300.0)
	gl.Uniform3f(fogColorLoc, 0.6, 0.7, 0.9)
}

func safeDeleteBuffers(vao, vbo, ebo *uint32) {
	if *vao != 0 {
		gl.DeleteVertexArrays(1, vao)
		*vao = 0
	}
	if *vbo != 0 {
		gl.DeleteBuffers(1, vbo)
		*vbo = 0
	}
	if *ebo != 0 {
		gl.DeleteBuffers(1, ebo)
		*ebo = 0
	}
}

// Создание VAO/VBO/EBO для чанка
func createChunkBuffers(chunk *world.Chunk) {
	var vao, vbo, ebo uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)

	gl.GenBuffers(1, &vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER,
		len(chunk.Vertices)*4, gl.Ptr(chunk.Vertices), gl.STATIC_DRAW)

	gl.GenBuffers(1, &ebo)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, ebo)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER,
		len(chunk.Indices)*4, gl.Ptr(chunk.Indices), gl.STATIC_DRAW)

	// Позиция (0) + нормаль (1) + цвет (2)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 9*4, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(1, 3, gl.FLOAT, false, 9*4, gl.PtrOffset(3*4))
	gl.EnableVertexAttribArray(1)
	gl.VertexAttribPointer(2, 3, gl.FLOAT, false, 9*4, gl.PtrOffset(6*4))
	gl.EnableVertexAttribArray(2)

	chunk.VAO, chunk.VBO, chunk.EBO = vao, vbo, ebo
	chunk.CreateBuf = false
}

func updateChunkBuffers(chunk *world.Chunk) {
	gl.BindVertexArray(chunk.VAO)

	gl.BindBuffer(gl.ARRAY_BUFFER, chunk.VBO)
	gl.BufferSubData(gl.ARRAY_BUFFER, 0,
		len(chunk.Vertices)*4, gl.Ptr(chunk.Vertices))

	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, chunk.EBO)
	gl.BufferSubData(gl.ELEMENT_ARRAY_BUFFER, 0,
		len(chunk.Indices)*4, gl.Ptr(chunk.Indices))

	chunk.UpdateBuf = false
}

func vramGC(vramGCCh chan [3]uint32) {
	for len(vramGCCh) > 0 {
		x := <-vramGCCh
		gl.DeleteVertexArrays(1, &x[0])
		gl.DeleteBuffers(1, &x[1])
		gl.DeleteBuffers(1, &x[2])
		Cunt_ch--
	}
}

func setUniformMatrix4fv(program uint32, name string, matrix mgl32.Mat4) {
	loc := gl.GetUniformLocation(program, gl.Str(name+"\x00"))
	gl.UniformMatrix4fv(loc, 1, false, &matrix[0])
}

func calculateFrustumPlanes(view, projection mgl32.Mat4) [6]mgl32.Vec4 {
	clip := projection.Mul4(view)
	return [6]mgl32.Vec4{
		clip.Row(3).Add(clip.Row(0)), // Left
		clip.Row(3).Sub(clip.Row(0)), // Right
		clip.Row(3).Add(clip.Row(1)), // Bottom
		clip.Row(3).Sub(clip.Row(1)), // Top
		clip.Row(3).Add(clip.Row(2)), // Near
		clip.Row(3).Sub(clip.Row(2)), // Far
	}
}

func isChunkVisible(frustumPlanes [6]mgl32.Vec4, chunkBounds [2]mgl32.Vec3) bool {
	corners := chunkCorners(chunkBounds)
	for _, plane := range frustumPlanes {
		inCount := 0
		for _, c := range corners {
			if plane.X()*c.X()+plane.Y()*c.Y()+plane.Z()*c.Z()+plane.W() > 0 {
				inCount++
			}
		}
		if inCount == 0 {
			return false
		}
	}
	return true
}

func chunkCorners(bounds [2]mgl32.Vec3) [8]mgl32.Vec3 {
	min, max := bounds[0], bounds[1]
	return [8]mgl32.Vec3{
		min,
		{min.X(), min.Y(), max.Z()},
		{min.X(), max.Y(), min.Z()},
		{min.X(), max.Y(), max.Z()},
		{max.X(), min.Y(), min.Z()},
		{max.X(), min.Y(), max.Z()},
		{max.X(), max.Y(), min.Z()},
		max,
	}
}
