@echo off
REM =============================================================================
REM Sirus Print Agent - One-Click Setup Wrapper
REM Klik kanan file ini -> Run as administrator
REM =============================================================================

REM Self-elevate: kalau belum admin, restart dengan UAC prompt
net session >nul 2>&1
if %errorLevel% neq 0 (
    echo Mengangkat ke Administrator...
    powershell -Command "Start-Process cmd -Verb RunAs -ArgumentList '/c \"%~f0\"'"
    exit /b
)

REM Run PowerShell installer (sudah admin di sini)
powershell -NoProfile -ExecutionPolicy Bypass -File "%~dp0setup.ps1"
