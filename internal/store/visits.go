package store

import (
	"context"
	"dooreye-backend/internal/model"
	"fmt"
	"time"

	"github.com/jackc/pgx"
)

func (db *DB) CheckoutVisit(ctx context.Context, visitID int64) error {
	query := `
		UPDATE visits
		SET is_active = false,
			check_out_time = $1
		WHERE id = $2 AND is_active = true`

	result, err := db.pool.Exec(ctx, query, time.Now(), visitID)
	if err != nil {
		return fmt.Errorf("updating visit: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

type VisitFilter struct {
	ResidenceID *int64 // pointer to handle empty case
	OnlyOngoing bool
}

func (db *DB) GetVisits(ctx context.Context, filter VisitFilter) ([]model.VisitWithVisitor, error) {
	query := `
        SELECT v.id, v.residence_id, v.visitor_id, v.checked_in_by,
               v.check_in_time, v.check_out_time, v.purpose,
               vis.name, vis.phone, vis.type
        FROM visits v
        JOIN visitors vis ON v.visitor_id = vis.id
        WHERE 1=1
    `
	args := []interface{}{}
	argCount := 1

	if filter.ResidenceID != nil {
		query += fmt.Sprintf(" AND v.residence_id = $%d", argCount)
		args = append(args, *filter.ResidenceID)
		argCount++
	}

	if filter.OnlyOngoing {
		query += " AND v.check_out_time IS NULL"
	}

	query += " ORDER BY v.check_in_time DESC"

	rows, err := db.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying visits: %w", err)
	}
	defer rows.Close()

	var visits []model.VisitWithVisitor
	for rows.Next() {
		var v model.VisitWithVisitor
		if err := rows.Scan(
			&v.ID, &v.ResidenceID, &v.VisitorID, &v.CheckedInBy,
			&v.CheckInTime, &v.CheckOutTime, &v.Purpose,
			&v.Name, &v.Phone, &v.Type,
		); err != nil {
			return nil, fmt.Errorf("scanning visit row: %w", err)
		}
		visits = append(visits, v)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating visits: %w", err)
	}

	return visits, nil
}

func (db *DB) GetVisitorByPhone(ctx context.Context, phone string) (*model.Visitor, error) {
	var visitor model.Visitor

	err := db.pool.QueryRow(ctx, `
    SELECT id, name, phone, photo_url, type, pre_approved_till, created_by
    FROM visitors
    WHERE phone = $1
    ORDER BY created_at DESC
    LIMIT 1
    `, phone).Scan(
		&visitor.ID, &visitor.Name, &visitor.Phone, &visitor.PhotoURL,
		&visitor.Type, &visitor.PreApprovedTill, &visitor.CreatedBy,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("getting visitor by phone: %w", err)
	}

	return &visitor, nil
}

type PreApprovedVisitor struct {
	ID              int64      `json:"id"`
	Name            string     `json:"name"`
	Phone           string     `json:"phone"`
	PhotoURL        string     `json:"photo_url"`
	Type            string     `json:"type"`
	PreApprovedTill *time.Time `json:"pre_approved_till"`
	CreatedBy       string     `json:"created_by"`
}

func (db *DB) CreatePreApprovedVisitor(ctx context.Context, input PreApprovedVisitor) (*PreApprovedVisitor, error) {
	query := `
        INSERT INTO visitors (
            name, phone, photo_url, type, pre_approved_till, created_by
        )
        VALUES ($1, $2, $3, $4, $5, $6)
        RETURNING id, name, phone, photo_url, type, pre_approved_till, created_by
    `

	var visitor PreApprovedVisitor
	err := db.pool.QueryRow(ctx, query,
		input.Name,
		input.Phone,
		input.PhotoURL,
		input.Type,
		input.PreApprovedTill,
		input.CreatedBy,
	).Scan(
		&visitor.ID,
		&visitor.Name,
		&visitor.Phone,
		&visitor.PhotoURL,
		&visitor.Type,
		&visitor.PreApprovedTill,
		&visitor.CreatedBy,
	)

	if err != nil {
		return nil, fmt.Errorf("creating pre-approved visitor: %w", err)
	}

	return &visitor, nil
}
