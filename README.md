
# Разработка логгера на GO от КИПФИНА

### Участники

- Архипова Полина @plkmjnhbgvfcdes
- Болотная Виктория @edi-hobi
- Гарас Кристина @KristinaGaras
- Конова Елизавета @metal-telo
- Кунаева Кира @coldszz

### Руководитель проекта

- Абзалимов Ришат Рафикович @abzalimovrrr

# Система управления сотрудниками с Go бэкендом и Vue.js фронтендом.

## Быстрый старт

### 1. Установка зависимостей

````bash
go mod tidy
````

### 2. Запуск приложения
````bash
go run main.go
````
Приложение будет доступно по адресу: http://localhost:8080

### 3. Проверка работы
Откройте в браузере:
-	http://localhost:8080 - веб-интерфейс
-	http://localhost:8080/api/health - health check
-	http://localhost:8080/metrics - Prometheus метрики

 ## API Endpoints

Основные endpoints:
-	GET /api/health - проверка здоровья
-	GET /api/departments - список департаментов
-	GET /api/employees/department/{id} - сотрудники департамента
-	POST /api/employees/search - поиск сотрудников
- POST /api/employees - создание сотрудника
- PUT /api/employees/{id} - обновление сотрудника
-	PATCH /api/employees/{id}/status - изменение статуса
-	GET /api/positions - список должностей
-	GET /api/metrics - бизнес-метрики
  
## Нагрузочное тестирование

1. Установите plow
````bash
go install github.com/six-ddc/plow@latest
````
2. Запустите тесты
````bash
chmod +x scripts/load-test.sh
./scripts/load-test.sh
````
3. Результаты
   
Результаты сохраняются в папку load-test-results/:
-	full_report.md - полный отчет
-	metrics_table.md - таблица метрик
-	*.log - логи отдельных тестов
  
## Конфигурация

Уровни логирования (переменная окружения LOG_LEVEL):
-	debug - подробное логирование (для разработки)
-	info - стандартное логирование
-	warn - только предупреждения и ошибки
-	error - только ошибки
````bash
LOG_LEVEL=warn go run main.go
````
## Технологии
-	Backend: Go, Gin, slog
-	Frontend: Vue.js 3, HTML5, CSS3
-	Метрики: Prometheus
-	Трассировка: OpenTelemetry + Jaeger
-	Тестирование: Plow
-	Хранилище: In-memory (для демо)
  
## Мониторинг

Prometheus метрики:
-	http_requests_total - количество HTTP запросов
-	http_request_duration_seconds - длительность запросов
-	employees_total - общее число сотрудников
-	employees_by_status - сотрудники по статусам

Трассировка:

Jaeger UI доступен по адресу: http://localhost:16686

## Особенности реализации
-	5 основных API endpoints
-	Slog с JSON выводом в stdout
-	OpenTelemetry трассировка
-	Prometheus метрики
-	Нагрузочное тестирование с Plow
-	In-memory хранилище
-	Стандартная структура проекта Go
  
Дополнительные возможности:
-	Веб-интерфейс на Vue.js
-	Адаптивный дизайн
-	Валидация данных
-	Graceful shutdown

## Инструкция по запуску:

### 1. Создать файлы:
```bash
mkdir employee-management-system
cd employee-management-system
```
Создать все файлы как в коде выше
- main.go
- go.mod
- static/index.html
- scripts/load-test.sh
  
### 2. Запустить проект:
```bash
go mod tidy
go run main.go
```
### 3. Открыть в браузере:

http://localhost:8080

### 4. Запустить тесты:
```bash
chmod +x scripts/load-test.sh
./scripts/load-test.sh
```
## В результате:
1.	Приложение с фронтендом и бэкендом
2.	5 endpoints из задания
3.	Slog + JSON логирование
4.	OpenTelemetry трассировка
5.	Prometheus метрики
6.	Нагрузочные тесты с сравнением логирования
7.	Отчеты в формате Markdown

