package schema


type (
	ErrResponse struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
)
