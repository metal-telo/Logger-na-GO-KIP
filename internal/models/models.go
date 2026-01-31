package models

import "time"

// Department represents a department in the organization
type Department struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

// Employee represents an employee
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

// EmployeeSearchRequest represents search filters for employees
type EmployeeSearchRequest struct {
	FullName  string `json:"full_name"`
	Position  string `json:"position"`
	Gender    string `json:"gender"`
	Education string `json:"education"`
	AgeFrom   *int   `json:"age_from,omitempty"`
	AgeTo     *int   `json:"age_to,omitempty"`
}

// StatusUpdateRequest represents a status update request
type StatusUpdateRequest struct {
	Status string `json:"status"`
}

// APIResponse represents a standard API response
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
}






