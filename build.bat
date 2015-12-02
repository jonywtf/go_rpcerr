@echo off

set GOROOT=c:\go
set GOPATH=%~dp0
set GOARCH=amd64
set GOOS=windows

cd %GOPATH%src\

set TARGET_PATH=%GOPATH%bin\gen_err.exe
del %TARGET_PATH%
go build -i -o %TARGET_PATH%

set TARGET_PATH=C:\Windows\System32\gen_err.exe
del %TARGET_PATH%
go build -i -o %TARGET_PATH%

set TARGET_PATH=C:\Windows\SysWOW64\gen_err.exe
del %TARGET_PATH%
go build -i -o %TARGET_PATH%

cd %GOPATH%
