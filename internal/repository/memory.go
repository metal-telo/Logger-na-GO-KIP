package repository

import (
	"context"
	"fmt"
	"sync"
	"time"

	"employee-management/internal/models"
)

// MemoryRepository is an in-memory implementation of Repository
type MemoryRepository struct {
	mu          sync.RWMutex
	departments map[string]models.Department
	employees   map[string]models.Employee
	positions   []string
}

// NewMemoryRepository creates a new in-memory repository with test data
func NewMemoryRepository() *MemoryRepository {
	repo := &MemoryRepository{
		departments: make(map[string]models.Department),
		employees:   make(map[string]models.Employee),
		positions: []string{
			"Программист", "Аналитик", "Тестировщик", "Менеджер по продажам",
			"HR-менеджер", "Бухгалтер", "Маркетолог", "Дизайнер",
			"Системный администратор", "Руководитель отдела",
		},
	}
	repo.initTestData()
	return repo
}

func (r *MemoryRepository) initTestData() {
	now := time.Now()
	depts := []models.Department{
		{ID: "dept1", Name: "IT-департамент", Description: "Разработка ПО", CreatedAt: now},
		{ID: "dept2", Name: "Отдел продаж", Description: "Продажи и маркетинг", CreatedAt: now},
		{ID: "dept3", Name: "HR-отдел", Description: "Управление персоналом", CreatedAt: now},
		{ID: "dept4", Name: "Финансовый отдел", Description: "Финансы и бухгалтерия", CreatedAt: now},
		{ID: "dept5", Name: "Маркетинг", Description: "Маркетинг и реклама", CreatedAt: now},
	}

	for _, dept := range depts {
		r.departments[dept.ID] = dept
	}

	employees := []models.Employee{
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

func (r *MemoryRepository) GetDepartments(ctx context.Context) ([]models.Department, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var departments []models.Department
	for _, dept := range r.departments {
		departments = append(departments, dept)
	}
	return departments, nil
}

func (r *MemoryRepository) GetEmployeesByDepartment(ctx context.Context, departmentID string) ([]models.Employee, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var employees []models.Employee
	for _, emp := range r.employees {
		if emp.DepartmentID == departmentID {
			employees = append(employees, emp)
		}
	}
	return employees, nil
}

func (r *MemoryRepository) SearchEmployees(ctx context.Context, req models.EmployeeSearchRequest) ([]models.Employee, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var employees []models.Employee
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

func (r *MemoryRepository) CreateEmployee(ctx context.Context, emp models.Employee) (*models.Employee, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, existing := range r.employees {
		if existing.Passport == emp.Passport {
			return nil, fmt.Errorf("сотрудник с таким паспортом уже существует")
		}
	}

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

func (r *MemoryRepository) UpdateEmployee(ctx context.Context, emp models.Employee) (*models.Employee, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, exists := r.employees[emp.ID]
	if !exists {
		return nil, fmt.Errorf("сотрудник не найден")
	}

	for _, e := range r.employees {
		if e.ID != emp.ID && e.Passport == emp.Passport {
			return nil, fmt.Errorf("сотрудник с таким паспортом уже существует")
		}
	}

	emp.CreatedAt = existing.CreatedAt
	emp.UpdatedAt = time.Now()
	emp.Status = existing.Status

	r.employees[emp.ID] = emp
	return &emp, nil
}

func (r *MemoryRepository) UpdateEmployeeStatus(ctx context.Context, id string, status string) (*models.Employee, error) {
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

func contains(str, substr string) bool {
	return len(str) >= len(substr) && str[:len(substr)] == substr
}

