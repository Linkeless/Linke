package service

// stringToPtr converts string to *string, returns nil for empty strings
// This helper is used for JSON fields that should be NULL when empty
func stringToPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// ptrToString converts *string to string, returns empty string for nil
func ptrToString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}