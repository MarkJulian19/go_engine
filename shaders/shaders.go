package shaders

import (
	"fmt"
	"log"

	"github.com/go-gl/gl/v4.1-core/gl"
)

func InitShaders() uint32 {
	program, err := CompileShaders()
	if err != nil {
		log.Fatalln("Error compiling shaders:", err)
	}
	gl.UseProgram(program)
	return program
}

func CompileShaders() (uint32, error) {
	vertexShaderSrc := "#version 410 core\n" +
		"layout(location = 0) in vec3 position;\n" +
		"layout(location = 1) in vec3 color;\n" +
		"out vec3 fragColor;\n" +
		"uniform mat4 model;\n" +
		"uniform mat4 view;\n" +
		"uniform mat4 projection;\n" +
		"void main() {\n" +
		"  fragColor = color;\n" +
		"  gl_Position = projection * view * model * vec4(position, 1.0);\n" +
		"}\n"

	fragmentShaderSrc := "#version 410 core\n" +
		"in vec3 fragColor;\n" +
		"out vec4 outputColor;\n" +
		"void main() {\n" +
		"  outputColor = vec4(fragColor, 1.0);\n" +
		"}\n"

	vertexShader, err := CompileShader(vertexShaderSrc, gl.VERTEX_SHADER)
	if err != nil {
		return 0, err
	}
	fragmentShader, err := CompileShader(fragmentShaderSrc, gl.FRAGMENT_SHADER)
	if err != nil {
		return 0, err
	}

	program := gl.CreateProgram()
	gl.AttachShader(program, vertexShader)
	gl.AttachShader(program, fragmentShader)
	gl.LinkProgram(program)

	var success int32
	gl.GetProgramiv(program, gl.LINK_STATUS, &success)
	if success == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(program, gl.INFO_LOG_LENGTH, &logLength)

		log := make([]byte, logLength+1)
		gl.GetProgramInfoLog(program, logLength, nil, &log[0])
		return 0, fmt.Errorf("program linking error: %s", string(log))
	}

	gl.DeleteShader(vertexShader)
	gl.DeleteShader(fragmentShader)

	return program, nil
}
func CompileShader(src string, shaderType uint32) (uint32, error) {
	shader := gl.CreateShader(shaderType)
	csrc, free := gl.Strs(src + "\x00")
	defer free()
	gl.ShaderSource(shader, 1, csrc, nil)
	gl.CompileShader(shader)

	var success int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &success)
	if success == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)

		log := make([]byte, logLength+1)
		gl.GetShaderInfoLog(shader, logLength, nil, &log[0])
		return 0, fmt.Errorf("shader compilation error: %s", string(log))
	}

	return shader, nil
}
