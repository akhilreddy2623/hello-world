package app

type ErrorResponse struct {
	Type       string `json:"type,omitempty" example:"error"`
	Message    string `json:"message,omitempty" example:"some error happened"`
	StatusCode int    `json:"-"`
}
