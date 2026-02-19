package usecase

import (
	"errors"

	"practice-3/internal/repository"
	"practice-3/internal/repository/_postgres/users"
	"practice-3/pkg/modules"

	"golang.org/x/crypto/bcrypt"
)

type UserUsecase struct {
	repo repository.UserRepository
}

func NewUserUsecase(repo repository.UserRepository) *UserUsecase {
	return &UserUsecase{repo: repo}
}

func (u *UserUsecase) GetUsers(limit, offset int) ([]modules.User, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	return u.repo.GetUsers(limit, offset)
}

func (u *UserUsecase) GetUserByID(id int) (*modules.User, error) {
	return u.repo.GetUserByID(id)
}

func (u *UserUsecase) CreateUser(name, email string, age int, password *string) (int, error) {
	var hash *string
	if password != nil && *password != "" {
		b, err := bcrypt.GenerateFromPassword([]byte(*password), bcrypt.DefaultCost)
		if err != nil {
			return 0, err
		}
		s := string(b)
		hash = &s
	}
	return u.repo.CreateUser(modules.User{Name: name, Email: email, Age: age, PasswordHash: hash})
}

func (u *UserUsecase) UpdateUser(id int, name, email string, age int) error {
	return u.repo.UpdateUser(id, modules.User{Name: name, Email: email, Age: age})
}

func (u *UserUsecase) DeleteUser(id int) (int64, error) {
	return u.repo.DeleteUser(id)
}

func IsNotFound(err error) bool {
	return errors.Is(err, users.ErrUserNotFound)
}
