package render

import (
	"image"
	"image/draw"
	"log"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font/gofont/goregular"
)

func RenderText(text string, x, y int, screenWidth, screenHeight int, program uint32, color [4]float32) {
	// Загрузим шрифт
	font, err := truetype.Parse(goregular.TTF)
	if err != nil {
		log.Fatalf("failed to parse font: %v", err)
	}

	// Создаем изображение для текста
	img := image.NewRGBA(image.Rect(0, 0, 512, 512))
	draw.Draw(img, img.Bounds(), image.Transparent, image.Point{}, draw.Src)

	// Настраиваем контекст для рендеринга текста
	c := freetype.NewContext()
	c.SetDPI(72)
	c.SetFont(font)
	c.SetFontSize(24)
	c.SetClip(img.Bounds())
	c.SetDst(img)
	c.SetSrc(image.White)

	// Рендерим текст
	pt := freetype.Pt(10, 10+int(c.PointToFixed(24)>>6))
	_, err = c.DrawString(text, pt)
	if err != nil {
		log.Fatalf("failed to draw string: %v", err)
	}

	// Создаем текстуру из изображения
	var texture uint32
	gl.GenTextures(1, &texture)
	gl.BindTexture(gl.TEXTURE_2D, texture)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, int32(img.Bounds().Dx()), int32(img.Bounds().Dy()), 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(img.Pix))
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)

	// Вычисляем ортографическую матрицу
	ortho := mgl32.Ortho2D(0, float32(screenWidth), float32(screenHeight), 0)

	// Передаём матрицу и цвет текста в шейдер
	gl.UseProgram(program)

	orthoLoc := gl.GetUniformLocation(program, gl.Str("ortho\x00"))
	gl.UniformMatrix4fv(orthoLoc, 1, false, &ortho[0])

	colorLoc := gl.GetUniformLocation(program, gl.Str("textColor\x00"))
	gl.Uniform4fv(colorLoc, 1, &color[0])

	// Рендерим текстуру
	renderTexturedQuad(texture, float32(x), float32(y), float32(img.Bounds().Dx()), float32(img.Bounds().Dy()))

	// Удаляем текстуру после использования
	gl.DeleteTextures(1, &texture)
}

func renderTexturedQuad(texture uint32, x, y, width, height float32) {
	vertices := []float32{
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

	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 5*4, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(1, 2, gl.FLOAT, false, 5*4, gl.PtrOffset(3*4))
	gl.EnableVertexAttribArray(1)

	gl.BindTexture(gl.TEXTURE_2D, texture)

	gl.DrawElements(gl.TRIANGLES, int32(len(indices)), gl.UNSIGNED_INT, gl.PtrOffset(0))

	gl.BindVertexArray(0)
	gl.DeleteBuffers(1, &vbo)
	gl.DeleteBuffers(1, &ebo)
	gl.DeleteVertexArrays(1, &vao)
}
