package services

import "fmt"

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error on field '%s': %s", e.Field, e.Message)
}

// DuplicateError represents a duplicate entry error
type DuplicateError struct {
	Entity string
	Name   string
}

func (e *DuplicateError) Error() string {
	return fmt.Sprintf("%s with name '%s' already exists", e.Entity, e.Name)
}

// NotFoundError represents a not found error
type NotFoundError struct {
	Entity string
	ID     interface{}
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s with id '%v' not found", e.Entity, e.ID)
}
