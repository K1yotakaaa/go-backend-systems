package entity

type CreateUserDTO struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginUserDTO struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type VerifyDTO struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}