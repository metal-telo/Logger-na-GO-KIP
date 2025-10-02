@echo off
chcp 65001 > nul
echo  ПРОВЕРКА ВСЕХ КОМПОНЕНТ СИСТЕМЫ
echo ================================
echo.

echo  1. ЛОГИРОВАНИЕ (slog + JSON):
echo    - Смотрим в терминале сервера
echo    - Формат: JSON с полями time, level, msg
echo.

echo  2. МЕТРИКИ (Prometheus):
echo    - Открываем: http://localhost:8080/metrics
echo    - Смотрим счетчики запросов и бизнес-метрики
echo.
echo  3. ТЕСТИРОВАНИЕ (Plow):
echo    - Результаты: папка plow-results/
echo    - Отчет: plow-results/report.md
echo.

echo  ДЛЯ ДЕМОНСТРАЦИИ:
echo    - Смотрим  логи в терминале
echo    - Смотрим   метрики по http://localhost:8080/metrics
echo    - Смотрим   отчеты plow
echo.
pause