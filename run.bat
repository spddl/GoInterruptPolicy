@echo off

SET filename=GoInterruptPolicy_debug

:loop
cls

gocritic check -enableAll -disable="#experimental,#opinionated,#commentedOutCode" ./...
go build -tags debug -o %filename%.exe

IF %ERRORLEVEL% EQU 0 %filename%.exe

pause
goto loop
