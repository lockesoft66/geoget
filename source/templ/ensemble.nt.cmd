@echo off
setlocal
set SCRIPT_DIR=%~dp0
set BASEBOX_DIR=%SCRIPT_DIR%basebox
set BASEBOX_EXEC=%BASEBOX_DIR%\binnt\basebox.exe
set USER_CONFIG_FILE=%BASEBOX_DIR%\basebox.conf

if not exist "%BASEBOX_EXEC%" (
    echo Error: Expected Basebox executable not found at "%BASEBOX_EXEC%".
    exit /b 1
)

"%BASEBOX_EXEC%" -noprimaryconf -nolocalconf -conf "%USER_CONFIG_FILE%" %*
