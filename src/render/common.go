package render

import (
	"engine/src/world"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

var (
	Cunt_ch int
)

func setUniformMatrix4fv(program uint32, name string, matrix mgl32.Mat4) {
	loc := gl.GetUniformLocation(program, gl.Str(name+"\x00"))
	gl.UniformMatrix4fv(loc, 1, false, &matrix[0])
}

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
	gl.Uniform1f(fogEndLoc, 200.0)
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
