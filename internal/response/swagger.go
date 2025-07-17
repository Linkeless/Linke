package response

// StandardResponse represents the standard API response structure for Swagger documentation
// @Description Standard API response format
type StandardResponse struct {
	Code    int         `json:"code" example:"0"`                     // Response code (0 for success, non-zero for errors)
	Message string      `json:"message" example:"success"`            // Response message
	Data    interface{} `json:"data,omitempty" swaggertype:"object"` // Response data (optional)
}

// StandardErrorResponse represents error response structure for Swagger documentation
// @Description Standard API error response format
type StandardErrorResponse struct {
	Code    int    `json:"code" example:"4000"`        // Error code
	Message string `json:"message" example:"Bad Request"` // Error message
}

// BadRequestResponse represents a 400 Bad Request response
// @Description Bad Request response format
type BadRequestResponse struct {
	Code    int    `json:"code" example:"4000"`
	Message string `json:"message" example:"Invalid request parameters"`
}

// UnauthorizedResponse represents a 401 Unauthorized response
// @Description Unauthorized response format
type UnauthorizedResponse struct {
	Code    int    `json:"code" example:"4001"`
	Message string `json:"message" example:"Authentication required"`
}

// ForbiddenResponse represents a 403 Forbidden response
// @Description Forbidden response format
type ForbiddenResponse struct {
	Code    int    `json:"code" example:"4003"`
	Message string `json:"message" example:"Access denied"`
}

// NotFoundResponse represents a 404 Not Found response
// @Description Not Found response format
type NotFoundResponse struct {
	Code    int    `json:"code" example:"4004"`
	Message string `json:"message" example:"Resource not found"`
}

// ConflictResponse represents a 409 Conflict response
// @Description Conflict response format
type ConflictResponse struct {
	Code    int    `json:"code" example:"4009"`
	Message string `json:"message" example:"Resource already exists"`
}

// InternalServerErrorResponse represents a 500 Internal Server Error response
// @Description Internal Server Error response format
type InternalServerErrorResponse struct {
	Code    int    `json:"code" example:"5000"`
	Message string `json:"message" example:"Internal server error"`
}

// StandardListResponse represents paginated list response structure for Swagger documentation
// @Description Standard API paginated list response format
type StandardListResponse struct {
	Code    int          `json:"code" example:"0"`
	Message string       `json:"message" example:"success"`
	Data    ListDataInfo `json:"data"`
}

// ListDataInfo represents the data structure for list responses
// @Description List response data structure
type ListDataInfo struct {
	Items      interface{}         `json:"items" swaggertype:"array,object"`      // List items
	Pagination PaginationResponse `json:"pagination"` // Pagination information
}

// MessageOnlyResponse represents a response with only a message (no data)
// @Description Message only response format
type MessageOnlyResponse struct {
	Code    int    `json:"code" example:"0"`
	Message string `json:"message" example:"Operation completed successfully"`
}

// SearchResponse represents a search response with pagination and query info
// @Description Search response format with pagination and query information
type SearchResponse struct {
	Code    int                `json:"code" example:"0"`
	Message string             `json:"message" example:"Search completed"`
	Data    SearchResponseData `json:"data"`
}

// SearchResponseData represents the data structure for search responses
// @Description Search response data structure
type SearchResponseData struct {
	Items      interface{}         `json:"items" swaggertype:"array,object"`      // Search result items
	Pagination PaginationResponse `json:"pagination"` // Pagination information
	Query      string             `json:"query" example:"search term"` // Search query
}

// ProviderFilterResponse represents a provider filter response with pagination and provider info
// @Description Provider filter response format with pagination and provider information
type ProviderFilterResponse struct {
	Code    int                       `json:"code" example:"0"`
	Message string                    `json:"message" example:"Users retrieved successfully"`
	Data    ProviderFilterResponseData `json:"data"`
}

// ProviderFilterResponseData represents the data structure for provider filter responses
// @Description Provider filter response data structure
type ProviderFilterResponseData struct {
	Items      interface{}         `json:"items" swaggertype:"array,object"`      // Filtered items
	Pagination PaginationResponse `json:"pagination"` // Pagination information
	Provider   string             `json:"provider" example:"google"` // OAuth provider
}