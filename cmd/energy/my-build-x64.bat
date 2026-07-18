SET CGO_ENABLED=0
g use 1.20.14
go version
go mod tidy
set GOARCH=amd64
set GOOS=windows
rm energy-windows64.exe
go build -trimpath -ldflags "-s -w" -o energy-windows64.exe energy.go
g use 1.11.13-386
pause