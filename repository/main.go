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
