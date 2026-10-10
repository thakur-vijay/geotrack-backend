package users

import (
	"context"
	"database/sql"
	"log"
	"time"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, user User) error {
	start := time.Now()

	defer func() {
		log.Printf("[DB] CreateUser: %v", time.Since(start))
	}()

	_, err := r.db.ExecContext(ctx, `
        INSERT INTO users (
            id,
            full_name,
            email,
			password_hash,
            created_at,
            updated_at
        )
        VALUES ($1, $2, $3, $4, $5, $6)
    `,
		user.ID,
		user.FullName,
		user.Email,
		user.PasswordHash,
		user.CreatedAt.UTC(),
		user.UpdatedAt.UTC(),
	)

	return err
}

func (r *Repository) GetByID(ctx context.Context, id string) (*User, error) {
	var user User

	err := r.db.QueryRowContext(ctx, `
		SELECT id, full_name, email, created_at, updated_at
		FROM users
		WHERE id = $1
	`, id).Scan(
		&user.ID,
		&user.FullName,
		&user.Email,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	user.CreatedAt = user.CreatedAt.UTC()
	user.UpdatedAt = user.UpdatedAt.UTC()

	return &user, nil
}

func (r *Repository) GetByEmail(
	ctx context.Context,
	email string,
) (*User, error) {
	var user User

	err := r.db.QueryRowContext(ctx, `
		SELECT id, full_name, email, password_hash, created_at, updated_at
		FROM users
		WHERE email = $1
	`, email).Scan(
		&user.ID,
		&user.FullName,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	user.CreatedAt = user.CreatedAt.UTC()
	user.UpdatedAt = user.UpdatedAt.UTC()

	return &user, nil
}

func (r *Repository) FetchUsers(
	ctx context.Context,
	groupID string,
) ([]User, error) {
	start := time.Now()

	defer func() {
		log.Printf("[DB] FetchUsers: %v", time.Since(start))
	}()

	rows, err := r.db.QueryContext(ctx, `
		SELECT
			u.id,
			u.full_name,
			u.email,
			u.created_at,
			u.updated_at
		FROM users u
		WHERE NOT EXISTS (
			SELECT 1
			FROM group_members gm
			WHERE gm.group_id = $1
			  AND gm.user_id = u.id
		)
		ORDER BY u.full_name ASC
	`, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]User, 0)

	for rows.Next() {
		var user User

		if err := rows.Scan(
			&user.ID,
			&user.FullName,
			&user.Email,
			&user.CreatedAt,
			&user.UpdatedAt,
		); err != nil {
			return nil, err
		}

		user.CreatedAt = user.CreatedAt.UTC()
		user.UpdatedAt = user.UpdatedAt.UTC()

		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}
