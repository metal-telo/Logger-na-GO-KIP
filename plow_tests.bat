@echo off
echo 🚀 Запуск нагрузочного тестирования...
echo.

set RESULTS_DIR=plow-results
if not exist "%RESULTS_DIR%" mkdir "%RESULTS_DIR%"

echo 📊 Тест 1: Сотрудники по департаменту
plow http://localhost:8080/api/employees/department/dept1 -c 25 -n 2000 > "%RESULTS_DIR%/test1_employees.log"
echo ✅ Результаты сохранены в %RESULTS_DIR%/test1_employees.log

echo.
echo 🔍 Тест 2: Поиск сотрудников
plow http://localhost:8080/api/employees/search -c 20 -n 1500 -T "application/json" --body "{\"full_name\":\"Иванов\"}" -m POST > "%RESULTS_DIR%/test2_search.log"
echo ✅ Результаты сохранены в %RESULTS_DIR%/test2_search.log

echo.
echo 🏢 Тест 3: Департаменты
plow http://localhost:8080/api/departments -c 30 -n 3000 > "%RESULTS_DIR%/test3_departments.log"
echo ✅ Результаты сохранены в %RESULTS_DIR%/test3_departments.log

echo.
echo 📈 Создание отчета...

echo # Отчет нагрузочного тестирования > "%RESULTS_DIR%/report.md"
echo. >> "%RESULTS_DIR%/report.md"
echo "## Результаты тестов" >> "%RESULTS_DIR%/report.md"
echo. >> "%RESULTS_DIR%/report.md"

for %%f in ("%RESULTS_DIR%/*.log") do (
    echo "### %%~nf" >> "%RESULTS_DIR%/report.md"
    echo "\`\`\`" >> "%RESULTS_DIR%/report.md"
    type "%%f" >> "%RESULTS_DIR%/report.md"
    echo "\`\`\`" >> "%RESULTS_DIR%/report.md"
    echo. >> "%RESULTS_DIR%/report.md"
)

echo.
echo 🎉 Тестирование завершено!
echo 📄 Отчет: %RESULTS_DIR%\report.md
echo 📊 Логи: %RESULTS_DIR%\
pause