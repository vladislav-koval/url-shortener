package response

type ErrorResponse struct {
	Code    string       `json:"code" example:"validation_failed"`
	Message string       `json:"message" example:"request is invalid"`
	Details []FieldError `json:"details,omitempty"`
}

type FieldError struct {
	Field string `json:"field" example:"url"`
	Rule  string `json:"rule" example:"max"`
	Param string `json:"param,omitempty" example:"2048"`
}
