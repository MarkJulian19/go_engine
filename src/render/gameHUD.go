package render

import (
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

// RenderCrosshair отрисовывает простое перекрестие в центре экрана.
func RenderCrosshair(window *glfw.Window, program uint32) {
	width, height := window.GetSize()

	// Включаем альфа-смешивание (прозрачность)
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	// Отключаем тест глубины, чтобы перекрестие всегда отображалось поверх
	gl.Disable(gl.DEPTH_TEST)

	// Устанавливаем ортографическую проекцию
	orthoProjection := mgl32.Ortho(0, float32(width), float32(height), 0, -1, 1)

	gl.UseProgram(program)

	// Установка униформ переменной ортографической матрицы
	orthoLoc := gl.GetUniformLocation(program, gl.Str("ortho\x00"))
	gl.UniformMatrix4fv(orthoLoc, 1, false, &orthoProjection[0])

	// Установка цвета перекрестия
	color := [4]float32{1.0, 1.0, 1.0, 1.0} // Белый цвет
	colorLoc := gl.GetUniformLocation(program, gl.Str("crosshairColor\x00"))
	gl.Uniform4fv(colorLoc, 1, &color[0])

	// Рендерим перекрестие
	renderCrosshairQuad(float32(width)/2, float32(height)/2, 10, 2)

	// Восстанавливаем настройки рендера
	gl.Enable(gl.DEPTH_TEST)
	gl.Disable(gl.BLEND)
}

// renderCrosshairQuad отрисовывает перекрестие в виде двух пересекающихся линий.
func renderCrosshairQuad(x, y, size, thickness float32) {
	vertices := []float32{
		// Горизонтальная линия
		x - size, y, 0.0,
		x + size, y, 0.0,
		x, y - thickness, 0.0,
		x, y + thickness, 0.0,

		// Вертикальная линия
		x, y - size, 0.0,
		x, y + size, 0.0,
		x - thickness, y, 0.0,
		x + thickness, y, 0.0,
	}

	indices := []uint32{
		0, 1, 2, 3, // Горизонтальная линия
		4, 5, 6, 7, // Вертикальная линия
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
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 3*4, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(0)

	// Рисуем
	gl.DrawElements(gl.LINES, int32(len(indices)), gl.UNSIGNED_INT, gl.PtrOffset(0))

	// Освобождаем
	gl.BindVertexArray(0)
	gl.DeleteBuffers(1, &vbo)
	gl.DeleteBuffers(1, &ebo)
	gl.DeleteVertexArrays(1, &vao)
}
