@echo off

SET filename=main

:loop
cls

@REM gocritic check -enable="#performance" ./...
gocritic check -enableAll -disable="#experimental,#opinionated,#commentedOutCode" ./...

@REM go build -tags debug
@REM go build -race -o %filename%.exe

go build -tags debug -o %filename%.exe

IF %ERRORLEVEL% EQU 0 %filename%.exe

pause
goto loop
