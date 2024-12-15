$env:CGO_ENABLED = 1
$env:GOOS = "windows"
$env:GOARCH = "amd64"
$env:CGO_CFLAGS = "-Wno-error"

go build main.go