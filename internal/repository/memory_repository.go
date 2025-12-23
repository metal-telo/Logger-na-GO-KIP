package repository

import (
	"context"

	"employee-management/internal/models"
)

// Repository defines the interface for data access
type Repository interface {
	GetDepartments(ctx context.Context) ([]models.Department, error)
	GetEmployeesByDepartment(ctx context.Context, departmentID string) ([]models.Employee, error)
	SearchEmployees(ctx context.Context, req models.EmployeeSearchRequest) ([]models.Employee, error)
	CreateEmployee(ctx context.Context, emp models.Employee) (*models.Employee, error)
	UpdateEmployee(ctx context.Context, emp models.Employee) (*models.Employee, error)
	UpdateEmployeeStatus(ctx context.Context, id string, status string) (*models.Employee, error)
	GetPositions(ctx context.Context) ([]string, error)
	GetEmployeeStats(ctx context.Context) (map[string]interface{}, error)
}

