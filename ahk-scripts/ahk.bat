@echo off
REM AutoHotkey Startup Script
REM Replace the paths below with your actual script locations

REM Start CapsLock Ctrl/Esc script
start "" "D:\projects\config\ahk-scripts\CapsCtrl-TapEscape.ahk"

REM Wait a moment before starting the second script
timeout /t 1 /nobreak >nul

REM Start Double Shift CapsLock toggle script
start "" "D:\projects\config\ahk-scripts\DoubleShift-CapsToggle.ahk"

REM Optional: Show confirmation
echo AutoHotkey scripts started successfully!
timeout /t 2 /nobreak >nul