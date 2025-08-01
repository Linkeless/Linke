package valueobject

import (
	sharedvo "linke/internal/shared/valueobject"
)

// ConvertToSharedUserID converts user domain UserID to shared UserID
func ConvertToSharedUserID(id UserID) (sharedvo.UserID, error) {
	// Check if it's a UUID format
	if id.IsUUID() {
		return sharedvo.NewUserIDFromString(id.String())
	}
	
	// Handle legacy uint format - check if ToUint() returns a valid number
	if numericID := id.ToUint(); numericID > 0 {
		return sharedvo.NewUserIDFromUint(numericID)
	}
	
	// Fallback: try string conversion
	return sharedvo.NewUserIDFromString(id.String())
}

// ConvertFromSharedUserID converts shared UserID to user domain UserID
func ConvertFromSharedUserID(id sharedvo.UserID) (UserID, error) {
	// Check if shared UserID is UUID format
	if id.IsUUID() {
		return NewUserIDFromString(id.String())
	}
	
	// Handle legacy uint format
	if id.IsNumeric() {
		numericID := id.ToUint()
		return NewUserIDFromUint(numericID), nil
	}
	
	// Fallback: use string value
	return NewUserIDFromString(id.String())
}

// ConvertToSharedUserIDFromUint converts uint to shared UserID (for legacy support)
func ConvertToSharedUserIDFromUint(id uint) (sharedvo.UserID, error) {
	if id == 0 {
		return sharedvo.UserID{}, nil // Allow zero for compatibility
	}
	return sharedvo.NewUserIDFromUint(id)
}

// ConvertFromSharedUserIDToUint converts shared UserID to uint (for legacy support)
func ConvertFromSharedUserIDToUint(id sharedvo.UserID) uint {
	if id.IsZero() {
		return 0
	}
	return id.ToUint()
}