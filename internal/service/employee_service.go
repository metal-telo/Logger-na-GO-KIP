package service

import (
	"context"
	"fmt"
	"log/slog"

	"employee-management-system/internal/domain"
)

type EmployeeService struct {
	repo domain.Repository
}

func NewEmployeeService(repo domain.Repository) *EmployeeService {
	return &EmployeeService{repo: repo}
}

func (s *EmployeeService) GetDepartments(ctx context.Context) ([]domain.Department, error) {
	slog.DebugContext(ctx, "getting departments")
	return s.repo.GetDepartments(ctx)
}

func (s *EmployeeService) GetEmployeesByDepartment(ctx context.Context, departmentID string) ([]domain.Employee, error) {
	slog.DebugContext(ctx, "getting employees by department", "department_id", departmentID)
	return s.repo.GetEmployeesByDepartment(ctx, departmentID)
}

func (s *EmployeeService) SearchEmployees(ctx context.Context, req domain.EmployeeSearchRequest) ([]domain.Employee, error) {
	slog.DebugContext(ctx, "searching employees", "filters", req)
	return s.repo.SearchEmployees(ctx, req)
}

func (s *EmployeeService) CreateEmployee(ctx context.Context, emp domain.Employee) (*domain.Employee, error) {
	slog.DebugContext(ctx, "creating employee", "employee", emp.FullName)
	
	if err := s.validateEmployee(emp); err != nil {
		return nil, err
	}

	return s.repo.CreateEmployee(ctx, emp)
}

func (s *EmployeeService) UpdateEmployee(ctx context.Context, emp domain.Employee) (*domain.Employee, error) {
	slog.DebugContext(ctx, "updating employee", "employee_id", emp.ID)
	
	if emp.ID == "" {
		return nil, fmt.Errorf("ID сотрудника обязателен")
	}

	if err := s.validateEmployee(emp); err != nil {
		return nil, err
	}

	return s.repo.UpdateEmployee(ctx, emp)
}

func (s *EmployeeService) UpdateEmployeeStatus(ctx context.Context, id string, status string) (*domain.Employee, error) {
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

func (s *EmployeeService) validateEmployee(emp domain.Employee) error {
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