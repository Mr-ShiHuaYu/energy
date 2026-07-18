@echo off
echo set env
set GOARCH=386
set CGO_ENABLED=1

set PATH=D:\5CPP\RedPanda-CPP\mingw32\bin;%PATH%;
echo build 
go build -i -ldflags="-s -w" -buildmode=c-shared  -o libenergy.dll

pause
