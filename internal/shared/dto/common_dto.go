package dto

import "time"

// BaseDTO represents common fields that all DTOs might have
type BaseDTO struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// PaginationDTO represents pagination information
type PaginationDTO struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
	Total    int `json:"total"`
	Pages    int `json:"pages"`
}

// ErrorDTO represents error information
type ErrorDTO struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}