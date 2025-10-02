package main

import (
	"context"
	"embed"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

//go:embed static/*
var staticFiles embed.FS

// МОДЕЛИ ДАННЫХ

type Department struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

type Employee struct {
	ID           string     `json:"id"`
	FullName     string     `json:"full_name"`
	Gender       string     `json:"gender"`
	Age          int        `json:"age"`
	Education    string     `json:"education"`
	Position     string     `json:"position"`
	Passport     string     `json:"passport"`
	DepartmentID string     `json:"department_id"`
	Status       string     `json:"status"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	FiredAt      *time.Time `json:"fired_at,omitempty"`
}

type EmployeeSearchRequest struct {
    FullName  string `json:"full_name"`
    Position  string `json:"position"`
    Gender    string `json:"gender"`
    Education string `json:"education"`
    AgeFrom   *int   `json:"age_from,omitempty"`
    AgeTo     *int   `json:"age_to,omitempty"`
}

type StatusUpdateRequest struct {
	Status string `json:"status"`
}

type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
}

// IN-MEMORY РЕПОЗИТОРИЙ


type Repository interface {
	GetDepartments(ctx context.Context) ([]Department, error)
	GetEmployeesByDepartment(ctx context.Context, departmentID string) ([]Employee, error)
	SearchEmployees(ctx context.Context, req EmployeeSearchRequest) ([]Employee, error)
	CreateEmployee(ctx context.Context, emp Employee) (*Employee, error)
	UpdateEmployee(ctx context.Context, emp Employee) (*Employee, error)
	UpdateEmployeeStatus(ctx context.Context, id string, status string) (*Employee, error)
	GetPositions(ctx context.Context) ([]string, error)
	GetEmployeeStats(ctx context.Context) (map[string]interface{}, error)
}

type MemoryRepository struct {
	mu          sync.RWMutex
	departments map[string]Department
	employees   map[string]Employee
	positions   []string
}

func NewMemoryRepository() *MemoryRepository {
	repo := &MemoryRepository{
		departments: make(map[string]Department),
		employees:   make(map[string]Employee),
		positions: []string{
			"Программист", "Аналитик", "Тестировщик", "Менеджер по продажам",
			"HR-менеджер", "Бухгалтер", "Маркетолог", "Дизайнер",
			"Системный администратор", "Руководитель отдела",
		},
	}

	// Инициализация тестовых данных
	repo.initTestData()
	return repo
}

func (r *MemoryRepository) initTestData() {
	now := time.Now()
	
	// Департаменты
	depts := []Department{
		{ID: "dept1", Name: "IT-департамент", Description: "Разработка ПО", CreatedAt: now},
		{ID: "dept2", Name: "Отдел продаж", Description: "Продажи и маркетинг", CreatedAt: now},
		{ID: "dept3", Name: "HR-отдел", Description: "Управление персоналом", CreatedAt: now},
		{ID: "dept4", Name: "Финансовый отдел", Description: "Финансы и бухгалтерия", CreatedAt: now},
		{ID: "dept5", Name: "Маркетинг", Description: "Маркетинг и реклама", CreatedAt: now},
	}

	for _, dept := range depts {
		r.departments[dept.ID] = dept
	}

	// Сотрудники
	employees := []Employee{
		{
			ID: "emp1", FullName: "Иванов Иван Иванович", Gender: "male", Age: 35,
			Education: "higher", Position: "Программист", Passport: "1234 567890",
			DepartmentID: "dept1", Status: "active", CreatedAt: now, UpdatedAt: now,
		},
		{
			ID: "emp2", FullName: "Петрова Анна Сергеевна", Gender: "female", Age: 28,
			Education: "higher", Position: "Аналитик", Passport: "2345 678901",
			DepartmentID: "dept1", Status: "vacation", CreatedAt: now, UpdatedAt: now,
		},
		{
			ID: "emp3", FullName: "Сидоров Петр Александрович", Gender: "male", Age: 42,
			Education: "higher", Position: "Менеджер по продажам", Passport: "3456 789012",
			DepartmentID: "dept2", Status: "active", CreatedAt: now, UpdatedAt: now,
		},
		{
			ID: "emp4", FullName: "Козлова Мария Викторовна", Gender: "female", Age: 31,
			Education: "higher", Position: "HR-менеджер", Passport: "4567 890123",
			DepartmentID: "dept3", Status: "active", CreatedAt: now, UpdatedAt: now,
		},
	}

	for _, emp := range employees {
		r.employees[emp.ID] = emp
	}
}

func (r *MemoryRepository) GetDepartments(ctx context.Context) ([]Department, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var departments []Department
	for _, dept := range r.departments {
		departments = append(departments, dept)
	}
	return departments, nil
}

func (r *MemoryRepository) GetEmployeesByDepartment(ctx context.Context, departmentID string) ([]Employee, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var employees []Employee
	for _, emp := range r.employees {
		if emp.DepartmentID == departmentID {
			employees = append(employees, emp)
		}
	}
	return employees, nil
}

func (r *MemoryRepository) SearchEmployees(ctx context.Context, req EmployeeSearchRequest) ([]Employee, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()

    var employees []Employee
    for _, emp := range r.employees {
        if req.FullName != "" && !contains(emp.FullName, req.FullName) {
            continue
        }
        if req.Position != "" && emp.Position != req.Position {
            continue
        }
        if req.Gender != "" && emp.Gender != req.Gender {
            continue
        }
        if req.Education != "" && emp.Education != req.Education {
            continue
        }
        if req.AgeFrom != nil && emp.Age < *req.AgeFrom {
            continue
        }
        if req.AgeTo != nil && emp.Age > *req.AgeTo {
            continue
        }

        employees = append(employees, emp)
    }
    return employees, nil
}

func (r *MemoryRepository) CreateEmployee(ctx context.Context, emp Employee) (*Employee, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Проверка уникальности паспорта
	for _, existing := range r.employees {
		if existing.Passport == emp.Passport {
			return nil, fmt.Errorf("сотрудник с таким паспортом уже существует")
		}
	}

	// Генерация ID и установка временных меток
	emp.ID = fmt.Sprintf("emp%d", len(r.employees)+1)
	now := time.Now()
	emp.CreatedAt = now
	emp.UpdatedAt = now
	if emp.Status == "" {
		emp.Status = "active"
	}

	r.employees[emp.ID] = emp
	return &emp, nil
}

func (r *MemoryRepository) UpdateEmployee(ctx context.Context, emp Employee) (*Employee, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, exists := r.employees[emp.ID]
	if !exists {
		return nil, fmt.Errorf("сотрудник не найден")
	}

	// Проверка уникальности паспорта (исключая текущего сотрудника)
	for _, e := range r.employees {
		if e.ID != emp.ID && e.Passport == emp.Passport {
			return nil, fmt.Errorf("сотрудник с таким паспортом уже существует")
		}
	}

	emp.CreatedAt = existing.CreatedAt
	emp.UpdatedAt = time.Now()
	emp.Status = existing.Status // Сохраняем статус

	r.employees[emp.ID] = emp
	return &emp, nil
}

func (r *MemoryRepository) UpdateEmployeeStatus(ctx context.Context, id string, status string) (*Employee, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	emp, exists := r.employees[id]
	if !exists {
		return nil, fmt.Errorf("сотрудник не найден")
	}

	emp.Status = status
	emp.UpdatedAt = time.Now()

	if status == "fired" {
		now := time.Now()
		emp.FiredAt = &now
	} else {
		emp.FiredAt = nil
	}

	r.employees[id] = emp
	return &emp, nil
}

func (r *MemoryRepository) GetPositions(ctx context.Context) ([]string, error) {
	return r.positions, nil
}

func (r *MemoryRepository) GetEmployeeStats(ctx context.Context) (map[string]interface{}, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	stats := make(map[string]interface{})
	statusCount := make(map[string]int)
	deptCount := make(map[string]int)

	total := 0
	for _, emp := range r.employees {
		total++
		statusCount[emp.Status]++
		deptCount[emp.DepartmentID]++
	}

	stats["total"] = total
	stats["by_status"] = statusCount
	stats["by_department"] = deptCount

	return stats, nil
}

func containsIgnoreCase(str, substr string) bool {
	return contains(str, substr) // Простая реализация
}

func contains(str, substr string) bool {
	return len(str) >= len(substr) && str[:len(substr)] == substr
}

// СЕРВИСНЫЙ СЛОЙ


type EmployeeService struct {
	repo Repository
}

func NewEmployeeService(repo Repository) *EmployeeService {
	return &EmployeeService{repo: repo}
}

func (s *EmployeeService) GetDepartments(ctx context.Context) ([]Department, error) {
	slog.DebugContext(ctx, "getting departments")
	return s.repo.GetDepartments(ctx)
}

func (s *EmployeeService) GetEmployeesByDepartment(ctx context.Context, departmentID string) ([]Employee, error) {
	slog.DebugContext(ctx, "getting employees by department", "department_id", departmentID)
	return s.repo.GetEmployeesByDepartment(ctx, departmentID)
}

func (s *EmployeeService) SearchEmployees(ctx context.Context, req EmployeeSearchRequest) ([]Employee, error) {
	slog.DebugContext(ctx, "searching employees", "filters", req)
	return s.repo.SearchEmployees(ctx, req)
}

func (s *EmployeeService) CreateEmployee(ctx context.Context, emp Employee) (*Employee, error) {
	slog.DebugContext(ctx, "creating employee", "employee", emp.FullName)
	
	// Валидация
	if err := s.validateEmployee(emp); err != nil {
		return nil, err
	}

	return s.repo.CreateEmployee(ctx, emp)
}

func (s *EmployeeService) UpdateEmployee(ctx context.Context, emp Employee) (*Employee, error) {
	slog.DebugContext(ctx, "updating employee", "employee_id", emp.ID)
	
	if emp.ID == "" {
		return nil, fmt.Errorf("ID сотрудника обязателен")
	}

	if err := s.validateEmployee(emp); err != nil {
		return nil, err
	}

	return s.repo.UpdateEmployee(ctx, emp)
}

func (s *EmployeeService) UpdateEmployeeStatus(ctx context.Context, id string, status string) (*Employee, error) {
	slog.DebugContext(ctx, "updating employee status", "employee_id", id, "status", status)
	
	validStatuses := map[string]bool{"active": true, "vacation": true, "fired": true}
	if !validStatuses[status] {
		return nil, fmt.Errorf("неверный статус: %s", status)
	}

	return s.repo.UpdateEmployeeStatus(ctx, id, status)
}

func (s *EmployeeService) GetPositions(ctx context.Context) ([]string, error) {
	slog.DebugContext(ctx, "getting positions")
	return s.repo.GetPositions(ctx)
}

func (s *EmployeeService) GetEmployeeStats(ctx context.Context) (map[string]interface{}, error) {
	return s.repo.GetEmployeeStats(ctx)
}

func (s *EmployeeService) validateEmployee(emp Employee) error {
	if emp.FullName == "" {
		return fmt.Errorf("ФИО обязательно")
	}
	if emp.Gender == "" {
		return fmt.Errorf("пол обязателен")
	}
	if emp.Age < 18 || emp.Age > 70 {
		return fmt.Errorf("возраст должен быть от 18 до 70 лет")
	}
	if emp.Education == "" {
		return fmt.Errorf("образование обязательно")
	}
	if emp.Position == "" {
		return fmt.Errorf("должность обязательна")
	}
	if emp.Passport == "" {
		return fmt.Errorf("паспортные данные обязательны")
	}
	if emp.DepartmentID == "" {
		return fmt.Errorf("департамент обязателен")
	}

	validGenders := map[string]bool{"male": true, "female": true}
	if !validGenders[emp.Gender] {
		return fmt.Errorf("неверный пол: %s", emp.Gender)
	}

	validEducation := map[string]bool{"secondary": true, "specialized": true, "higher": true}
	if !validEducation[emp.Education] {
		return fmt.Errorf("неверное образование: %s", emp.Education)
	}

	return nil
}

// METRICS (PROMETHEUS)

var (
	HttpRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total number of HTTP requests",
	}, []string{"method", "path", "status"})

	HttpRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "HTTP request duration in seconds",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "path"})

	EmployeesTotal = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "employees_total",
		Help: "Total number of employees",
	})

	EmployeesByStatus = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "employees_by_status",
		Help: "Number of employees by status",
	}, []string{"status"})
)

func InitMetrics() {
	// Метрики автоматически регистрируются при импорте
}

func UpdateEmployeeMetrics(stats map[string]interface{}) {
	if total, ok := stats["total"].(int); ok {
		EmployeesTotal.Set(float64(total))
	}

	if byStatus, ok := stats["by_status"].(map[string]int); ok {
		for status, count := range byStatus {
			EmployeesByStatus.WithLabelValues(status).Set(float64(count))
		}
	}
}

// TRACING (OPENTELEMETRY)

func InitTracer(jaegerURL, serviceName string) (*sdktrace.TracerProvider, error) {
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(jaegerURL)))
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(serviceName),
		)),
	)

	otel.SetTracerProvider(tp)
	return tp, nil
}

// HTTP HANDLERS

type Handler struct {
	service *EmployeeService
	tracer  trace.Tracer
}

func NewHandler(service *EmployeeService) *Handler {
	return &Handler{
		service: service,
		tracer:  otel.Tracer("employee-handler"),
	}
}

func (h *Handler) InitRoutes() *gin.Engine {
	router := gin.New()

	// Middleware
	router.Use(h.loggingMiddleware())
	router.Use(h.tracingMiddleware())
	router.Use(gin.Recovery())

	// API routes
	api := router.Group("/api")
	{
		api.GET("/departments", h.getDepartments)
		api.GET("/employees/department/:departmentId", h.getEmployeesByDepartment)
		api.POST("/employees/search", h.searchEmployees)
		api.POST("/employees", h.createEmployee)
		api.PUT("/employees/:id", h.updateEmployee)
		api.PATCH("/employees/:id/status", h.updateEmployeeStatus)
		api.GET("/positions", h.getPositions)
		api.GET("/metrics", h.getMetrics)
		api.GET("/health", h.healthCheck)
	}

	// Prometheus metrics
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

// Статические файлы (фронтенд)
router.StaticFS("/static", http.FS(staticFiles))

// Главная страница - отдаем index.html
router.GET("/", func(c *gin.Context) {
    data, err := staticFiles.ReadFile("static/index.html")
    if err != nil {
        c.String(http.StatusInternalServerError, "Ошибка загрузки страницы")
        return
    }
    c.Data(http.StatusOK, "text/html; charset=utf-8", data)
})

	return router
}

// Middleware
func (h *Handler) loggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start)

		// Логирование в JSON формате
		slog.Info("HTTP request",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"duration", duration.String(),
			"client_ip", c.ClientIP(),
		)

		// Метрики
		HttpRequestsTotal.WithLabelValues(
			c.Request.Method,
			c.Request.URL.Path,
			strconv.Itoa(c.Writer.Status()),
		).Inc()

		HttpRequestDuration.WithLabelValues(
			c.Request.Method,
			c.Request.URL.Path,
		).Observe(duration.Seconds())
	}
}

func (h *Handler) tracingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, span := h.tracer.Start(c.Request.Context(), fmt.Sprintf("%s %s", c.Request.Method, c.Request.URL.Path))
		defer span.End()

		span.SetAttributes(
			attribute.String("http.method", c.Request.Method),
			attribute.String("http.path", c.Request.URL.Path),
			attribute.String("http.client_ip", c.ClientIP()),
		)

		c.Request = c.Request.WithContext(ctx)
		c.Next()

		span.SetAttributes(
			attribute.Int("http.status_code", c.Writer.Status()),
		)
	}
}

// Handlers
func (h *Handler) getDepartments(c *gin.Context) {
	ctx := c.Request.Context()
	departments, err := h.service.GetDepartments(ctx)
	if err != nil {
		h.sendError(c, http.StatusInternalServerError, "Ошибка получения департаментов: "+err.Error())
		return
	}
	h.sendSuccess(c, departments)
}

func (h *Handler) getEmployeesByDepartment(c *gin.Context) {
	ctx := c.Request.Context()
	departmentID := c.Param("departmentId")
	
	employees, err := h.service.GetEmployeesByDepartment(ctx, departmentID)
	if err != nil {
		h.sendError(c, http.StatusInternalServerError, "Ошибка получения сотрудников: "+err.Error())
		return
	}
	h.sendSuccess(c, employees)
}

func (h *Handler) searchEmployees(c *gin.Context) {
	ctx := c.Request.Context()
	var req EmployeeSearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.sendError(c, http.StatusBadRequest, "Неверный формат запроса: "+err.Error())
		return
	}

	employees, err := h.service.SearchEmployees(ctx, req)
	if err != nil {
		h.sendError(c, http.StatusInternalServerError, "Ошибка поиска сотрудников: "+err.Error())
		return
	}
	h.sendSuccess(c, employees)
}

func (h *Handler) createEmployee(c *gin.Context) {
	ctx := c.Request.Context()
	var emp Employee
	if err := c.ShouldBindJSON(&emp); err != nil {
		h.sendError(c, http.StatusBadRequest, "Неверный формат данных: "+err.Error())
		return
	}

	createdEmp, err := h.service.CreateEmployee(ctx, emp)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "сотрудник с таким паспортом уже существует" {
			status = http.StatusBadRequest
		}
		h.sendError(c, status, "Ошибка создания сотрудника: "+err.Error())
		return
	}

	h.sendSuccessWithMessage(c, createdEmp, "Сотрудник успешно создан")
}

func (h *Handler) updateEmployee(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")
	
	var emp Employee
	if err := c.ShouldBindJSON(&emp); err != nil {
		h.sendError(c, http.StatusBadRequest, "Неверный формат данных: "+err.Error())
		return
	}

	emp.ID = id
	updatedEmp, err := h.service.UpdateEmployee(ctx, emp)
	if err != nil {
		h.sendError(c, http.StatusInternalServerError, "Ошибка обновления сотрудника: "+err.Error())
		return
	}

	h.sendSuccessWithMessage(c, updatedEmp, "Данные сотрудника обновлены")
}

func (h *Handler) updateEmployeeStatus(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")
	
	var req StatusUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.sendError(c, http.StatusBadRequest, "Неверный формат данных: "+err.Error())
		return
	}

	updatedEmp, err := h.service.UpdateEmployeeStatus(ctx, id, req.Status)
	if err != nil {
		h.sendError(c, http.StatusInternalServerError, "Ошибка обновления статуса: "+err.Error())
		return
	}

	message := "Статус сотрудника обновлен"
	switch req.Status {
	case "active":
		message = "Сотрудник активирован"
	case "vacation":
		message = "Сотрудник отправлен в отпуск"
	case "fired":
		message = "Сотрудник уволен"
	}

	h.sendSuccessWithMessage(c, updatedEmp, message)
}

func (h *Handler) getPositions(c *gin.Context) {
	ctx := c.Request.Context()
	positions, err := h.service.GetPositions(ctx)
	if err != nil {
		h.sendError(c, http.StatusInternalServerError, "Ошибка получения должностей: "+err.Error())
		return
	}
	h.sendSuccess(c, positions)
}

func (h *Handler) getMetrics(c *gin.Context) {
	ctx := c.Request.Context()
	stats, err := h.service.GetEmployeeStats(ctx)
	if err != nil {
		h.sendError(c, http.StatusInternalServerError, "Ошибка получения метрик: "+err.Error())
		return
	}

	// Обновляем Prometheus метрики
	UpdateEmployeeMetrics(stats)

	h.sendSuccess(c, map[string]interface{}{
		"timestamp": time.Now(),
		"stats":     stats,
		"message":   "Метрики обновлены",
	})
}

func (h *Handler) healthCheck(c *gin.Context) {
	h.sendSuccess(c, map[string]string{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
		"service":   "employee-management-system",
	})
}

// Вспомогательные методы
func (h *Handler) sendSuccess(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    data,
	})
}

func (h *Handler) sendSuccessWithMessage(c *gin.Context, data interface{}, message string) {
	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    data,
		Message: message,
	})
}

func (h *Handler) sendError(c *gin.Context, status int, message string) {
	slog.Error("API error", 
		"status", status, 
		"message", message,
		"path", c.Request.URL.Path,
	)

	c.JSON(status, APIResponse{
		Success: false,
		Error:   message,
	})
}

// MAIN FUNCTION

func main() {
	// Настройка JSON логгера
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

// Трассировка отключена (Jaeger не требуется для работы)
slog.Info("Трассировка отключена - Jaeger не запущен")

	// Инициализация метрик
	InitMetrics()

	// Инициализация сервисов
	repo := NewMemoryRepository()
	service := NewEmployeeService(repo)
	handler := NewHandler(service)

	// Создание HTTP сервера
	server := &http.Server{
		Addr:         ":8080",
		Handler:      handler.InitRoutes(),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Запуск сервера
	go func() {
		slog.Info("Запуск сервера", "port", "8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Ошибка запуска сервера", "error", err)
			os.Exit(1)
		}
	}()

	// Ожидание сигнала для graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("Завершение работы сервера...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("Ошибка завершения работы сервера", "error", err)
	}

	slog.Info("Сервер остановлен")
}

