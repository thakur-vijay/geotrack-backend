package groups

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/thakur-vijay/geotrack-backend/internal/modules/users"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) Create(
	ctx context.Context,
	group Group,
	creatorID string,
) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin create group transaction: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `
		INSERT INTO groups (
			id,
			name,
			created_by,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5)
	`,
		group.ID,
		group.Name,
		creatorID,
		group.CreatedAt.UTC(),
		group.UpdatedAt.UTC(),
	)
	if err != nil {
		return fmt.Errorf("insert group: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO group_members (
			group_id,
			user_id
		)
		VALUES ($1, $2)
	`,
		group.ID,
		creatorID,
	)
	if err != nil {
		return fmt.Errorf("insert group creator: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit create group transaction: %w", err)
	}

	return nil
}

func (r *Repository) GetMembers(
	ctx context.Context,
	groupID string,
) ([]users.User, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			u.id,
			u.full_name,
			u.email,
			u.created_at,
			u.updated_at
		FROM users u
		INNER JOIN group_members gm
			ON gm.user_id = u.id
		WHERE gm.group_id = $1
		ORDER BY gm.joined_at ASC
	`, groupID)
	if err != nil {
		return nil, fmt.Errorf("query group members: %w", err)
	}
	defer rows.Close()

	members := make([]users.User, 0)

	for rows.Next() {
		var user users.User

		err := rows.Scan(
			&user.ID,
			&user.FullName,
			&user.Email,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan group member: %w", err)
		}

		user.CreatedAt = user.CreatedAt.UTC()
		user.UpdatedAt = user.UpdatedAt.UTC()

		members = append(members, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate group members: %w", err)
	}

	return members, nil
}

func (r *Repository) GetByID(
	ctx context.Context,
	id string,
) (Group, error) {
	var group Group

	err := r.db.QueryRowContext(ctx, `
		SELECT
			id,
			name,
			created_by,
			created_at,
			updated_at
		FROM groups
		WHERE id = $1
	`, id).Scan(
		&group.ID,
		&group.Name,
		&group.CreatedBy,
		&group.CreatedAt,
		&group.UpdatedAt,
	)
	if err != nil {
		return Group{}, err
	}

	members, err := r.GetMembers(ctx, group.ID)
	if err != nil {
		return Group{}, fmt.Errorf("get group members: %w", err)
	}

	group.Members = members
	group.CreatedAt = group.CreatedAt.UTC()
	group.UpdatedAt = group.UpdatedAt.UTC()

	return group, nil
}

func (r *Repository) GetByUserID(
	ctx context.Context,
	userID string,
) ([]Group, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			g.id,
			g.name,
			g.created_by,
			g.created_at,
			g.updated_at
		FROM groups g
		INNER JOIN group_members gm
			ON gm.group_id = g.id
		WHERE gm.user_id = $1
		ORDER BY g.created_at DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("query user groups: %w", err)
	}
	defer rows.Close()

	groups := make([]Group, 0)

	for rows.Next() {
		var group Group

		err := rows.Scan(
			&group.ID,
			&group.Name,
			&group.CreatedBy,
			&group.CreatedAt,
			&group.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan user group: %w", err)
		}

		group.CreatedAt = group.CreatedAt.UTC()
		group.UpdatedAt = group.UpdatedAt.UTC()

		groups = append(groups, group)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate user groups: %w", err)
	}

	for i := range groups {
		members, err := r.GetMembers(ctx, groups[i].ID)
		if err != nil {
			return nil, fmt.Errorf(
				"get members for group %s: %w",
				groups[i].ID,
				err,
			)
		}

		groups[i].Members = members
	}

	return groups, nil
}

func (r *Repository) AddMember(
	ctx context.Context,
	groupID string,
	userID string,
) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO group_members (
			group_id,
			user_id
		)
		VALUES ($1, $2)
		ON CONFLICT (group_id, user_id) DO NOTHING
	`, groupID, userID)
	if err != nil {
		return fmt.Errorf("add group member: %w", err)
	}

	return nil
}

func (r *Repository) RemoveMember(
	ctx context.Context,
	groupID string,
	userID string,
) error {
	result, err := r.db.ExecContext(ctx, `
		DELETE FROM group_members
		WHERE group_id = $1
		  AND user_id = $2
	`, groupID, userID)
	if err != nil {
		return fmt.Errorf("remove group member: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get removed member count: %w", err)
	}

	if affected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *Repository) Delete(
	ctx context.Context,
	groupID string,
) error {
	result, err := r.db.ExecContext(ctx, `
		DELETE FROM groups
		WHERE id = $1
	`, groupID)
	if err != nil {
		return fmt.Errorf("delete group: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get deleted group count: %w", err)
	}

	if affected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

var _ = time.Time{}
