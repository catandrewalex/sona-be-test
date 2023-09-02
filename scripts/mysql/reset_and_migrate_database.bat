@echo off
setlocal enabledelayedexpansion

rem Load environment variables from .env file (if it exists)
if exist .env (
  for /f "usebackq delims== tokens=1,2" %%i in (.env) do (
    if not "%%~$%%i" == "" (
      set "%%i=%%j"
    )
  )
)

rem Check that DB_NAME variable is set
if "%DB_NAME%" == "" (
  echo DB_NAME environment variable not set.
  exit /b 1
)

rem Set default values for other environment variables if they are not already set
if "%DB_USER%" == "" set "DB_USER=root"
if "%DB_PASSWORD%" == "" set "DB_PASSWORD=password"
if "%DB_HOST%" == "" set "DB_HOST=localhost"
if "%DB_PORT%" == "" set "DB_PORT=3306"

echo Recreating database %DB_NAME%...
mysql -u %DB_USER% -p%DB_PASSWORD% -h %DB_HOST% -P %DB_PORT% -e "DROP DATABASE IF EXISTS %DB_NAME%; CREATE DATABASE %DB_NAME%;"
echo Database has been recreated.
echo.

echo Running migrations and triggers...
for %%f in (.\data\sql\migrations\*.sql) do (
  echo Running migration %%f...
  mysql -u %DB_USER% -p%DB_PASSWORD% -h %DB_HOST% -P %DB_PORT% %DB_NAME% < "%%f"
)
for %%f in (.\data\sql\triggers\*.sql) do (
  echo Running trigger %%f...
  mysql -u %DB_USER% -p%DB_PASSWORD% -h %DB_HOST% -P %DB_PORT% %DB_NAME% < "%%f"
)
echo Migrations executed successfully.
echo.

:ask_input
REM Ask for user input
set /p user_input=Do you want to populate with development seed? (y/n): 

REM Convert user input to lowercase
set "user_input=%user_input:~0,1%"
set "user_input=%user_input:l=%"

REM Check user input
if "%user_input%"=="y" (
  echo Populating database with development seed...
  for %%f in (.\data\sql\dev\*.sql) do (
    echo Executing %%f...
    mysql -u %DB_USER% -p%DB_PASSWORD% -h %DB_HOST% -P %DB_PORT% %DB_NAME% < "%%f"
  )
  echo Database population executed successfully.
  echo.
) else if "%user_input%"=="n" (
  REM Do nothing
) else (
  echo Invalid input. Please enter either "y" or "n".
  goto ask_input
)

echo Done.
endlocal
