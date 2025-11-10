package dto

type CreateUserRequest struct {
	Email string `json:"email" validate:"required,email" binding:"required"`
	Name  string `json:"name" validate:"required,min=2,max=100" binding:"required"`
}

type UpdateUserRequest struct {
	Name string `json:"name" validate:"omitempty,min=2,max=100"`
}

type UserResponse struct {
	ID        uint   `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}