package repository

import (
	"context"
	"fmt"
	"hime-backend/db"
	"hime-backend/models"

	"github.com/jackc/pgx/v5"
)

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

func InsertVisitor(visitor models.VisitorCollector, isPreapproved bool) (models.VisitorCollector, error) {
	var visitorResponse models.VisitorCollector
	query := "INSERT INTO visitors (name, phone_number, photo, purpose, is_preapproved) VALUES ($1, $2, $3, $4, $5) RETURNING id, name, phone_number, photo, purpose"
	row := db.PGPool.QueryRow(context.Background(), query, visitor.Name, visitor.PhoneNumber, visitor.Photo, visitor.Purpose, isPreapproved)
	err := row.Scan(
		&visitorResponse.ID,
		&visitorResponse.Name,
		&visitorResponse.PhoneNumber,
		&visitorResponse.Photo,
		&visitorResponse.Purpose,
	)
	return visitorResponse, err
}

func GetVisitorByMobile(phoneNumber string) (models.Visitor, error) {
	var visitor models.Visitor
	query := `SELECT id, name, phone_number, photo, purpose, is_preapproved, created_at, updated_at
	FROM visitors v WHERE v.phone_number = $1`
	err := db.PGPool.QueryRow(context.Background(), query, phoneNumber).Scan(
		&visitor.ID,
		&visitor.Name,
		&visitor.PhoneNumber,
		&visitor.Photo,
		&visitor.Purpose,
		&visitor.IsPreapproved,
		&visitor.CreatedAt,
		&visitor.UpdatedAt,
	)

	return visitor, err
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
