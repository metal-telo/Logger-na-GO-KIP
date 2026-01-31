@echo off
chcp 65001 > nul
setlocal enabledelayedexpansion

REM ============================================================
REM Load tests with plow in 2 modes:
REM 1) with HTTP request logging (HTTP_LOGGING=true, LOG_LEVEL=info)
REM 2) without HTTP request logging (HTTP_LOGGING=false, LOG_LEVEL=error)
REM Results are saved under plow-results\with-logging and \no-logging
REM ============================================================

where plow >nul 2>nul
if errorlevel 1 (
  echo [ERROR] plow не найден в PATH. Установите https://github.com/six-ddc/plow и добавьте в PATH.
  pause
  exit /b 1
)

if not exist "server.exe" (
  echo [INFO] server.exe не найден. Соберите проект: go build -o server.exe ./cmd/server
  pause
  exit /b 1
)

set BASE_URL=http://localhost:8080
set API=%BASE_URL%/api

set RESULTS_DIR=plow-results
if not exist "%RESULTS_DIR%" mkdir "%RESULTS_DIR%"

call :run_mode with-logging true info
call :run_mode no-logging false error

echo.
echo Готово. Результаты в %RESULTS_DIR%\
pause
exit /b 0

REM =========================
REM mode_name, http_logging, log_level
REM =========================
:run_mode
set MODE=%1
set HTTPLOG=%2
set LOGLEVEL=%3

set OUTDIR=%RESULTS_DIR%\%MODE%
if not exist "%OUTDIR%" mkdir "%OUTDIR%"

echo.
echo ========================================
echo MODE: %MODE%
echo HTTP_LOGGING=%HTTPLOG%
echo LOG_LEVEL=%LOGLEVEL%
echo ========================================

REM Configure server via env
set PORT=:8080
set LOG_DIR=logs
set LOG_FILE=app.log
set METRICS_DIR=metrics
set METRICS_FILE=metrics.log
set HTTP_LOGGING=%HTTPLOG%
set LOG_LEVEL=%LOGLEVEL%
set TRACING_ENABLED=false
REM Allow duplicates only for load tests of POST /employees
set ALLOW_DUPLICATE_PASSPORTS=true

call :stop_server
call :start_server
call :wait_health || (
  echo [ERROR] Сервер не поднялся на %BASE_URL%
  call :stop_server
  exit /b 1
)

call :plow_suite "%OUTDIR%"

call :stop_server
exit /b 0

:start_server
start "" /B server.exe
exit /b 0

:stop_server
taskkill /IM server.exe /F >nul 2>nul
exit /b 0

:wait_health
REM wait up to ~15 seconds
set /a i=0
:wait_loop
set /a i+=1
curl -s "%API%/health" >nul 2>nul
if %errorlevel%==0 exit /b 0
if %i% GEQ 15 exit /b 1
timeout /t 1 /nobreak >nul
goto wait_loop

:plow_suite
set OUT=%~1
echo [INFO] Запуск тестов. Логи в %OUT%

REM Concurrency ladder
for %%C in (1 5 10) do (
  echo.
  echo --- Concurrency: %%C ---

  REM 1) Departments
  plow "%API%/departments" -c %%C -d 20s > "%OUT%\departments_c%%C.log" 2>&1

  REM 2) Employees by department
  plow "%API%/employees/department/dept1" -c %%C -d 20s > "%OUT%\employees_dept1_c%%C.log" 2>&1

  REM 3) Positions
  plow "%API%/positions" -c %%C -d 20s > "%OUT%\positions_c%%C.log" 2>&1

  REM 4) Search (POST)
  plow "%API%/employees/search" -c %%C -d 20s -T "application/json" --body "{\"full_name\":\"Иванов\"}" -m POST > "%OUT%\search_c%%C.log" 2>&1

  REM 5) Create employee (POST) - duplicates allowed in load test mode
  plow "%API%/employees" -c %%C -d 20s -T "application/json" --body "{\"full_name\":\"Load Test User\",\"gender\":\"male\",\"age\":30,\"education\":\"higher\",\"position\":\"Аналитик\",\"passport\":\"0000 000000\",\"department_id\":\"dept1\",\"status\":\"active\"}" -m POST > "%OUT%\create_c%%C.log" 2>&1

  REM 6) Update employee (PUT) - update emp1 repeatedly
  plow "%API%/employees/emp1" -c %%C -d 20s -T "application/json" --body "{\"full_name\":\"Иванов Иван Иванович\",\"gender\":\"male\",\"age\":35,\"education\":\"higher\",\"position\":\"Программист\",\"passport\":\"1234 567890\",\"department_id\":\"dept1\"}" -m PUT > "%OUT%\update_emp1_c%%C.log" 2>&1

  REM 7) Update status (PATCH) - toggle to vacation (idempotency not required for load test)
  plow "%API%/employees/emp1/status" -c %%C -d 20s -T "application/json" --body "{\"status\":\"vacation\"}" -m PATCH > "%OUT%\status_emp1_c%%C.log" 2>&1

  REM 8) Metrics endpoint (Prometheus)
  plow "%BASE_URL%/metrics" -c %%C -d 20s > "%OUT%\prom_metrics_c%%C.log" 2>&1
)

REM Simple report
echo # Load test report (%MODE%)> "%OUT%\report.md"
echo.>> "%OUT%\report.md"
echo Mode: **%MODE%**>> "%OUT%\report.md"
echo HTTP_LOGGING=%HTTPLOG%>> "%OUT%\report.md"
echo LOG_LEVEL=%LOGLEVEL%>> "%OUT%\report.md"
echo.>> "%OUT%\report.md"
echo Files:>> "%OUT%\report.md"
for %%F in ("%OUT%\*.log") do echo - %%~nxF>> "%OUT%\report.md"

exit /b 0




