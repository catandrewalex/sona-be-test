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

echo Dropping database %DB_NAME%...
mysql -u %DB_USER% -p%DB_PASSWORD% -h %DB_HOST% -P %DB_PORT% -e "DROP DATABASE IF EXISTS %DB_NAME%; CREATE DATABASE %DB_NAME%;"

echo Running migrations...
for %%f in (.\data\sql\migrations\*.sql) do (
  echo Running migration %%f...
  mysql -u %DB_USER% -p%DB_PASSWORD% -h %DB_HOST% -P %DB_PORT% %DB_NAME% < "%%f"
)

echo Done.
endlocal
