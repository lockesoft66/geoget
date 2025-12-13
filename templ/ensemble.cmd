@echo off
setlocal
set SCRIPT_DIR=%~dp0
set BASEBOX_DIR=%SCRIPT_DIR%basebox
set USER_CONFIG_FILE=%BASEBOX_DIR%\basebox.conf

if exist "%BASEBOX_DIR%\binnt\basebox.exe" (
    set BASEBOX_EXEC=%BASEBOX_DIR%\binnt\basebox.exe
) else if exist "%BASEBOX_DIR%\binl64\basebox" (
    set BASEBOX_EXEC=%BASEBOX_DIR%\binl64\basebox
) else if exist "%BASEBOX_DIR%\binl\basebox" (
    set BASEBOX_EXEC=%BASEBOX_DIR%\binl\basebox
) else if exist "%BASEBOX_DIR%\binmac\basebox" (
    set BASEBOX_EXEC=%BASEBOX_DIR%\binmac\basebox
) else (
    echo Error: Unable to locate the Basebox executable.
    exit /b 1
)

"%BASEBOX_EXEC%" -noprimaryconf -nolocalconf -conf "%USER_CONFIG_FILE%" %*
