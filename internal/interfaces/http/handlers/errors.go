package handlers

// ErrorResponse represents a standard API error response.
// Used by Swagger for documentation purposes.
type ErrorResponse struct {
	Code    string `json:"code" example:"INVALID_INPUT"`
	Message string `json:"message" example:"validation failed"`
}
