@echo off

set GOROOT=c:\go
set GOPATH=%~dp0
set GOARCH=amd64
set GOOS=windows

cd %GOPATH%

set TARGET_PATH=%GOPATH%bin\gen_err.exe
del %TARGET_PATH%
cd %GOPATH%src\
go build -i -o %TARGET_PATH%
cd %GOPATH%
