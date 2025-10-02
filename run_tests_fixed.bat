@echo off
chcp 65001 > nul
echo 🚀 Запуск нагрузочного тестирования...
echo.

set RESULTS_DIR=plow-results
if not exist "%RESULTS_DIR%" mkdir "%RESULTS_DIR%"

echo 📊 Тест 1: Сотрудники по департаменту
plow http://localhost:8080/api/employees/department/dept1 -c 25 -n 2000 > "%RESULTS_DIR%\test1_employees.log" 2>&1
echo ✅ Результаты сохранены в %RESULTS_DIR%\test1_employees.log

echo.
echo 🔍 Тест 2: Поиск сотрудников  
plow http://localhost:8080/api/employees/search -c 20 -n 1500 -T "application/json" --body "{\"full_name\":\"Иванов\"}" -m POST > "%RESULTS_DIR%\test2_search.log" 2>&1
echo ✅ Результаты сохранены в %RESULTS_DIR%\test2_search.log

echo.
echo 🏢 Тест 3: Департаменты
plow http://localhost:8080/api/departments -c 30 -n 3000 > "%RESULTS_DIR%\test3_departments.log" 2>&1
echo ✅ Результаты сохранены в %RESULTS_DIR%\test3_departments.log

echo.
echo 📈 Анализ результатов...
echo # Отчет нагрузочного тестирования > "%RESULTS_DIR%\report.md"
echo. >> "%RESULTS_DIR%\report.md"
echo "## Результаты тестов" >> "%RESULTS_DIR%\report.md"
echo. >> "%RESULTS_DIR%\report.md"

if exist "%RESULTS_DIR%\test1_employees.log" (
    echo "### 1. Сотрудники по департаменту" >> "%RESULTS_DIR%\report.md"
    echo "\`\`\`text" >> "%RESULTS_DIR%\report.md"
    type "%RESULTS_DIR%\test1_employees.log" >> "%RESULTS_DIR%\report.md"
    echo "\`\`\`" >> "%RESULTS_DIR%\report.md"
    echo. >> "%RESULTS_DIR%\report.md"
)

if exist "%RESULTS_DIR%\test2_search.log" (
    echo "### 2. Поиск сотрудников" >> "%RESULTS_DIR%\report.md"
    echo "\`\`\`text" >> "%RESULTS_DIR%\report.md"
    type "%RESULTS_DIR%\test2_search.log" >> "%RESULTS_DIR%\report.md"
    echo "\`\`\`" >> "%RESULTS_DIR%\report.md"
    echo. >> "%RESULTS_DIR%\report.md"
)

if exist "%RESULTS_DIR%\test3_departments.log" (
    echo "### 3. Департаменты" >> "%RESULTS_DIR%\report.md"
    echo "\`\`\`text" >> "%RESULTS_DIR%\report.md"
    type "%RESULTS_DIR%\test3_departments.log" >> "%RESULTS_DIR%\report.md"
    echo "\`\`\`" >> "%RESULTS_DIR%\report.md"
)

echo.
echo 🎉 Тестирование завершено!
echo 📄 Отчет: %RESULTS_DIR%\report.md
echo 📊 Логи: %RESULTS_DIR%\
echo.
echo 📋 Для просмотра результатов выполни:
echo    type "%RESULTS_DIR%\report.md"
echo.
pause