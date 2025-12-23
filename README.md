
# Разработка логгера на GO от КИПФИНА

## Изменение структуры проекта

Согласно требованиям, структура была изменена исходя из стандарта https://github.com/golang-standards/project-layout/blob/master/README_ru.md

Итоговая структура выглядит так:
```
employee-management/
├── cmd/
│   └── server/
│       ├── main.go              # Точка входа
│       └── static/
│           └── index.html       # Веб-интерфейс
|           └── style.css        # Стилизация веб-приложения
├── internal/
│   ├── handler/
│   │   └── handler.go           # HTTP обработчики
│   ├── logger/
│   │   └── logger.go            # Настройка логирования
│   ├── models/
│   │   └── models.go            # Модели данных
│   ├── repository/
│   │   ├── repository.go        # Интерфейс репозитория
│   │   └── memory.go            # In-memory реализация
│   ├── service/
│   │   └── employee.go          # Бизнес-логика
│   └── telemetry/
│       ├── metrics.go           # Prometheus метрики
│       └── tracing.go           # OpenTelemetry трассировка
├── logs/                        # Логи приложения 
│   ├── app.log
├── metrics/                     # Файлы метрик 
│   ├── metrics.log
├── go.mod                       # Зависимости Go
├── go.sum                       # Чексуммы зависимостей
└── README.md                    # Документация
```
