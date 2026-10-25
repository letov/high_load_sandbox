package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/letov/high_load_sandbox/internal/model"
)

var ErrNotFound = errors.New("not found")

var ErrLoginTaken = errors.New("login already taken")

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, u model.User, passwordHash string) (int64, error) {
	const query = `
		INSERT INTO users (login, password_hash, first_name, last_name, birth_date, gender, interests, city)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`

	var id int64
	err := r.db.QueryRowContext(ctx, query,
		u.Login, passwordHash, u.FirstName, u.LastName, u.BirthDate, u.Gender, u.Interests, u.City,
	).Scan(&id)
	if err != nil {
		if isUniqueViolation(err) {
			return 0, ErrLoginTaken
		}
		return 0, fmt.Errorf("insert user: %w", err)
	}

	return id, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id int64) (model.User, error) {
	const query = `
		SELECT id, login, first_name, last_name, birth_date::text, gender, interests, city
		FROM users
		WHERE id = $1
	`

	var u model.User
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&u.ID, &u.Login, &u.FirstName, &u.LastName, &u.BirthDate, &u.Gender, &u.Interests, &u.City,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return model.User{}, ErrNotFound
	}
	if err != nil {
		return model.User{}, fmt.Errorf("select user: %w", err)
	}

	return u, nil
}

func (r *UserRepository) GetPasswordHashByLogin(ctx context.Context, login string) (int64, string, error) {
	const query = `SELECT id, password_hash FROM users WHERE login = $1`

	var (
		id   int64
		hash string
	)
	err := r.db.QueryRowContext(ctx, query, login).Scan(&id, &hash)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, "", ErrNotFound
	}
	if err != nil {
		return 0, "", fmt.Errorf("select password hash: %w", err)
	}

	return id, hash, nil
}

func isUniqueViolation(err error) bool {
	type pgError interface {
		SQLState() string
	}
	var pgErr pgError
	if errors.As(err, &pgErr) {
		return pgErr.SQLState() == "23505"
	}
	return false
}
