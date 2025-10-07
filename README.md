REST API для управления сотрудниками на Go

## Точки доступа

| Метод | Путь | Описание |
|-------|------|-----------|
| `GET` | `/departments/{id}/employees` | Список сотрудников по ID департамента |
| `POST` | `/employees` | Создать сотрудника |
| `GET` | `/employees/search` | Поиск сотрудников с фильтрами |
| `PUT` | `/employees/{id}/dismiss` | Уволить сотрудника |
| `PUT` | `/employees/{id}` | Редактировать данные сотрудника |

## Модель сотрудника
```json
{
  "id": "uuid",
  "full_name": "string",
  "gender": "string",
  "age": "number",
  "education": "среднее|средне-специальное|высшее",
  "position": "string",
  "department_id": "uuid",
  "passport": "string",
  "status": "active|dismissed"
}
## Технические требования
Структура проекта: golang-standards

Логирование: slog + JSON в stdout

Трассировка: OpenTelemetry

Метрики: Prometheus

Тестирование: Нагрузочное тестирование с помощью plow

Хранилище: In-memory

