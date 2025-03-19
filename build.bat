@echo off
REM Set environment variables for CGO
set CGO_ENABLED=1
set CC=C:\msys64\ucrt64\bin\gcc.exe

REM Print environment info
echo CGO_ENABLED=%CGO_ENABLED%
echo CC=%CC%

REM Build the project
go build -o ha-mi.exe cmd/server/main.go

echo Build completed! 