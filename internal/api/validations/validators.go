package validations

type ValidationError struct {
	Message string            `json:"message"`
	Details map[string]string `json:"data"`
}

func (e ValidationError) Error() string {
	return e.Message
}
