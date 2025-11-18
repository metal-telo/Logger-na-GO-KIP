package domain

import "context"

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