package regions

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(
	ctx context.Context,
	region SafeRegion,
) error {
	start := time.Now()

	defer func() {
		log.Printf("[DB] CreateSafeRegion: %v", time.Since(start))
	}()

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO safe_regions (
			id,
			user_id,
			group_id,
			name,
			latitude,
			longitude,
			radius_meters,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`,
		region.ID,
		region.UserID,
		region.GroupID,
		region.Name,
		region.Latitude,
		region.Longitude,
		region.RadiusMeters,
		region.CreatedAt.UTC(),
		region.UpdatedAt.UTC(),
	)

	return err
}

func (r *Repository) Fetch(
	ctx context.Context,
	userID string,
) ([]SafeRegion, error) {
	start := time.Now()

	defer func() {
		log.Printf("[DB] FetchSafeRegions: %v", time.Since(start))
	}()

	rows, err := r.db.QueryContext(ctx, `
		SELECT
			id,
			user_id,
			group_id,
			name,
			latitude,
			longitude,
			radius_meters,
			created_at,
			updated_at
		FROM safe_regions
		WHERE user_id = $1
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var regions []SafeRegion

	for rows.Next() {
		var region SafeRegion

		if err := rows.Scan(
			&region.ID,
			&region.UserID,
			&region.GroupID,
			&region.Name,
			&region.Latitude,
			&region.Longitude,
			&region.RadiusMeters,
			&region.CreatedAt,
			&region.UpdatedAt,
		); err != nil {
			return nil, err
		}

		region.CreatedAt = region.CreatedAt.UTC()
		region.UpdatedAt = region.UpdatedAt.UTC()

		regions = append(regions, region)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return regions, nil
}

func (r *Repository) FetchByID(
	ctx context.Context,
	id string,
	userID string,
) (SafeRegion, error) {
	start := time.Now()

	defer func() {
		log.Printf("[DB] FetchSafeRegionByID: %v", time.Since(start))
	}()

	var region SafeRegion

	err := r.db.QueryRowContext(ctx, `
		SELECT
			id,
			user_id,
			group_id,
			name,
			latitude,
			longitude,
			radius_meters,
			created_at,
			updated_at
		FROM safe_regions
		WHERE id = $1
		  AND user_id = $2
	`,
		id,
		userID,
	).Scan(
		&region.ID,
		&region.UserID,
		&region.GroupID,
		&region.Name,
		&region.Latitude,
		&region.Longitude,
		&region.RadiusMeters,
		&region.CreatedAt,
		&region.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return SafeRegion{}, sql.ErrNoRows
		}

		return SafeRegion{}, err
	}

	region.CreatedAt = region.CreatedAt.UTC()
	region.UpdatedAt = region.UpdatedAt.UTC()

	return region, nil
}

func (r *Repository) Update(
	ctx context.Context,
	id string,
	userID string,
	request UpdateRegionDTO,
) (SafeRegion, error) {
	start := time.Now()

	defer func() {
		log.Printf("[DB] UpdateSafeRegion: %v", time.Since(start))
	}()

	set := []string{}
	args := []any{}
	arg := 1

	if request.GroupID != nil {
		set = append(set, fmt.Sprintf("group_id = $%d", arg))
		args = append(args, *request.GroupID)
		arg++
	}

	if request.Name != nil {
		set = append(set, fmt.Sprintf("name = $%d", arg))
		args = append(args, *request.Name)
		arg++
	}

	if request.Latitude != nil {
		set = append(set, fmt.Sprintf("latitude = $%d", arg))
		args = append(args, *request.Latitude)
		arg++
	}

	if request.Longitude != nil {
		set = append(set, fmt.Sprintf("longitude = $%d", arg))
		args = append(args, *request.Longitude)
		arg++
	}

	if request.RadiusMeters != nil {
		set = append(set, fmt.Sprintf("radius_meters = $%d", arg))
		args = append(args, *request.RadiusMeters)
		arg++
	}

	if len(set) == 0 {
		return r.FetchByID(ctx, id, userID)
	}

	set = append(set, fmt.Sprintf("updated_at = $%d", arg))
	args = append(args, time.Now().UTC())
	arg++

	args = append(args, id, userID)

	query := fmt.Sprintf(
		`UPDATE safe_regions
		 SET %s
		 WHERE id = $%d
		   AND user_id = $%d
		 RETURNING
			id,
			user_id,
			group_id,
			name,
			latitude,
			longitude,
			radius_meters,
			created_at,
			updated_at`,
		strings.Join(set, ", "),
		arg,
		arg+1,
	)

	var updatedRegion SafeRegion

	err := r.db.QueryRowContext(
		ctx,
		query,
		args...,
	).Scan(
		&updatedRegion.ID,
		&updatedRegion.UserID,
		&updatedRegion.GroupID,
		&updatedRegion.Name,
		&updatedRegion.Latitude,
		&updatedRegion.Longitude,
		&updatedRegion.RadiusMeters,
		&updatedRegion.CreatedAt,
		&updatedRegion.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return SafeRegion{}, sql.ErrNoRows
		}

		return SafeRegion{}, err
	}

	updatedRegion.CreatedAt = updatedRegion.CreatedAt.UTC()
	updatedRegion.UpdatedAt = updatedRegion.UpdatedAt.UTC()

	return updatedRegion, nil
}

func (r *Repository) Delete(
	ctx context.Context,
	regionID string,
	userID string,
) error {
	start := time.Now()

	defer func() {
		log.Printf("[DB] DeleteSafeRegion: %v", time.Since(start))
	}()

	result, err := r.db.ExecContext(ctx, `
		DELETE FROM safe_regions
		WHERE id = $1
		  AND user_id = $2
	`, regionID, userID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *Repository) FetchByGroupID(
	ctx context.Context,
	groupID string,
	userID string,
) ([]SafeRegion, error) {
	start := time.Now()

	defer func() {
		log.Printf("[DB] FetchSafeRegionsByGroupID: %v", time.Since(start))
	}()

	rows, err := r.db.QueryContext(ctx, `
		SELECT
			id,
			user_id,
			group_id,
			name,
			latitude,
			longitude,
			radius_meters,
			created_at,
			updated_at
		FROM safe_regions
		WHERE group_id = $1
		  AND user_id = $2
		ORDER BY created_at DESC
	`, groupID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var regions []SafeRegion

	for rows.Next() {
		var region SafeRegion

		if err := rows.Scan(
			&region.ID,
			&region.UserID,
			&region.GroupID,
			&region.Name,
			&region.Latitude,
			&region.Longitude,
			&region.RadiusMeters,
			&region.CreatedAt,
			&region.UpdatedAt,
		); err != nil {
			return nil, err
		}

		region.CreatedAt = region.CreatedAt.UTC()
		region.UpdatedAt = region.UpdatedAt.UTC()

		regions = append(regions, region)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return regions, nil
}
