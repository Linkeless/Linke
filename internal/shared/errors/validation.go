package errors

import (
	"fmt"
	"strings"
)

// ValidationError represents field validation errors
type ValidationError struct {
	*BusinessError
	Fields []FieldError `json:"fields,omitempty"`
}

// FieldError represents a single field validation error
type FieldError struct {
	Field       string      `json:"field"`
	Value       interface{} `json:"value,omitempty"`
	Message     string      `json:"message"`
	Code        string      `json:"code"`
	Constraints interface{} `json:"constraints,omitempty"`
}

// Error implements the error interface
func (ve *ValidationError) Error() string {
	if len(ve.Fields) == 0 {
		return ve.BusinessError.Error()
	}

	var fieldMessages []string
	for _, field := range ve.Fields {
		fieldMessages = append(fieldMessages, fmt.Sprintf("%s: %s", field.Field, field.Message))
	}

	return fmt.Sprintf("%s: %s", ve.BusinessError.Error(), strings.Join(fieldMessages, ", "))
}

// Unwrap returns the underlying error
func (ve *ValidationError) Unwrap() error {
	return ve.BusinessError
}

// AddField adds a field error to the validation error
func (ve *ValidationError) AddField(field, message, code string) *ValidationError {
	ve.Fields = append(ve.Fields, FieldError{
		Field:   field,
		Message: message,
		Code:    code,
	})
	return ve
}

// AddFieldWithValue adds a field error with the invalid value
func (ve *ValidationError) AddFieldWithValue(field string, value interface{}, message, code string) *ValidationError {
	ve.Fields = append(ve.Fields, FieldError{
		Field:   field,
		Value:   value,
		Message: message,
		Code:    code,
	})
	return ve
}

// AddFieldWithConstraints adds a field error with constraints
func (ve *ValidationError) AddFieldWithConstraints(field string, value interface{}, message, code string, constraints interface{}) *ValidationError {
	ve.Fields = append(ve.Fields, FieldError{
		Field:       field,
		Value:       value,
		Message:     message,
		Code:        code,
		Constraints: constraints,
	})
	return ve
}

// HasFields returns true if the validation error has field errors
func (ve *ValidationError) HasFields() bool {
	return len(ve.Fields) > 0
}

// GetFieldErrors returns all field errors
func (ve *ValidationError) GetFieldErrors() []FieldError {
	return ve.Fields
}

// GetFieldError returns the field error for a specific field
func (ve *ValidationError) GetFieldError(field string) (FieldError, bool) {
	for _, fe := range ve.Fields {
		if fe.Field == field {
			return fe, true
		}
	}
	return FieldError{}, false
}

// NewValidationError creates a new validation error
func NewValidationError(message string) *ValidationError {
	return &ValidationError{
		BusinessError: New(ErrCodeInvalidInput, message),
		Fields:        make([]FieldError, 0),
	}
}

// NewValidationErrorf creates a new validation error with formatted message
func NewValidationErrorf(format string, args ...interface{}) *ValidationError {
	return &ValidationError{
		BusinessError: Newf(ErrCodeInvalidInput, format, args...),
		Fields:        make([]FieldError, 0),
	}
}

// Common validation error codes
const (
	ValidationCodeRequired     = "required"
	ValidationCodeMinLength    = "min_length"
	ValidationCodeMaxLength    = "max_length"
	ValidationCodePattern      = "pattern"
	ValidationCodeInvalidEmail = "invalid_email"
	ValidationCodeInvalidURL   = "invalid_url"
	ValidationCodeInvalidUUID  = "invalid_uuid"
	ValidationCodeMin          = "min"
	ValidationCodeMax          = "max"
	ValidationCodeInvalidEnum  = "invalid_enum"
	ValidationCodeInvalidType  = "invalid_type"
	ValidationCodeDuplicate    = "duplicate"
)

// Common validation constraints
type Constraints struct {
	Min       *int      `json:"min,omitempty"`
	Max       *int      `json:"max,omitempty"`
	MinLength *int      `json:"min_length,omitempty"`
	MaxLength *int      `json:"max_length,omitempty"`
	Pattern   *string   `json:"pattern,omitempty"`
	Enum      []string  `json:"enum,omitempty"`
	Type      *string   `json:"type,omitempty"`
}

// Predefined validation errors for common cases
var (
	ErrValidationRequired     = "This field is required"
	ErrValidationMinLength    = "This field must be at least %d characters"
	ErrValidationMaxLength    = "This field must be at most %d characters"
	ErrValidationPattern      = "This field must match the required pattern"
	ErrValidationInvalidEmail = "This field must be a valid email address"
	ErrValidationInvalidURL   = "This field must be a valid URL"
	ErrValidationInvalidUUID  = "This field must be a valid UUID"
	ErrValidationMin          = "This field must be at least %v"
	ErrValidationMax          = "This field must be at most %v"
	ErrValidationInvalidEnum  = "This field must be one of: %s"
	ErrValidationInvalidType  = "This field must be of type %s"
	ErrValidationDuplicate    = "This field must be unique"
)

// Validation helper functions

// RequiredField validates that a field is not empty
func RequiredField(field string, value interface{}) *FieldError {
	if value == nil {
		return &FieldError{
			Field:   field,
			Value:   value,
			Message: ErrValidationRequired,
			Code:    ValidationCodeRequired,
		}
	}

	switch v := value.(type) {
	case string:
		if strings.TrimSpace(v) == "" {
			return &FieldError{
				Field:   field,
				Value:   value,
				Message: ErrValidationRequired,
				Code:    ValidationCodeRequired,
			}
		}
	case []interface{}:
		if len(v) == 0 {
			return &FieldError{
				Field:   field,
				Value:   value,
				Message: ErrValidationRequired,
				Code:    ValidationCodeRequired,
			}
		}
	}

	return nil
}

// MinLengthField validates minimum length for strings
func MinLengthField(field string, value string, minLength int) *FieldError {
	if len(value) < minLength {
		return &FieldError{
			Field:   field,
			Value:   value,
			Message: fmt.Sprintf(ErrValidationMinLength, minLength),
			Code:    ValidationCodeMinLength,
			Constraints: Constraints{
				MinLength: &minLength,
			},
		}
	}
	return nil
}

// MaxLengthField validates maximum length for strings
func MaxLengthField(field string, value string, maxLength int) *FieldError {
	if len(value) > maxLength {
		return &FieldError{
			Field:   field,
			Value:   value,
			Message: fmt.Sprintf(ErrValidationMaxLength, maxLength),
			Code:    ValidationCodeMaxLength,
			Constraints: Constraints{
				MaxLength: &maxLength,
			},
		}
	}
	return nil
}

// EnumField validates that a value is in a list of allowed values
func EnumField(field string, value string, allowedValues []string) *FieldError {
	for _, allowed := range allowedValues {
		if value == allowed {
			return nil
		}
	}

	return &FieldError{
		Field:   field,
		Value:   value,
		Message: fmt.Sprintf(ErrValidationInvalidEnum, strings.Join(allowedValues, ", ")),
		Code:    ValidationCodeInvalidEnum,
		Constraints: Constraints{
			Enum: allowedValues,
		},
	}
}

// IsValidationError checks if an error is a ValidationError
func IsValidationError(err error) bool {
	_, ok := err.(*ValidationError)
	return ok
}

// AsValidationError attempts to extract a ValidationError from an error
func AsValidationError(err error) (*ValidationError, bool) {
	if ve, ok := err.(*ValidationError); ok {
		return ve, true
	}
	return nil, false
}

// ValidateAndCollect validates multiple fields and collects errors
type FieldValidator struct {
	validationError *ValidationError
}

// NewFieldValidator creates a new field validator
func NewFieldValidator() *FieldValidator {
	return &FieldValidator{
		validationError: NewValidationError("Validation failed"),
	}
}

// Required validates a required field
func (fv *FieldValidator) Required(field string, value interface{}) *FieldValidator {
	if fieldErr := RequiredField(field, value); fieldErr != nil {
		fv.validationError.Fields = append(fv.validationError.Fields, *fieldErr)
	}
	return fv
}

// MinLength validates minimum length
func (fv *FieldValidator) MinLength(field string, value string, minLength int) *FieldValidator {
	if fieldErr := MinLengthField(field, value, minLength); fieldErr != nil {
		fv.validationError.Fields = append(fv.validationError.Fields, *fieldErr)
	}
	return fv
}

// MaxLength validates maximum length
func (fv *FieldValidator) MaxLength(field string, value string, maxLength int) *FieldValidator {
	if fieldErr := MaxLengthField(field, value, maxLength); fieldErr != nil {
		fv.validationError.Fields = append(fv.validationError.Fields, *fieldErr)
	}
	return fv
}

// Enum validates enum values
func (fv *FieldValidator) Enum(field string, value string, allowedValues []string) *FieldValidator {
	if fieldErr := EnumField(field, value, allowedValues); fieldErr != nil {
		fv.validationError.Fields = append(fv.validationError.Fields, *fieldErr)
	}
	return fv
}

// Custom adds a custom validation error
func (fv *FieldValidator) Custom(field, message, code string) *FieldValidator {
	fv.validationError.Fields = append(fv.validationError.Fields, FieldError{
		Field:   field,
		Message: message,
		Code:    code,
	})
	return fv
}

// HasErrors returns true if there are validation errors
func (fv *FieldValidator) HasErrors() bool {
	return len(fv.validationError.Fields) > 0
}

// Error returns the validation error if there are any errors
func (fv *FieldValidator) Error() error {
	if fv.HasErrors() {
		return fv.validationError
	}
	return nil
}