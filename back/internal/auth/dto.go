package auth

import "time"

type RegisterRequest struct {
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
	Password    string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserResponse struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	DisplayName   string `json:"display_name"`
	EmailVerified bool   `json:"email_verified"`
	CreatedAt     string `json:"created_at"`
}

type AuthResponse struct {
	User      UserResponse `json:"user"`
	CSRFToken string       `json:"csrf_token"`
}

func toUserResponse(user User) UserResponse {
	return UserResponse{
		ID:            user.ID.String(),
		Email:         user.Email,
		DisplayName:   user.DisplayName,
		EmailVerified: user.EmailVerifiedAt != nil,
		CreatedAt:     user.CreatedAt.Format(time.RFC3339),
	}
}