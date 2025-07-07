@echo off
setlocal enabledelayedexpansion
if not exist "peakpip.exe" (
    echo error: peakpip.exe not found in current directory
    echo please run this script from the directory containing peakpip.exe
    pause
    exit /b 1
)
set "CURRENT_DIR=%cd%"
echo %PATH% | findstr /i "%CURRENT_DIR%" >nul
if %errorlevel% == 0 (
    echo peakpip is already in PATH
    goto :test_install
)
echo adding peakpip to PATH...
for /f "tokens=2*" %%a in ('reg query "HKCU\Environment" /v PATH 2^>nul') do set "USER_PATH=%%b"
if "%USER_PATH%"=="" (
    reg add "HKCU\Environment" /v PATH /t REG_EXPAND_SZ /d "%CURRENT_DIR%" /f >nul
) else (
    reg add "HKCU\Environment" /v PATH /t REG_EXPAND_SZ /d "%USER_PATH%;%CURRENT_DIR%" /f >nul
)
if %errorlevel% == 0 (
    echo successfully added peakpip to PATH
    echo.
    echo important: you need to restart your command prompt or terminal
    echo for the changes to take effect, or run:
    echo   set PATH=%%PATH%%;%CURRENT_DIR%
) else (
    echo failed to add peakpip to PATH
    echo you may need to run this script as administrator
    pause
    exit /b 1
)
:test_install
echo.
echo testing peakpip installation...
peakpip
if %errorlevel% == 0 (
    echo.
    echo peakpip is working correctly!!
    echo you can now use peakpip !!!
) else (
    echo.
    echo peakpip test failed.. you may need to restart your terminal..
)
echo.
echo installation complete!!!
pause
