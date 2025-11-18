package handler

import (
	"embed"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"employee-management-system/internal/domain"
	"employee-management-system/internal/service"
	"employee-management-system/internal/telemetry"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

//go:embed ../../static/*
var staticFiles embed.FS

type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
}

type Handler struct {
	service *service.EmployeeService
	tracer  trace.Tracer
	server  *http.Server
}

func NewHandler(service *service.EmployeeService) *Handler {
	return &Handler{
		service: service,
		tracer:  otel.Tracer("employee-handler"),
	}
}

func (h *Handler) StartServer(addr string) *http.Server {
	router := h.InitRoutes()

	server := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		slog.Info("Запуск сервера", "port", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Ошибка запуска сервера", "error", err)
		}
	}()

	h.server = server
	return server
}

func (h *Handler) Shutdown() {
	if h.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		h.server.Shutdown(ctx)
	}
}

func (h *Handler) InitRoutes() *gin.Engine {
	router := gin.New()

	router.Use(h.loggingMiddleware())
	router.Use(h.tracingMiddleware())
	router.Use(gin.Recovery())

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

	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	router.StaticFS("/static", http.FS(staticFiles))

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

func (h *Handler) loggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start)

		slog.Info("HTTP request",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"duration", duration.String(),
			"client_ip", c.ClientIP(),
		)

		telemetry.HttpRequestsTotal.WithLabelValues(
			c.Request.Method,
			c.Request.URL.Path,
			strconv.Itoa(c.Writer.Status()),
		).Inc()

		telemetry.HttpRequestDuration.WithLabelValues(
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
	var req domain.EmployeeSearchRequest
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
	var emp domain.Employee
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
	
	var emp domain.Employee
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
	
	var req domain.StatusUpdateRequest
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

	telemetry.UpdateEmployeeMetrics(stats)

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