package domain

import "time"

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