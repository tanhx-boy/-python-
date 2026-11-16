@echo off
setlocal
cd /d "%~dp0"

if not exist dist mkdir dist

set CGO_ENABLED=0
go build -trimpath -ldflags "-s -w -H=windowsgui" -o dist\FileChecker.exe .
if errorlevel 1 (
    echo.
    echo Build failed.
    pause
    exit /b 1
)

echo.
echo Built dist\FileChecker.exe
pause
