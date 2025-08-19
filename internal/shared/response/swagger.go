package response

// RESTful API response structures for Swagger documentation

// HALCollectionResponse represents a HAL-style collection response for Swagger documentation
// @Description HAL-style collection response format
type HALCollectionResponse struct {
	Embedded map[string]any  `json:"_embedded" swaggertype:"object"` // Embedded resources
	Links    HALLinksSwagger `json:"_links"`                         // Navigation links
	Page     PageInfoSwagger `json:"page,omitempty"`                 // Page information
	Total    int64           `json:"total,omitempty" example:"100"`  // Total number of items
}

// HALResourceResponse represents a HAL-style resource response for Swagger documentation
// @Description HAL-style resource response format
type HALResourceResponse struct {
	ID    int             `json:"id" example:"1"`         // Resource ID
	Name  string          `json:"name" example:"example"` // Resource name
	Links HALLinksSwagger `json:"_links,omitempty"`       // Navigation links
}

// HALLinksSwagger represents HAL navigation links for Swagger documentation
// @Description HAL navigation links structure
type HALLinksSwagger struct {
	Self  *HALLinkSwagger `json:"self,omitempty"`  // Self link
	First *HALLinkSwagger `json:"first,omitempty"` // First page link
	Prev  *HALLinkSwagger `json:"prev,omitempty"`  // Previous page link
	Next  *HALLinkSwagger `json:"next,omitempty"`  // Next page link
	Last  *HALLinkSwagger `json:"last,omitempty"`  // Last page link
}

// HALLinkSwagger represents a single HAL link for Swagger documentation
// @Description HAL link structure
type HALLinkSwagger struct {
	Href      string `json:"href" example:"/api/v1/users/1"`            // Link URL
	Templated bool   `json:"templated,omitempty" example:"false"`       // Whether the link is templated
	Type      string `json:"type,omitempty" example:"application/json"` // Media type hint
	Title     string `json:"title,omitempty" example:"User Profile"`    // Human-readable title
}

// PageInfoSwagger represents pagination metadata for Swagger documentation
// @Description Pagination metadata structure
type PageInfoSwagger struct {
	Size          int   `json:"size" example:"10"`           // Items per page
	TotalElements int64 `json:"totalElements" example:"100"` // Total number of elements
	TotalPages    int   `json:"totalPages" example:"10"`     // Total number of pages
	Number        int   `json:"number" example:"0"`          // Current page number (0-based)
}

// ProblemJSONSwagger represents RFC 9457 Problem Details for Swagger documentation
// @Description RFC 9457 Problem Details for HTTP APIs
type ProblemJSONSwagger struct {
	Type     string `json:"type" example:"/problems/not-found"`                            // Problem type URI
	Title    string `json:"title" example:"Not Found"`                                     // Short summary
	Status   int    `json:"status" example:"404"`                                          // HTTP status code
	Detail   string `json:"detail,omitempty" example:"The user with id 123 was not found"` // Detailed explanation
	Instance string `json:"instance,omitempty" example:"/api/v1/users/123"`                // Problem instance URI
}

// BadRequestResponse represents a 400 Bad Request problem response
// @Description Bad Request problem response format
type BadRequestResponse struct {
	Type     string            `json:"type" example:"/problems/bad-request"`
	Title    string            `json:"title" example:"Bad Request"`
	Status   int               `json:"status" example:"400"`
	Detail   string            `json:"detail" example:"Invalid request parameters"`
	Instance string            `json:"instance,omitempty" example:"/api/v1/users"`
	Errors   map[string]string `json:"errors,omitempty" example:"{\"email\":\"Invalid email format\"}"`
}

// UnauthorizedResponse represents a 401 Unauthorized problem response
// @Description Unauthorized problem response format
type UnauthorizedResponse struct {
	Type     string `json:"type" example:"/problems/unauthorized"`
	Title    string `json:"title" example:"Unauthorized"`
	Status   int    `json:"status" example:"401"`
	Detail   string `json:"detail" example:"Authentication credentials are required"`
	Instance string `json:"instance,omitempty" example:"/api/v1/users"`
}

// ForbiddenResponse represents a 403 Forbidden problem response
// @Description Forbidden problem response format
type ForbiddenResponse struct {
	Type     string `json:"type" example:"/problems/forbidden"`
	Title    string `json:"title" example:"Forbidden"`
	Status   int    `json:"status" example:"403"`
	Detail   string `json:"detail" example:"Access to this resource is denied"`
	Instance string `json:"instance,omitempty" example:"/api/v1/admin/users"`
}

// NotFoundResponse represents a 404 Not Found problem response
// @Description Not Found problem response format
type NotFoundResponse struct {
	Type     string `json:"type" example:"/problems/not-found"`
	Title    string `json:"title" example:"Not Found"`
	Status   int    `json:"status" example:"404"`
	Detail   string `json:"detail" example:"The requested resource was not found"`
	Instance string `json:"instance,omitempty" example:"/api/v1/users/123"`
}

// ConflictResponse represents a 409 Conflict problem response
// @Description Conflict problem response format
type ConflictResponse struct {
	Type     string `json:"type" example:"/problems/conflict"`
	Title    string `json:"title" example:"Conflict"`
	Status   int    `json:"status" example:"409"`
	Detail   string `json:"detail" example:"A user with this email already exists"`
	Instance string `json:"instance,omitempty" example:"/api/v1/users"`
}

// UnprocessableEntityResponse represents a 422 Unprocessable Entity problem response
// @Description Unprocessable Entity problem response format
type UnprocessableEntityResponse struct {
	Type     string            `json:"type" example:"/problems/unprocessable-entity"`
	Title    string            `json:"title" example:"Unprocessable Entity"`
	Status   int               `json:"status" example:"422"`
	Detail   string            `json:"detail" example:"The request was well-formed but contains semantic errors"`
	Instance string            `json:"instance,omitempty" example:"/api/v1/users"`
	Errors   map[string]string `json:"errors,omitempty" example:"{\"email\":\"Email is already taken\"}"`
}

// InternalServerErrorResponse represents a 500 Internal Server Error problem response
// @Description Internal Server Error problem response format
type InternalServerErrorResponse struct {
	Type     string `json:"type" example:"/problems/internal-server-error"`
	Title    string `json:"title" example:"Internal Server Error"`
	Status   int    `json:"status" example:"500"`
	Detail   string `json:"detail" example:"An internal server error occurred"`
	Instance string `json:"instance,omitempty" example:"/api/v1/users"`
}

// Legacy response structures (deprecated, for backward compatibility)

// StandardResponse - DEPRECATED: Use direct resource responses instead
// @Description DEPRECATED: Standard API response format
type StandardResponse struct {
	Code    int    `json:"code" example:"0"`                    // Response code (0 for success, non-zero for errors)
	Message string `json:"message" example:"success"`           // Response message
	Data    any    `json:"data,omitempty" swaggertype:"object"` // Response data (optional)
}

// StandardErrorResponse - DEPRECATED: Use ProblemJSON responses instead
// @Description DEPRECATED: Standard API error response format
type StandardErrorResponse struct {
	Code    int    `json:"code" example:"4000"`           // Error code
	Message string `json:"message" example:"Bad Request"` // Error message
}
