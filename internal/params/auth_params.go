package params

type (
	RegisterUserRequest struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	LoginUserRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	ProcessTokenResponse struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
)
