@echo off

SET filename=GoInterruptPolicy_debug

:loop
cls

gocritic check -enableAll -disable="#experimental,#opinionated,#commentedOutCode" ./...
go build -tags debug -buildvcs=false -o %filename%.exe

@REM IF %ERRORLEVEL% EQU 0 %filename%.exe -devobj \Device\NTPNP_PCI0015 -policy 4 -cpu 1,2,3 -restart
@REM IF %ERRORLEVEL% EQU 0 %filename%.exe -devobj \Device\NTPNP_PCI0015 -policy 4 -cpu 1,2,3,4 -restart-on-change
@REM IF %ERRORLEVEL% EQU 0 %filename%.exe -devobj \Device\NTPNP_PCI0015 -msisupported 0
IF %ERRORLEVEL% EQU 0 %filename%.exe

pause
goto loop
