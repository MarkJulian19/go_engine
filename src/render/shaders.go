package render

import (
	"fmt"
	"log"

	"github.com/go-gl/gl/v4.1-core/gl"
)

// Инициализируем основной шейдер (с тенями и отражением)
func InitShaders() uint32 {
	program, err := CompileRenderShaders()
	if err != nil {
		log.Fatalln("Error compiling main render shaders:", err)
	}
	gl.UseProgram(program)
	return program
}

// Инициализируем шейдер для рендера карты глубины (теней)
func InitDepthShader() uint32 {
	program, err := CompileDepthShader()
	if err != nil {
		log.Fatalln("Error compiling depth shaders:", err)
	}
	gl.UseProgram(program)
	return program
}

func InitCrosshairShader() uint32 {
	program, err := CompileCrosshairShader()
	if err != nil {
		log.Fatalln("Error compiling crosshair shaders:", err)
	}
	gl.UseProgram(program)
	return program
}

func InitTextShader() uint32 {
	program, err := compileTextShader()
	if err != nil {
		log.Fatalln("Error compiling text shaders:", err)
	}
	gl.UseProgram(program)
	return program
}

/*
-----------------------------------------------------------------------------

	ШЕЙДЕР ГЛУБИНЫ (Depth Shader) — рендерим только глубину в gl_FragDepth

-----------------------------------------------------------------------------
*/
func CompileDepthShader() (uint32, error) {
	vertexShaderSrc := `#version 410 core

layout(location = 0) in vec3 inPosition;
layout(location = 1) in vec3 inNormal;
layout(location = 2) in vec3 inColor;

uniform mat4 lightSpaceMatrix;
uniform mat4 model;

void main()
{
    vec4 worldPos = model * vec4(inPosition, 1.0);
    gl_Position = lightSpaceMatrix * worldPos;
}
` + "\x00"

	fragmentShaderSrc := `#version 410 core
void main()
{
    // Здесь выводим только глубину
    // gl_FragDepth обновляется автоматически
}
` + "\x00"

	return compileProgram(vertexShaderSrc, fragmentShaderSrc)
}

/*
-----------------------------------------------------------------------------

	Основной шейдер для рендера (с тенями + отражение)

-----------------------------------------------------------------------------
*/
func CompileRenderShaders() (uint32, error) {
	// Vertex Shader
	vertexShaderSrc := `#version 410 core

layout(location = 0) in vec3 inPosition;
layout(location = 1) in vec3 inNormal;   
layout(location = 2) in vec3 inColor;    

out vec3 fragPos;         
out vec3 fragNormal;      
out vec3 fragColor;       
out float fragDist;       
out vec4 fragPosLightSpace;

// Для water отражения (простейший вариант: отдадим координаты, чтобы генерировать UV)
out vec3 reflectionCoords;

uniform mat4 model;
uniform mat4 view;
uniform mat4 projection;
uniform mat4 lightSpaceMatrix;

void main()
{
    vec4 worldPos = model * vec4(inPosition, 1.0);
    fragPos = worldPos.xyz;

    // Преобразуем нормаль (учитывая модельное преобразование)
    fragNormal = mat3(transpose(inverse(model))) * inNormal;

    fragColor = inColor;

    vec4 viewPos = view * worldPos;
    fragDist = length(viewPos.xyz);

    fragPosLightSpace = lightSpaceMatrix * worldPos;

    // Для простого варианта отражения возьмём XZ как UV (или любую проекцию)
    reflectionCoords = fragPos;

    gl_Position = projection * viewPos;
}
` + "\x00"

	// Fragment Shader
	fragmentShaderSrc := `#version 410 core

in vec3 fragPos;
in vec3 fragNormal;
in vec3 fragColor;
in float fragDist;
in vec4 fragPosLightSpace;
in vec3 reflectionCoords;

out vec4 outputColor;

// === ПАРАМЕТРЫ ОСВЕЩЕНИЯ ===
uniform vec3 lightDir;       
uniform vec3 lightColor;     
uniform vec3 ambientColor;   
uniform vec3 viewPos;        
uniform float shininess;     
uniform float specularStrength; 

// === ПАРАМЕТРЫ ТУМАНА ===
uniform float fogStart;
uniform float fogEnd;
uniform vec3 fogColor;

// === ТЕНИ ===
uniform sampler2D shadowMap;

// === ОТРАЖЕНИЕ (зеркальная карта) ===
uniform sampler2D reflectionMap;

// Простая функция shadow mapping (без PCF)
float calculateShadow(vec4 fragPosLightSpace, vec3 normal, vec3 lightDir)
{
    vec3 projCoords = fragPosLightSpace.xyz / fragPosLightSpace.w;
    projCoords = projCoords * 0.5 + 0.5; // Приводим координаты в диапазон [0, 1]

    // Проверка выхода за пределы shadow map
    if (projCoords.x < 0.0 || projCoords.x > 1.0 ||
        projCoords.y < 0.0 || projCoords.y > 1.0 ||
        projCoords.z > 1.0)
    {
        return 0.0; // Нет теней вне shadow map
    }

    float closestDepth = texture(shadowMap, projCoords.xy).r; // Глубина из shadow map
    float currentDepth = projCoords.z;

    // Динамический bias для устранения самошейдинга
    float bias = max(0.00005 * (1.0 - dot(normal, lightDir)), 0.00001);
    //float bias = 0.00001; 
    // Применение PCF (Percentage Closer Filtering)
    float shadow = 0.0;
    vec2 texelSize = 1.0 / textureSize(shadowMap, 0); 
    // Увеличиваем диапазон шага - например, 2, а не 1:
    for (int x = -2; x <= 2; ++x) {
        for (int y = -2; y <= 2; ++y) {
            float pcfDepth = texture(shadowMap, projCoords.xy + vec2(x, y) * texelSize).r;
            shadow += currentDepth - bias > pcfDepth ? 1.0 : 0.0;
        }
    }
    // Теперь получим 25 выборок вместо 9.
    shadow /= 25.0;

    return shadow;
}

void main()
{
    // (1) Освещение
    vec3 N = normalize(fragNormal);
    vec3 L = normalize(lightDir);
    vec3 V = normalize(viewPos - fragPos);

    // Ambient
    vec3 ambient = ambientColor * fragColor;

    // Diffuse
    float diff = max(dot(N, L), 0.0);
    vec3 diffuse = diff * lightColor * fragColor;

    // Specular (Blinn-Phong)
    vec3 H = normalize(L + V);
    float specAngle = max(dot(N, H), 0.0);
    float spec = pow(specAngle, shininess);
    vec3 specular = specularStrength * spec * lightColor;

    // (2) Тени
    float shadow = calculateShadow(fragPosLightSpace, N, L);
    vec3 lightingColor = ambient + (1.0 - shadow) * (diffuse + specular);

    // (3) Проверка «материала»: предположим, что inColor = (0,0.2,1.0) (синий) → "вода"
    //     Можно сделать любую логику в движке, где вы определяете воду по ID блока и т.п.
    //     Ниже — чисто демонстрационный пример:
    bool isWater = (fragColor.r < 0.01 && fragColor.g > 0.15 && fragColor.b > 0.9);

    // Если это вода, семплим reflectionMap
    if(isWater) {
        // Считаем простые UV: XZ из reflectionCoords
        // Нормализуем их в [0..1], как-то примитивно. 
        // В реальном проекте надо проектировать правильно с учётом камеры/углов.
        vec2 uv = reflectionCoords.xz * 0.01; // масштаб 0.01
        // Примитивно «заворачиваем» в 0..1
        uv = fract(uv);

        // Цвет «зеркала»
        vec3 mirrorColor = texture(reflectionMap, uv).rgb;

        // Смешиваем часть «отражения» с частью «цвета воды»
        // Коэффициент «прозрачности» воды
        float waterReflectFactor = 0.5;
        vec3 waterColor = mix(lightingColor, mirrorColor, waterReflectFactor);

        // (4) Туман
        float fogFactor = clamp((fogEnd - fragDist) / (fogEnd - fogStart), 0.0, 1.0);
        vec3 finalColor = mix(fogColor, waterColor, fogFactor);

        outputColor = vec4(finalColor, 1.0);
        return;
    }

    // (3') Если это не вода — обычный Blinn-Phong с тенями
    // Туман
    float fogFactor = clamp((fogEnd - fragDist) / (fogEnd - fogStart), 0.0, 1.0);
    vec3 finalColor = mix(fogColor, lightingColor, fogFactor);

    outputColor = vec4(finalColor, 1.0);
}
` + "\x00"
	return compileProgram(vertexShaderSrc, fragmentShaderSrc)
}

func CompileCrosshairShader() (uint32, error) {
	vertexShaderSrc := `#version 410 core

layout(location = 0) in vec3 inPosition;

uniform mat4 ortho;

void main()
{
    gl_Position = ortho * vec4(inPosition, 1.0);
}
` + "\x00"

	fragmentShaderSrc := `#version 410 core

uniform vec4 crosshairColor;

out vec4 fragColor;

void main()
{
    fragColor = crosshairColor;
}
` + "\x00"

	return compileProgram(vertexShaderSrc, fragmentShaderSrc)
}

func compileTextShader() (uint32, error) {
	vertexShaderSrc := `#version 410 core

layout(location = 0) in vec3 inPosition; // Позиция вершины
layout(location = 1) in vec2 inTexCoord; // Координаты текстуры

uniform mat4 ortho; // Ортографическая матрица

out vec2 fragTexCoord; // Передача координат текстуры в фрагментный шейдер

void main()
{
    fragTexCoord = inTexCoord;
    gl_Position = ortho * vec4(inPosition, 1.0);
}
` + "\x00"

	fragmentShaderSrc := `#version 410 core

in vec2 fragTexCoord; // Координаты текстуры из вершинного шейдера
uniform sampler2D textTexture; // Текстура с текстом
uniform vec4 textColor; // Цвет текста (RGBA)

out vec4 fragColor; // Итоговый цвет фрагмента

void main()
{
    vec4 sampled = texture(textTexture, fragTexCoord);
    fragColor = vec4(textColor.rgb, sampled.a * textColor.a);
}
` + "\x00"

	return compileProgram(vertexShaderSrc, fragmentShaderSrc)
}

// ---------------------------------------------------
//
//	Вспомогательные функции компиляции
//
// ---------------------------------------------------
func compileProgram(vShaderSource, fShaderSource string) (uint32, error) {
	vs, err := CompileShader(vShaderSource, gl.VERTEX_SHADER)
	if err != nil {
		return 0, err
	}
	fs, err := CompileShader(fShaderSource, gl.FRAGMENT_SHADER)
	if err != nil {
		return 0, err
	}

	program := gl.CreateProgram()
	gl.AttachShader(program, vs)
	gl.AttachShader(program, fs)
	gl.LinkProgram(program)

	var success int32
	gl.GetProgramiv(program, gl.LINK_STATUS, &success)
	if success == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(program, gl.INFO_LOG_LENGTH, &logLength)
		logMsg := make([]byte, logLength+1)
		gl.GetProgramInfoLog(program, logLength, nil, &logMsg[0])
		return 0, fmt.Errorf("program linking error: %s", string(logMsg))
	}

	gl.DeleteShader(vs)
	gl.DeleteShader(fs)

	return program, nil
}

func CompileShader(src string, shaderType uint32) (uint32, error) {
	shader := gl.CreateShader(shaderType)
	csrc, free := gl.Strs(src)
	defer free()
	gl.ShaderSource(shader, 1, csrc, nil)
	gl.CompileShader(shader)

	var success int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &success)
	if success == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)
		logMsg := make([]byte, logLength+1)
		gl.GetShaderInfoLog(shader, logLength, nil, &logMsg[0])
		return 0, fmt.Errorf("shader compilation error: %s", string(logMsg))
	}
	return shader, nil
}
