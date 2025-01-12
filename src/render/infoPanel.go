package render

import (
	"engine/src/player"
	"engine/src/world"
	"fmt"
	"image"
	"image/draw"
	"log"
	"runtime"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font/gofont/goregular"
)

// RenderDebugHUD отрисовывает отладочную информацию в HUD.
func RenderDebugHUD(window *glfw.Window, program uint32, debugInfo []string) {
	width, height := window.GetSize()

	// Включаем альфа-смешивание (прозрачность)
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	// Отключаем тест глубины, чтобы текст всегда отображался поверх
	gl.Disable(gl.DEPTH_TEST)

	// Устанавливаем ортографическую проекцию
	orthoProjection := mgl32.Ortho(0, float32(width), float32(height), 0, -1, 1)

	gl.UseProgram(program)

	// Установка униформ переменной ортографической матрицы
	orthoLoc := gl.GetUniformLocation(program, gl.Str("ortho\x00"))
	gl.UniformMatrix4fv(orthoLoc, 1, false, &orthoProjection[0])

	// Установка цвета текста
	color := [4]float32{1.0, 1.0, 1.0, 1.0} // Белый цвет
	colorLoc := gl.GetUniformLocation(program, gl.Str("textColor\x00"))
	gl.Uniform4fv(colorLoc, 1, &color[0])

	// Привязываем текстурный юнит к тексту
	textureLoc := gl.GetUniformLocation(program, gl.Str("textTexture\x00"))
	gl.Uniform1i(textureLoc, 0) // TEXTURE0

	// Рендерим каждую строку отладочной информации
	for i, line := range debugInfo {
		RenderText(line, 5, 15*(i+1), width, height, program, color)
	}

	// Восстанавливаем настройки рендера
	gl.Enable(gl.DEPTH_TEST)
	gl.Disable(gl.BLEND)
}

// RenderText отрисовывает текстовую строку на экране.
func RenderText(text string, x, y, screenWidth, screenHeight int, program uint32, color [4]float32) {
	font, err := truetype.Parse(goregular.TTF)
	if err != nil {
		log.Fatalf("failed to parse font: %v", err)
	}

	// Создаём изображение для текста
	img := image.NewRGBA(image.Rect(0, 0, 512, 512))
	draw.Draw(img, img.Bounds(), image.Transparent, image.Point{}, draw.Src)

	c := freetype.NewContext()
	c.SetDPI(72)
	c.SetFont(font)
	c.SetFontSize(15)
	c.SetClip(img.Bounds())
	c.SetDst(img)
	c.SetSrc(image.White)

	// Отрисовываем текст в изображение
	pt := freetype.Pt(10, 10+int(c.PointToFixed(24)>>6))
	_, err = c.DrawString(text, pt)
	if err != nil {
		log.Fatalf("failed to draw string: %v", err)
	}

	// Создаём текстуру OpenGL
	var texture uint32
	gl.GenTextures(1, &texture)
	defer gl.DeleteTextures(1, &texture)
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, texture)

	gl.TexImage2D(
		gl.TEXTURE_2D, 0, gl.RGBA,
		int32(img.Bounds().Dx()), int32(img.Bounds().Dy()),
		0, gl.RGBA, gl.UNSIGNED_BYTE,
		gl.Ptr(img.Pix),
	)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)

	// Рендерим прямоугольник с текстурой
	renderTexturedQuad(texture, float32(x), float32(y), float32(img.Bounds().Dx()), float32(img.Bounds().Dy()))

	// Удаляем текстуру после рендера
	gl.DeleteTextures(1, &texture)
}

// renderTexturedQuad отрисовывает прямоугольник с текстурой.
func renderTexturedQuad(texture uint32, x, y, width, height float32) {
	vertices := []float32{
		// Положение (x,y,z)    // Текс-коорд (u,v)
		x, y, 0.0, 0.0, 0.0, // нижний левый
		x + width, y, 0.0, 1.0, 0.0, // нижний правый
		x + width, y + height, 0.0, 1.0, 1.0, // верхний правый
		x, y + height, 0.0, 0.0, 1.0, // верхний левый
	}

	indices := []uint32{
		0, 1, 2, // первый треугольник
		2, 3, 0, // второй треугольник
	}

	var vao, vbo, ebo uint32
	gl.GenVertexArrays(1, &vao)
	gl.GenBuffers(1, &vbo)
	gl.GenBuffers(1, &ebo)

	gl.BindVertexArray(vao)

	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, ebo)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(indices)*4, gl.Ptr(indices), gl.STATIC_DRAW)

	// Позиция (location = 0)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 5*4, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(0)

	// Текстурные координаты (location = 1)
	gl.VertexAttribPointer(1, 2, gl.FLOAT, false, 5*4, gl.PtrOffset(3*4))
	gl.EnableVertexAttribArray(1)

	// Привязка текстуры (активный юнит уже TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, texture)

	// Рисуем
	gl.DrawElements(gl.TRIANGLES, int32(len(indices)), gl.UNSIGNED_INT, gl.PtrOffset(0))

	// Освобождаем
	gl.BindVertexArray(0)
	gl.DeleteBuffers(1, &vbo)
	gl.DeleteBuffers(1, &ebo)
	gl.DeleteVertexArrays(1, &vao)
}
func Get_hud_info(deltaTime float64, worldObj *world.World, cameraObj *player.Camera) []string {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// Замер видеопамяти
	var totalVRAM, availableVRAM int32
	gl.GetIntegerv(0x9048 /* GL_GPU_MEMORY_INFO_TOTAL_AVAILABLE_MEMORY_NVX */, &totalVRAM)       // Общее количество VRAM
	gl.GetIntegerv(0x9049 /* GL_GPU_MEMORY_INFO_CURRENT_AVAILABLE_VIDMEM_NVX */, &availableVRAM) // Доступная VRAM
	var usedVRAM int32
	if totalVRAM > 0 {
		usedVRAM = totalVRAM - availableVRAM
	}
	return []string{
		fmt.Sprintf("FPS: %.2f", 1.0/deltaTime),
		fmt.Sprintf("Camera Position: X=%.2f Y=%.2f Z=%.2f", cameraObj.Position.X(), cameraObj.Position.Y(), cameraObj.Position.Z()),
		fmt.Sprintf("Chunks Loaded: %d", len(worldObj.Chunks)),
		fmt.Sprintf("Allocated RAM: %.2f MB", float64(memStats.Alloc)/1024/1024),
		fmt.Sprintf("Total Allocated RAM: %.2f MB", float64(memStats.TotalAlloc)/1024/1024),
		fmt.Sprintf("System RAM: %.2f MB", float64(memStats.Sys)/1024/1024),
		fmt.Sprintf("Total VRAM: %.2f MB", float64(totalVRAM)/1024),
		fmt.Sprintf("Used VRAM: %.2f MB", float64(usedVRAM)/1024),
	}
}
