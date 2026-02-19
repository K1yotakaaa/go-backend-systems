package users

import (
	"database/sql"
	"errors"

	"practice-3/pkg/modules"

	"github.com/jmoiron/sqlx"
)

var ErrUserNotFound = errors.New("user not found")

type Repository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetUsers(limit, offset int) ([]modules.User, error) {
	var out []modules.User
	q := `
		SELECT id,name,email,age,created_at,updated_at,deleted_at,password_hash
		FROM users
		WHERE deleted_at IS NULL
		ORDER BY id
		LIMIT $1 OFFSET $2;
	`
	if err := r.db.Select(&out, q, limit, offset); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *Repository) GetUserByID(id int) (*modules.User, error) {
	var u modules.User
	q := `
		SELECT id,name,email,age,created_at,updated_at,deleted_at,password_hash
		FROM users
		WHERE id=$1 AND deleted_at IS NULL;
	`
	if err := r.db.Get(&u, q, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (r *Repository) CreateUser(u modules.User) (int, error) {
	var id int
	q := `INSERT INTO users (name,email,age,password_hash) VALUES ($1,$2,$3,$4) RETURNING id;`
	if err := r.db.QueryRow(q, u.Name, u.Email, u.Age, u.PasswordHash).Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

func (r *Repository) UpdateUser(id int, u modules.User) error {
	q := `UPDATE users SET name=$1,email=$2,age=$3,updated_at=now() WHERE id=$4 AND deleted_at IS NULL;`
	res, err := r.db.Exec(q, u.Name, u.Email, u.Age, id)
	if err != nil {
		return err
	}
	ra, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if ra == 0 {
		return ErrUserNotFound
	}
	return nil
}

func (r *Repository) DeleteUser(id int) (int64, error) {
	q := `UPDATE users SET deleted_at=now(), updated_at=now() WHERE id=$1 AND deleted_at IS NULL;`
	res, err := r.db.Exec(q, id)
	if err != nil {
		return 0, err
	}
	ra, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}
	if ra == 0 {
		return 0, ErrUserNotFound
	}
	return ra, nil
}

func (r *Repository) CreateUserWithAuditTx(u modules.User, action string) (int, error) {
	tx, err := r.db.Beginx()
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback() }()

	var id int
	if err := tx.QueryRow(
		`INSERT INTO users (name,email,age,password_hash) VALUES ($1,$2,$3,$4) RETURNING id;`,
		u.Name, u.Email, u.Age, u.PasswordHash,
	).Scan(&id); err != nil {
		return 0, err
	}

	if _, err := tx.Exec(`INSERT INTO audit_logs (user_id,action) VALUES ($1,$2);`, id, action); err != nil {
		return 0, err
	}

	return id, tx.Commit()
}
