@echo off
REM Wrapper - Run uninstall.ps1 as Administrator (auto UAC prompt)

net session >nul 2>&1
if %errorLevel% neq 0 (
    echo Mengangkat ke Administrator...
    powershell -Command "Start-Process cmd -Verb RunAs -ArgumentList '/c \"%~f0\"'"
    exit /b
)

powershell -NoProfile -ExecutionPolicy Bypass -File "%~dp0uninstall.ps1"
