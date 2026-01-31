@echo off
chcp 65001 > nul
echo ========================================
echo Проверка работоспособности приложения
echo ========================================
echo.

echo 1. Health Check:
curl -s http://localhost:8080/api/health
echo.
echo.

echo 2. Получение департаментов:
curl -s http://localhost:8080/api/departments | findstr /C:"success" /C:"name"
echo.
echo.

echo 3. Получение сотрудников департамента dept1:
curl -s http://localhost:8080/api/employees/department/dept1 | findstr /C:"success" /C:"full_name"
echo.
echo.

echo 4. Поиск сотрудников:
curl -s -X POST http://localhost:8080/api/employees/search -H "Content-Type: application/json" -d "{\"full_name\":\"Иванов\"}" | findstr /C:"success" /C:"full_name"
echo.
echo.

echo 5. Проверка метрик:
curl -s http://localhost:8080/metrics | findstr /C:"http_requests_total" | findstr /N "." | findstr "^1:"
echo.
echo.

echo ========================================
echo Если все запросы вернули данные - приложение работает!
echo ========================================
pause





