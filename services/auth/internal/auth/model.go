package auth

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Name     string `json:"name" binding:"required,min=2"`
}

type AuthResponse struct {
	Token     string  `json:"token"`
	ExpiresIn float64 `json:"expires_in"`
	User      UserDTO `json:"user"`
}

type UserDTO struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}