package repository

import (
	"context"
	"database/sql"

	"bulbul/internal/database"
	"bulbul/internal/models"
)

type UserRepository struct {
	db *database.DB
}

func NewUserRepository(db *database.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) GetByID(ctx context.Context, id int64) (*models.User, error) {
	user := &models.User{}
	query := `
		SELECT user_id, email, password_hash, password_plain, first_name, surname, 
		       birthday, registered_at, is_active, last_logged_in
		FROM users
		WHERE user_id = $1`

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.UserID,
		&user.Email,
		&user.PasswordHash,
		&user.PasswordPlain,
		&user.FirstName,
		&user.Surname,
		&user.Birthday,
		&user.RegisteredAt,
		&user.IsActive,
		&user.LastLoggedIn,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	return user, err
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	user := &models.User{}
	query := `
		SELECT user_id, email, password_hash, password_plain, first_name, surname, 
		       birthday, registered_at, is_active, last_logged_in
		FROM users
		WHERE email = $1`

	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.UserID,
		&user.Email,
		&user.PasswordHash,
		&user.PasswordPlain,
		&user.FirstName,
		&user.Surname,
		&user.Birthday,
		&user.RegisteredAt,
		&user.IsActive,
		&user.LastLoggedIn,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	return user, err
}

func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (email, password_hash, password_plain, first_name, surname, birthday, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING user_id, registered_at, last_logged_in`

	err := r.db.QueryRowContext(ctx, query,
		user.Email,
		user.PasswordHash,
		user.PasswordPlain,
		user.FirstName,
		user.Surname,
		user.Birthday,
		user.IsActive,
	).Scan(&user.UserID, &user.RegisteredAt, &user.LastLoggedIn)

	return err
}