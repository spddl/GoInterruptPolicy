@echo off

SET GOOS=windows
SET GOARCH=amd64
SET filename=GoInterruptPolicy

:loop
CLS

gocritic check -enableAll -disable="#experimental,#opinionated,#commentedOutCode" ./...

IF exist %filename%.exe (
    FOR /F "usebackq" %%A IN ('%filename%.exe') DO SET /A beforeSize=%%~zA
) ELSE (
    SET /A beforeSize=0
)

: Build https://golang.org/cmd/go/
go build -tags debug -buildvcs=false -o %filename%_debug.exe
go build -ldflags="-w -s -H windowsgui" -o %filename%.exe

FOR /F "usebackq" %%A IN ('%filename%.exe') DO SET /A size=%%~zA
SET /A diffSize = %size% - %beforeSize%
SET /A size=(%size%/1024)+1
IF %diffSize% EQU 0 (
    ECHO %size% kb
) ELSE (
    IF %diffSize% GTR 0 (
        ECHO %size% kb [+%diffSize% b]
    ) ELSE (
        ECHO %size% kb [%diffSize% b]
    )
)

: Run
@REM IF %ERRORLEVEL% EQU 0 %filename%.exe

PAUSE
GOTO loop
