package _postgres

import (
	"practice-3/internal/repository"
	"practice-3/internal/repository/_postgres/users"
)

func NewRepositories(d *Dialect) *repository.Repositories {
	return &repository.Repositories{
		UserRepository: users.NewUserRepository(d.DB),
	}
}
