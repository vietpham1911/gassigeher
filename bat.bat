@echo off
REM Gassigeher - Build and Test Script for Windows
REM Usage: bat.bat

echo ========================================
echo Gassigeher - Build and Test
echo ========================================
echo.

REM Check if Go is installed
where go >nul 2>nul
if %ERRORLEVEL% NEQ 0 (
    echo [ERROR] Go is not installed or not in PATH
    exit /b 1
)

echo [1/4] Checking Go version...
go version
echo.

echo [2/4] Downloading dependencies...
go mod download
if %ERRORLEVEL% NEQ 0 (
    echo [ERROR] Failed to download dependencies
    exit /b 1
)
echo [OK] Dependencies downloaded
echo.

echo [3/4] Building application...
go build -o gassigeher.exe cmd/server/main.go
if %ERRORLEVEL% NEQ 0 (
    echo [ERROR] Build failed
    exit /b 1
)
echo [OK] Build successful: gassigeher.exe
echo.

echo [4/4] Running tests...
go test -v -cover ./...
if %ERRORLEVEL% NEQ 0 (
    echo [WARNING] Some tests failed
    echo.
) else (
    echo [OK] All tests passed
    echo.
)

echo ========================================
echo Build and Test Complete!
echo ========================================
echo.
echo To run the application:
echo   .\gassigeher.exe
echo.
echo To run with custom port:
echo   set PORT=3000 ^&^& .\gassigeher.exe
echo.
