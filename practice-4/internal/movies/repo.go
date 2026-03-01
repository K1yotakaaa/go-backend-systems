package movies

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repo struct {
	DB *pgxpool.Pool
}

func (r Repo) List(ctx context.Context) ([]Movie, error) {
	rows, err := r.DB.Query(ctx, `SELECT id, title, genre, budget, created_at FROM movies ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Movie
	for rows.Next() {
		var m Movie
		if err := rows.Scan(&m.ID, &m.Title, &m.Genre, &m.Budget, &m.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (r Repo) Get(ctx context.Context, id int) (Movie, bool, error) {
	var m Movie
	err := r.DB.QueryRow(ctx,
		`SELECT id, title, genre, budget, created_at FROM movies WHERE id=$1`, id,
	).Scan(&m.ID, &m.Title, &m.Genre, &m.Budget, &m.CreatedAt)

	if err == pgx.ErrNoRows {
		return Movie{}, false, nil
	}
	if err != nil {
		return Movie{}, false, err
	}
	return m, true, nil
}

func (r Repo) Create(ctx context.Context, title, genre string, budget int64) (Movie, error) {
	var m Movie
	err := r.DB.QueryRow(ctx, `
		INSERT INTO movies (title, genre, budget)
		VALUES ($1, $2, $3)
		RETURNING id, title, genre, budget, created_at
	`, title, genre, budget).Scan(&m.ID, &m.Title, &m.Genre, &m.Budget, &m.CreatedAt)
	return m, err
}

func (r Repo) Update(ctx context.Context, id int, title, genre string, budget int64) (Movie, bool, error) {
	var m Movie
	err := r.DB.QueryRow(ctx, `
		UPDATE movies
		SET title=$2, genre=$3, budget=$4
		WHERE id=$1
		RETURNING id, title, genre, budget, created_at
	`, id, title, genre, budget).Scan(&m.ID, &m.Title, &m.Genre, &m.Budget, &m.CreatedAt)

	if err == pgx.ErrNoRows {
		return Movie{}, false, nil
	}
	if err != nil {
		return Movie{}, false, err
	}
	return m, true, nil
}

func (r Repo) Delete(ctx context.Context, id int) (bool, error) {
	ct, err := r.DB.Exec(ctx, `DELETE FROM movies WHERE id=$1`, id)
	if err != nil {
		return false, err
	}
	return ct.RowsAffected() > 0, nil
}