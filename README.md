## ТЗ
```
 Модель сотрудника
1. получить список сотрудников по id департамента (guid)
2. создать сотрудника, привязав к департаменту:
- фио
- пол
- возраст
- тип образования (среднее, средне-специальное, высшее)
- должность (придумайте справочник)
- департамент
- паспортные данные
3. найти сотрудника (подумайте над фильтром)
4. уволить сотрудника
5. отредактировать сотрудника (к примеру, изменить должность, отправить в отпуск...)
 Технические требования
- Структура проекта: golang-standards
- Логирование: slog + JSON в stdout
- Трассировка: OpenTelemetry
- Метрики: Prometheus
- Тестирование: Нагрузочное тестирование с помощью plow
- Хранилище: In-memory
```
REST API для управления сотрудниками на Go
## Точки доступа
| Метод | Путь | Описание |
|-------|------|-----------|
| `GET` | `/departments/{id}/employees` | Список сотрудников по ID департамента |
| `POST` | `/employees` | Создать сотрудника |
| `GET` | `/employees/search` | Поиск сотрудников с фильтрами |
| `PUT` | `/employees/{id}/dismiss` | Уволить сотрудника |
| `PUT` | `/employees/{id}` | Редактировать данные сотрудника |

## Тесты

## Журналируется в файл app.log
{"time":"2025-10-08T02:20:15.8205877+03:00","level":"INFO","msg":"Логгер инициализирован","log_file":"logs/app.log"}

{"time":"2025-10-08T02:20:15.8460679+03:00","level":"INFO","msg":"Запись метрик в файл инициализирована","metrics_file":"metrics/metrics.log"}

{"time":"2025-10-08T02:20:15.8460679+03:00","level":"INFO","msg":"Трассировка отключена - Jaeger не запущен"}

{"time":"2025-10-08T02:20:15.8481023+03:00","level":"INFO","msg":"Запуск сервера","port":"8080"}

{"time":"2025-10-08T02:20:15.8481023+03:00","level":"ERROR","msg":"Ошибка запуска сервера","error":"listen tcp :8080: bind: Only one usage of each socket address (protocol/network address/port) is normally permitted."}

{"time":"2025-10-08T02:21:27.0920088+03:00","level":"INFO","msg":"Логгер инициализирован","log_file":"logs/app.log"}

{"time":"2025-10-08T02:21:27.1171321+03:00","level":"INFO","msg":"Запись метрик в файл инициализирована","metrics_file":"metrics/metrics.log"}

{"time":"2025-10-08T02:21:27.1172201+03:00","level":"INFO","msg":"Трассировка отключена - Jaeger не запущен"}

{"time":"2025-10-08T02:21:27.119142+03:00","level":"INFO","msg":"Запуск сервера","port":"8080"}

{"time":"2025-10-08T02:21:27.120132+03:00","level":"ERROR","msg":"Ошибка запуска сервера","error":"listen tcp :8080: bind: Only one usage of each socket address (protocol/network address/port) is normally permitted."}

{"time":"2025-10-08T02:22:14.3712293+03:00","level":"INFO","msg":"HTTP request","method":"GET","path":"/api/employees/department/dept3","status":200,"duration":"0s","client_ip":"::1"}

{"time":"2025-10-08T02:22:16.8292608+03:00","level":"INFO","msg":"HTTP request","method":"GET","path":"/api/employees/department/dept4","status":200,"duration":"0s","client_ip":"::1"}

{"time":"2025-10-08T02:22:19.4831373+03:00","level":"INFO","msg":"HTTP request","method":"GET","path":"/","status":200,"duration":"47.4µs","client_ip":"::1"}

{"time":"2025-10-08T02:22:19.561775+03:00","level":"INFO","msg":"HTTP request","method":"GET","path":"/api/departments","status":200,"duration":"0s","client_ip":"::1"}

{"time":"2025-10-08T02:22:19.5697221+03:00","level":"INFO","msg":"HTTP request","method":"GET","path":"/api/positions","status":200,"duration":"0s","client_ip":"::1"}

{"time":"2025-10-08T02:22:22.2892342+03:00","level":"INFO","msg":"HTTP request","method":"GET","path":"/api/employees/department/dept1","status":200,"duration":"0s","client_ip":"::1"}

{"time":"2025-10-08T02:22:27.6163644+03:00","level":"INFO","msg":"HTTP request","method":"PATCH","path":"/api/employees/emp1/status","status":200,"duration":"0s","client_ip":"::1"}

{"time":"2025-10-08T02:22:27.6361443+03:00","level":"INFO","msg":"HTTP request","method":"GET","path":"/api/employees/department/dept1","status":200,"duration":"0s","client_ip":"::1"}

{"time":"2025-10-08T02:22:31.9554001+03:00","level":"INFO","msg":"HTTP request","method":"PATCH","path":"/api/employees/emp1/status","status":200,"duration":"52.3µs","client_ip":"::1"}

{"time":"2025-10-08T02:22:31.9740588+03:00","level":"INFO","msg":"HTTP request","method":"GET","path":"/api/employees/department/dept1","status":200,"duration":"0s","client_ip":"::1"}

{"time":"2025-10-08T02:22:35.9147469+03:00","level":"INFO","msg":"HTTP request","method":"PATCH","path":"/api/employees/emp2/status","status":200,"duration":"0s","client_ip":"::1"}

{"time":"2025-10-08T02:22:35.932661+03:00","level":"INFO","msg":"HTTP request","method":"GET","path":"/api/employees/department/dept1","status":200,"duration":"0s","client_ip":"::1"}

{"time":"2025-10-08T02:22:51.3980207+03:00","level":"INFO","msg":"HTTP request","method":"PUT","path":"/api/employees/emp1","status":200,"duration":"50.4µs","client_ip":"::1"}

{"time":"2025-10-08T02:22:51.4025942+03:00","level":"INFO","msg":"HTTP request","method":"GET","path":"/api/employees/department/dept1","status":200,"duration":"0s","client_ip":"::1"}

## Нагрузочное тестирование

PS C:\Go\Project> plow http://localhost:8080/api/departments -c 1 -d 30s

Benchmarking http://localhost:8080/api/departments for 30s using 1 connection(s).

@ Real-time charts is listening on http://[::]:18888


### Summary: 
| Параметр | Значение |
|-------|------|
  | Elapsed | 30s |
  | Count | 147321 | 
  | 2xx | 147321 |
  | RPS | 4910.691 |
  | Reads | 4.177MB/s |
  | Writes | 0.342MB/s |

| Statistics | Min | Mean | StdDev | Max |
|-------|------|-------|------|-------|
| Latency | 0s | 196µs | 604µs | 53.074ms |
|  RPS  | 4098.58 | 4918.69 | 332.29 | 5539.25 |

### Latency Percentile: 

| P50 | P75 | P90 | P95 | P99 | P99.9 | P99.99 |
|-------|-------|-------|-------|-------|-------|-------|
 | 0s |  512µs|527µs | 999µs| 1.001ms| 1.475ms | 32.514ms

### Latency Histogram:
| Параметр | Значение | % |
|-------|-------|-------|
  |2µs    |   90329 | 61.31%|
 | 436µs  |   49938 | 33.90%|
 | 816µs   |   5518 |  3.75%|
 | 1.162ms  |  1504  | 1.02%|
 | 12.889ms  |  21 |  0.01%|
 | 31.003ms  |    8 |  0.01%|
 | 46.103ms   |   2  | 0.00%|
 | 53.074ms   |   1  | 0.00%|

PS C:\Go\Project> plow http://localhost:8080/api/departments -c 5 -d 30s  
Benchmarking http://localhost:8080/api/departments for 30s using 5 connection(s).
@ Real-time charts is listening on http://[::]:18888
### Summary: 
| Параметр | Значение |
|-------|------|
 | Elapsed   |     30s|
 | Count     |  187841|
  |  2xx      | 187841|
 | RPS     |  6261.291|
 | Reads  |  5.326MB/s|
 | Writes |  0.436MB/s|

|Statistics |   Min  |   Mean  |  StdDev |    Max|
|-------|-------|-------|-------|-------|
 | Latency   |  0s |     792µs |  1.467ms | 67.337ms|
 | RPS     |  5077.94 | 6260.76 |  605.87  |  7599.14| 

### Latency Percentile: 

  |P50   |   P75   |  P90    |  P95  |    P99  |   P99.9  |  P99.99|
  |-------|-------|-------|-------|-------|-------|-------|
  | 626µs | 1.002ms | 1.12ms | 1.217ms | 1.974ms | 32.086ms | 56.29ms|

### Latency Histogram:
| Параметр | Значение | % |
  | 702µs  | 165061 | 87.87% |
  | 1.101ms  |  18782 | 10.00% |
  | 1.618ms  |   3511  | 1.87%|
  |7.777ms  | 423 |  0.23%|
  |45.51ms   |    49 |  0.03%|
  |54.103ms   |    9 |  0.00%|
  |64.172ms    |   5 |  0.00%|
  |66.799ms    |   1 |  0.00%|

PS C:\Go\Project> plow http://localhost:8080/api/departments -c 10 -d 1m
Benchmarking http://localhost:8080/api/departments for 1m0s using 10 connection(s).
@ Real-time charts is listening on http://[::]:18888

### Summary:
| Параметр | Значение |
|-------|------|
 | Elapsed    |   1m0s|
 | Count   |    322537|
  |  2xx   |    322537|
 | RPS  |     5375.612|
|  Reads |   4.573MB/s|
 | Writes  | 0.374MB/s|

|Statistics |   Min  |   Mean  |  StdDev   |  Max|
|-------|-------|-------|-------|-------|
 | Latency  |   0s   |  1.852ms | 3.146ms | 374.13ms|
 | RPS     |  2634.04 | 5383.03 | 971.37  | 7341.19|

### Latency Percentile:
  |P50   |   P75 |   P90  |    P95   |   P99  |  P99.9   |  P99.99|
  |-------|-------|-------|-------|-------|-------|-------|
 | 1.715ms | 2ms | 2.359ms | 2.806ms | 4.405ms | 47.474ms | 74.139ms|

### Latency Histogram:
| Параметр | Значение | % |
|-------|-------|-------|
 |1.751ms  |  302550 | 93.80%|
  |2.699ms  |   18752 |  5.81%|
  |5.918ms    |  1042  | 0.32%|
  |24.398ms   |   120 |  0.04%|
 | 55.314ms    |   36 |  0.01%|
 | 70.759ms      | 23  | 0.01%|
 | 85.423ms     |   4 |  0.00%|
  |373.384ms    |  10  | 0.00%|

http://localhost:8080/metrics
# METRICS
## HELP employees_total Total number of employees
## TYPE employees_total gauge
employees_total 0
## HELP go_gc_duration_seconds A summary of the pause duration of garbage collection cycles.
## TYPE go_gc_duration_seconds summary
go_gc_duration_seconds{quantile="0"} 0
go_gc_duration_seconds{quantile="0.25"} 0
go_gc_duration_seconds{quantile="0.5"} 0
go_gc_duration_seconds{quantile="0.75"} 5.06e-05
go_gc_duration_seconds{quantile="1"} 0.002001
go_gc_duration_seconds_sum 0.1816512
go_gc_duration_seconds_count 2416
## HELP go_goroutines Number of goroutines that currently exist.
## TYPE go_goroutines gauge
go_goroutines 10
## HELP go_info Information about the Go environment.
## TYPE go_info gauge
go_info{version="go1.25.0"} 1
## HELP go_memstats_alloc_bytes Number of bytes allocated and still in use.
## TYPE go_memstats_alloc_bytes gauge
go_memstats_alloc_bytes 3.140808e+06
## HELP go_memstats_alloc_bytes_total Total number of bytes allocated, even if freed.
## TYPE go_memstats_alloc_bytes_total counter
go_memstats_alloc_bytes_total 3.938079416e+09
## HELP go_memstats_buck_hash_sys_bytes Number of bytes used by the profiling bucket hash table.
## TYPE go_memstats_buck_hash_sys_bytes gauge
go_memstats_buck_hash_sys_bytes 7674
## HELP go_memstats_frees_total Total number of frees.
## TYPE go_memstats_frees_total counter
go_memstats_frees_total 3.5817594e+07
## HELP go_memstats_gc_sys_bytes Number of bytes used for garbage collection system metadata.
## TYPE go_memstats_gc_sys_bytes gauge
go_memstats_gc_sys_bytes 3.21336e+06
## HELP go_memstats_heap_alloc_bytes Number of heap bytes allocated and still in use.
## TYPE go_memstats_heap_alloc_bytes gauge
go_memstats_heap_alloc_bytes 3.140808e+06
## HELP go_memstats_heap_idle_bytes Number of heap bytes waiting to be used.
## TYPE go_memstats_heap_idle_bytes gauge
go_memstats_heap_idle_bytes 5.111808e+06
## HELP go_memstats_heap_inuse_bytes Number of heap bytes that are in use.
## TYPE go_memstats_heap_inuse_bytes gauge
go_memstats_heap_inuse_bytes 5.636096e+06
## HELP go_memstats_heap_objects Number of allocated objects.
## TYPE go_memstats_heap_objects gauge
go_memstats_heap_objects 11151
## HELP go_memstats_heap_released_bytes Number of heap bytes released to OS.
## TYPE go_memstats_heap_released_bytes gauge
go_memstats_heap_released_bytes 4.898816e+06
## HELP go_memstats_heap_sys_bytes Number of heap bytes obtained from system.
## TYPE go_memstats_heap_sys_bytes gauge
go_memstats_heap_sys_bytes 1.0747904e+07
## HELP go_memstats_last_gc_time_seconds Number of seconds since 1970 of last garbage collection.
## TYPE go_memstats_last_gc_time_seconds gauge
go_memstats_last_gc_time_seconds 1.759878554228138e+09
## HELP go_memstats_lookups_total Total number of pointer lookups.
## TYPE go_memstats_lookups_total counter
go_memstats_lookups_total 0
## HELP go_memstats_mallocs_total Total number of mallocs.
## TYPE go_memstats_mallocs_total counter
go_memstats_mallocs_total 3.5828745e+07
## HELP go_memstats_mcache_inuse_bytes Number of bytes in use by mcache structures.
## TYPE go_memstats_mcache_inuse_bytes gauge
go_memstats_mcache_inuse_bytes 9408
## HELP go_memstats_mcache_sys_bytes Number of bytes used for mcache structures obtained from system.
## TYPE go_memstats_mcache_sys_bytes gauge
go_memstats_mcache_sys_bytes 15288
## HELP go_memstats_mspan_inuse_bytes Number of bytes in use by mspan structures.
## TYPE go_memstats_mspan_inuse_bytes gauge
go_memstats_mspan_inuse_bytes 168320
## HELP go_memstats_mspan_sys_bytes Number of bytes used for mspan structures obtained from system.
## TYPE go_memstats_mspan_sys_bytes gauge
go_memstats_mspan_sys_bytes 212160
## HELP go_memstats_next_gc_bytes Number of heap bytes when next garbage collection will take place.
## TYPE go_memstats_next_gc_bytes gauge
go_memstats_next_gc_bytes 4.807778e+06
## HELP go_memstats_other_sys_bytes Number of bytes used for other system allocations.
## TYPE go_memstats_other_sys_bytes gauge
go_memstats_other_sys_bytes 1.978718e+06
## HELP go_memstats_stack_inuse_bytes Number of bytes in use by the stack allocator.
## TYPE go_memstats_stack_inuse_bytes gauge
go_memstats_stack_inuse_bytes 1.80224e+06
## HELP go_memstats_stack_sys_bytes Number of bytes obtained from system for stack allocator.
## TYPE go_memstats_stack_sys_bytes gauge
go_memstats_stack_sys_bytes 1.80224e+06
## HELP go_memstats_sys_bytes Number of bytes obtained from system.
## TYPE go_memstats_sys_bytes gauge
go_memstats_sys_bytes 1.7977344e+07
## HELP go_threads Number of OS threads created.
## TYPE go_threads gauge
go_threads 15
## HELP http_request_duration_seconds HTTP request duration in seconds
## TYPE http_request_duration_seconds histogram
http_request_duration_seconds_bucket{method="GET",path="/",le="0.005"} 1
http_request_duration_seconds_bucket{method="GET",path="/",le="0.01"} 1
http_request_duration_seconds_bucket{method="GET",path="/",le="0.025"} 1
http_request_duration_seconds_bucket{method="GET",path="/",le="0.05"} 1
http_request_duration_seconds_bucket{method="GET",path="/",le="0.1"} 1
http_request_duration_seconds_bucket{method="GET",path="/",le="0.25"} 1
http_request_duration_seconds_bucket{method="GET",path="/",le="0.5"} 1
http_request_duration_seconds_bucket{method="GET",path="/",le="1"} 1
http_request_duration_seconds_bucket{method="GET",path="/",le="2.5"} 1
http_request_duration_seconds_bucket{method="GET",path="/",le="5"} 1
http_request_duration_seconds_bucket{method="GET",path="/",le="10"} 1
http_request_duration_seconds_bucket{method="GET",path="/",le="+Inf"} 1
http_request_duration_seconds_sum{method="GET",path="/"} 0
http_request_duration_seconds_count{method="GET",path="/"} 1
http_request_duration_seconds_bucket{method="GET",path="/api/departments",le="0.005"} 657694
http_request_duration_seconds_bucket{method="GET",path="/api/departments",le="0.01"} 657695
http_request_duration_seconds_bucket{method="GET",path="/api/departments",le="0.025"} 657704
http_request_duration_seconds_bucket{method="GET",path="/api/departments",le="0.05"} 657716
http_request_duration_seconds_bucket{method="GET",path="/api/departments",le="0.1"} 657716
http_request_duration_seconds_bucket{method="GET",path="/api/departments",le="0.25"} 657716
http_request_duration_seconds_bucket{method="GET",path="/api/departments",le="0.5"} 657716
http_request_duration_seconds_bucket{method="GET",path="/api/departments",le="1"} 657716
http_request_duration_seconds_bucket{method="GET",path="/api/departments",le="2.5"} 657716
http_request_duration_seconds_bucket{method="GET",path="/api/departments",le="5"} 657716
http_request_duration_seconds_bucket{method="GET",path="/api/departments",le="10"} 657716
http_request_duration_seconds_bucket{method="GET",path="/api/departments",le="+Inf"} 657716
http_request_duration_seconds_sum{method="GET",path="/api/departments"} 14.225759599999966
http_request_duration_seconds_count{method="GET",path="/api/departments"} 657716
http_request_duration_seconds_bucket{method="GET",path="/api/positions",le="0.005"} 1
http_request_duration_seconds_bucket{method="GET",path="/api/positions",le="0.01"} 1
http_request_duration_seconds_bucket{method="GET",path="/api/positions",le="0.025"} 1
http_request_duration_seconds_bucket{method="GET",path="/api/positions",le="0.05"} 1
http_request_duration_seconds_bucket{method="GET",path="/api/positions",le="0.1"} 1
http_request_duration_seconds_bucket{method="GET",path="/api/positions",le="0.25"} 1
http_request_duration_seconds_bucket{method="GET",path="/api/positions",le="0.5"} 1
http_request_duration_seconds_bucket{method="GET",path="/api/positions",le="1"} 1
http_request_duration_seconds_bucket{method="GET",path="/api/positions",le="2.5"} 1
http_request_duration_seconds_bucket{method="GET",path="/api/positions",le="5"} 1
http_request_duration_seconds_bucket{method="GET",path="/api/positions",le="10"} 1
http_request_duration_seconds_bucket{method="GET",path="/api/positions",le="+Inf"} 1
http_request_duration_seconds_sum{method="GET",path="/api/positions"} 0
http_request_duration_seconds_count{method="GET",path="/api/positions"} 1
http_request_duration_seconds_bucket{method="GET",path="/favicon.ico",le="0.005"} 1
http_request_duration_seconds_bucket{method="GET",path="/favicon.ico",le="0.01"} 1
http_request_duration_seconds_bucket{method="GET",path="/favicon.ico",le="0.025"} 1
http_request_duration_seconds_bucket{method="GET",path="/favicon.ico",le="0.05"} 1
http_request_duration_seconds_bucket{method="GET",path="/favicon.ico",le="0.1"} 1
http_request_duration_seconds_bucket{method="GET",path="/favicon.ico",le="0.25"} 1
http_request_duration_seconds_bucket{method="GET",path="/favicon.ico",le="0.5"} 1
http_request_duration_seconds_bucket{method="GET",path="/favicon.ico",le="1"} 1
http_request_duration_seconds_bucket{method="GET",path="/favicon.ico",le="2.5"} 1
http_request_duration_seconds_bucket{method="GET",path="/favicon.ico",le="5"} 1
http_request_duration_seconds_bucket{method="GET",path="/favicon.ico",le="10"} 1
http_request_duration_seconds_bucket{method="GET",path="/favicon.ico",le="+Inf"} 1
http_request_duration_seconds_sum{method="GET",path="/favicon.ico"} 0
http_request_duration_seconds_count{method="GET",path="/favicon.ico"} 1
http_request_duration_seconds_bucket{method="GET",path="/metrics",le="0.005"} 2
http_request_duration_seconds_bucket{method="GET",path="/metrics",le="0.01"} 2
http_request_duration_seconds_bucket{method="GET",path="/metrics",le="0.025"} 2
http_request_duration_seconds_bucket{method="GET",path="/metrics",le="0.05"} 2
http_request_duration_seconds_bucket{method="GET",path="/metrics",le="0.1"} 2
http_request_duration_seconds_bucket{method="GET",path="/metrics",le="0.25"} 2
http_request_duration_seconds_bucket{method="GET",path="/metrics",le="0.5"} 2
http_request_duration_seconds_bucket{method="GET",path="/metrics",le="1"} 2
http_request_duration_seconds_bucket{method="GET",path="/metrics",le="2.5"} 2
http_request_duration_seconds_bucket{method="GET",path="/metrics",le="5"} 2
http_request_duration_seconds_bucket{method="GET",path="/metrics",le="10"} 2
http_request_duration_seconds_bucket{method="GET",path="/metrics",le="+Inf"} 2
http_request_duration_seconds_sum{method="GET",path="/metrics"} 0.0020251
http_request_duration_seconds_count{method="GET",path="/metrics"} 2
## HELP http_requests_total Total number of HTTP requests
## TYPE http_requests_total counter
http_requests_total{method="GET",path="/",status="200"} 1
http_requests_total{method="GET",path="/api/departments",status="200"} 657716
http_requests_total{method="GET",path="/api/positions",status="200"} 1
http_requests_total{method="GET",path="/favicon.ico",status="404"} 1
http_requests_total{method="GET",path="/metrics",status="200"} 2
## HELP process_cpu_seconds_total Total user and system CPU time spent in seconds.
## TYPE process_cpu_seconds_total counter
process_cpu_seconds_total 145.390625
## HELP process_max_fds Maximum number of open file descriptors.
## TYPE process_max_fds gauge
process_max_fds 1.6777216e+07
## HELP process_open_fds Number of open file descriptors.
## TYPE process_open_fds gauge
process_open_fds 192
## HELP process_resident_memory_bytes Resident memory size in bytes.
## TYPE process_resident_memory_bytes gauge
process_resident_memory_bytes 2.1700608e+07
## HELP process_start_time_seconds Start time of the process since unix epoch in seconds.
## TYPE process_start_time_seconds gauge
process_start_time_seconds 1.759877833e+09
## HELP process_virtual_memory_bytes Virtual memory size in bytes.
## TYPE process_virtual_memory_bytes gauge
process_virtual_memory_bytes 2.3957504e+07
## HELP promhttp_metric_handler_requests_in_flight Current number of scrapes being served.
## TYPE promhttp_metric_handler_requests_in_flight gauge
promhttp_metric_handler_requests_in_flight 1
## HELP promhttp_metric_handler_requests_total Total number of scrapes by HTTP status code.
# TYPE promhttp_metric_handler_requests_total counter
promhttp_metric_handler_requests_total{code="200"} 2
promhttp_metric_handler_requests_total{code="500"} 0
promhttp_metric_handler_requests_total{code="503"} 0
