package errors

type KernelError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (r KernelError) Error() string {
	return r.Message
}

func New(code string, message string) *KernelError {
	return &KernelError{code, message}
}
