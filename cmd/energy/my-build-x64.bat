SET CGO_ENABLED=0
g use 1.20.14
go version
go mod tidy
set GOARCH=amd64
set GOOS=windows
rm e.exe
go build -trimpath -ldflags "-s -w" -o e.exe energy.go
g use 1.11.13-386
pause