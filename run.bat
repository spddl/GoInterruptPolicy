@echo off

SET filename=GoInterruptPolicy

net session >nul 2>&1
set nilIsAdmin=%errorLevel%

:loop
cls

if %nilIsAdmin% EQU 0 (
    go run -tags debug -buildvcs=false .
) else (
    go build -tags debug -buildvcs=false -o %filename%_debug.exe
    IF %ERRORLEVEL% EQU 0 %filename%_debug.exe
)

@REM IF %ERRORLEVEL% EQU 0 %filename%.exe -devobj \Device\NTPNP_PCI0015 -policy 4 -cpu 1,2,3 -restart
@REM IF %ERRORLEVEL% EQU 0 %filename%.exe -devobj \Device\NTPNP_PCI0015 -policy 4 -cpu 1,2,3,4 -restart-on-change
@REM IF %ERRORLEVEL% EQU 0 %filename%.exe -devobj \Device\NTPNP_PCI0015 -msisupported 0

pause
goto loop
