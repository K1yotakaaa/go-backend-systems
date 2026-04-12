package repo

import (
	"practice-7/internal/entity"
	"practice-7/pkg/postgres"
)

type UserRepo struct {
	PG *postgres.Postgres
}

func NewUserRepo(pg *postgres.Postgres) *UserRepo {
	return &UserRepo{pg}
}

func (r *UserRepo) Create(u *entity.User) error {
	return r.PG.Conn.Create(u).Error
}

func (r *UserRepo) GetByUsername(username string) (*entity.User, error) {
	var u entity.User
	err := r.PG.Conn.Where("username = ?", username).First(&u).Error
	return &u, err
}

func (r *UserRepo) GetByID(id string) (*entity.User, error) {
	var u entity.User
	err := r.PG.Conn.First(&u, "id = ?", id).Error
	return &u, err
}

func (r *UserRepo) Verify(email, code string) error {
	var u entity.User
	r.PG.Conn.Where("email = ?", email).First(&u)

	if u.VerifyCode != code {
		return nil
	}

	return r.PG.Conn.Model(&u).Update("verified", true).Error
}

func (r *UserRepo) Promote(id string) error {
	return r.PG.Conn.Model(&entity.User{}).
		Where("id = ?", id).
		Update("role", "admin").Error
}