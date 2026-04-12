package usecase

import "practice-7/internal/entity"

type UserInterface interface {
	Register(*entity.User) error
	Login(*entity.LoginUserDTO) (string, string, error)
	GetMe(string) (*entity.User, error)
	Verify(string, string) error
	Promote(string) error
}