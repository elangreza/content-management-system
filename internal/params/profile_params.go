package params

import "time"

type (
	UserProfileResponse struct {
		Name      string     `json:"name"`
		Email     string     `json:"email"`
		Role      int64      `json:"role"`
		RoleName  string     `json:"role_name"`
		CreatedAt time.Time  `json:"created_at"`
		UpdatedAt *time.Time `json:"updated_at"`
	}
)
