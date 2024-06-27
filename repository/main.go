package repository

import (
	"context"
	"fmt"
	"hime-backend/db"
	"hime-backend/models"

	"github.com/jackc/pgx/v5"
)

func CheckAuthByID(id int, role_level int) (bool, error) {
	var exists bool

	if role_level == 1 {
		query := `SELECT EXISTS(
		SELECT 1
		    FROM users u
		            INNER JOIN residences rc ON rc.id = u.residence_id
		            INNER JOIN societies sc ON sc.id = rc.society_id
		    WHERE u.role_level = 1
		    AND sc.access_revoked_at IS NULL
		    AND u.access_revoked_at IS NULL
		    AND u.id = :rId);`
		err := db.PGPool.QueryRow(context.Background(), query, id).Scan(&exists)
		if err != nil {
			return false, err
		}
		return exists, nil
	} else if role_level < 5 {
		query := `SELECT EXISTS(SELECT 1
		    FROM users u
		            INNER JOIN societies sc ON sc.id = u.society_id
		    WHERE u.role_level > 1
		    AND u.role_level < 5
		    AND sc.access_revoked_at IS NULL
		    AND u.access_revoked_at IS NULL
		    AND u.id = :rId);`
		err := db.PGPool.QueryRow(context.Background(), query, id).Scan(&exists)
		if err != nil {
			return false, err
		}
		return exists, nil
	}
	return false, nil

}

func BulkInsertBlocks(data models.BlocksCollector) error {
	context := context.Background()
	conn, err := db.PGPool.Acquire(context)
	if err != nil {
		return fmt.Errorf("!! failed to acquire db conn.: %w", err)
	}
	defer conn.Release()

	tx, err := conn.Begin(context)
	if err != nil {
		return fmt.Errorf("!! failed to start txn: %w", err)
	}

	batch := &pgx.Batch{}

	for _, name := range data.BlockNames {
		batch.Queue("INSERT INTO blocks (name, society_id) VALUES ($1, $2)",
			name, data.SocietyID,
		)
	}

	br := tx.SendBatch(context, batch)
	err = br.Close()
	if err != nil {
		tx.Rollback(context)
		return fmt.Errorf("!! failed to execute batch insert: %w", err)
	}

	err = tx.Commit(context)
	if err != nil {
		return fmt.Errorf("!! failed to commit transaction: %w", err)
	}

	return nil
}

func BulkInsertResidences(data models.ResidencesCollector) error {
	context := context.Background()
	conn, err := db.PGPool.Acquire(context)
	if err != nil {
		return fmt.Errorf("!! failed to acquire db conn.: %w", err)
	}

	defer conn.Release()

	tx, err := conn.Begin(context)
	if err != nil {
		return fmt.Errorf("!! failed to start txn: %w", err)
	}

	batch := &pgx.Batch{}

	for _, number := range data.ResidenceNumbers {
		batch.Queue("INSERT INTO residences (number, block_id, society_id) VALUES ($1, $2, $3)",
			number, data.BlockID, data.SocietyID,
		)
	}

	br := tx.SendBatch(context, batch)
	err = br.Close()
	if err != nil {
		tx.Rollback(context)
		return fmt.Errorf("!! failed to execute batch insert: %w", err)
	}

	err = tx.Commit(context)
	if err != nil {
		return fmt.Errorf("!! failed to commit transaction: %w", err)
	}

	return nil
}

func InsertVisitor(visitor models.VisitorCollector) (int, error) {
	var id int
	query := "INSERT INTO visitors (name, mobile, photo) VALUES ($1, $2, $3) RETURNING id"
	err := db.PGPool.QueryRow(context.Background(), query, visitor.Name, visitor.Mobile, visitor.Photo).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func GetVisitorIDByMobile(mobile string) (int, error) {
	var visitorID int
	query := "SELECT id FROM visitors v WHERE v.mobile = $1"
	err := db.PGPool.QueryRow(context.Background(), query, mobile).Scan(&visitorID)
	if err != nil {
		return 0, err
	}
	return visitorID, nil
}

func InsertResident(resident models.ResidentCollector) (string, error) {
	var id string
	query := `INSERT INTO residents (residence_id, is_primary) VALUES ($1, $2) RETURNING id`
	err := db.PGPool.QueryRow(context.Background(), query, resident.ResidenceID, resident.IsPrimary).Scan(&id)
	if err != nil {
		return "", err
	}
	return id, nil
}

func InsertResidenceVisit(visit models.ResidenceVisitCollector) (int, error) {
	var id int
	query := "INSERT INTO visits (residence_id, visitor_id, status) VALUES ($1, $2, $3) RETURNING id"
	err := db.PGPool.QueryRow(context.Background(), query, visit.ResidenceID, visit.VisitorID, visit.Status).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}
