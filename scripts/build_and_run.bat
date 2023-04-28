@echo off
setlocal enabledelayedexpansion

set "EXE_NAME=sonamusica-backend.exe"
set "GO_FILES=."
set "GO_FLAGS=-v"

rem Build the executable
go build %GO_FLAGS% -o %EXE_NAME% %GO_FILES%

rem Terminate on any error
if %ERRORLEVEL% NEQ 0 (
  echo Build failed, there were compile errors.
  pause
) else (
  rem Run on no error
  echo Build successful, running %EXE_NAME%.
  %EXE_NAME%
)

endlocal