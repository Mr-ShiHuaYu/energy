@echo off
echo set env
set GOARCH=amd64
set CGO_ENABLED=1
set PATH=%mingw64%;%mingw64%\bin;%PATH%;
echo build 
go build -ldflags="-s -w" -buildmode=c-shared  -o libenergy.dll
 
pause
